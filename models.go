package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// SystemStats represents current system resource usage
type SystemStats struct {
	CPU    CPUStats      `json:"cpu"`
	Memory MemoryStats   `json:"memory"`
	GPU    GPUStats      `json:"gpu"`
	Uptime time.Duration `json:"uptime"`
}

// CPUStats holds CPU usage information
type CPUStats struct {
	Usage   float64    `json:"usage"`    // Overall CPU usage percentage
	Cores   []float64  `json:"cores"`    // Per-core usage percentages
	LoadAvg [3]float64 `json:"load_avg"` // 1, 5, 15 minute load averages
	Temp    float64    `json:"temp"`     // CPU temperature in Celsius
}

// MemoryStats holds memory usage information
type MemoryStats struct {
	Total     uint64    `json:"total"`     // Total memory in bytes
	Used      uint64    `json:"used"`      // Used memory in bytes
	Available uint64    `json:"available"` // Available memory in bytes
	Usage     float64   `json:"usage"`     // Memory usage percentage
	Swap      SwapStats `json:"swap"`
}

// SwapStats holds swap usage information
type SwapStats struct {
	Total uint64  `json:"total"` // Total swap in bytes
	Used  uint64  `json:"used"`  // Used swap in bytes
	Usage float64 `json:"usage"` // Swap usage percentage
}

// GPUStats holds GPU usage information
type GPUStats struct {
	Usage       float64 `json:"usage"`        // GPU usage percentage
	MemoryUsage float64 `json:"memory_usage"` // GPU memory usage percentage
	MemoryUsed  uint64  `json:"memory_used"`  // GPU memory used in bytes
	MemoryTotal uint64  `json:"memory_total"` // Total GPU memory in bytes
	Temp        float64 `json:"temp"`         // GPU temperature in Celsius
}

// ViewMode represents different display modes
type ViewMode int

const (
	OverviewMode ViewMode = iota
	CPUDetailMode
	MemoryDetailMode
	GPUDetailMode
)

type model struct {
	stats        SystemStats
	viewMode     ViewMode
	refreshRate  time.Duration
	lastUpdate   time.Time
	width        int
	height       int
	quit         bool
	lastError    string
}

func initialModel() model {
	m := model{
		viewMode:    OverviewMode,
		refreshRate: time.Second,
		lastUpdate:  time.Now(),
		width:       80,
		height:      24,
		quit:        false,
		lastError:   "",
	}

	// Initialize with real system data
	if stats, err := collectSystemStats(); err == nil {
		m.stats = stats
	} else {
		m.lastError = fmt.Sprintf("Failed to initialize system stats: %v", err)
		// Provide default stats as fallback
		m.stats = SystemStats{
			CPU: CPUStats{
				Usage:   0.0,
				Cores:   make([]float64, 8),
				LoadAvg: [3]float64{0.0, 0.0, 0.0},
				Temp:    0.0,
			},
			Memory: MemoryStats{
				Total:     0,
				Used:      0,
				Available: 0,
				Usage:     0.0,
				Swap: SwapStats{
					Total: 0,
					Used:  0,
					Usage: 0.0,
				},
			},
			GPU: GPUStats{
				Usage:       0.0,
				MemoryUsage: 0.0,
				MemoryUsed:  0,
				MemoryTotal: 0,
				Temp:        0.0,
			},
			Uptime: 0,
		}
	}

	return m
}

// TickMsg represents a periodic update message
type TickMsg time.Time

func (m model) Init() tea.Cmd {
	return tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case TickMsg:
		// Update system stats with real data
		if newStats, err := collectSystemStats(); err == nil {
			m.stats = newStats
			m.lastError = "" // Clear any previous errors
		} else {
			m.lastError = fmt.Sprintf("Error collecting stats: %v", err)
		}
		m.lastUpdate = time.Time(msg)
		
		// Return next tick command
		return m, tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
			return TickMsg(t)
		})

	case tea.KeyMsg:
		switch msg.String() {

		// Exit the program
		case "ctrl+c", "q":
			m.quit = true
			return m, tea.Quit

		// Switch between view modes
		case "1":
			m.viewMode = OverviewMode
		case "2":
			m.viewMode = CPUDetailMode
		case "3":
			m.viewMode = MemoryDetailMode
		case "4":
			m.viewMode = GPUDetailMode

		// Refresh rate controls
		case "+", "=":
			if m.refreshRate > 100*time.Millisecond {
				m.refreshRate -= 100 * time.Millisecond
			}
		case "-", "_":
			if m.refreshRate < 5*time.Second {
				m.refreshRate += 100 * time.Millisecond
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quit {
		return ""
	}

	var s string

	// Header with title and current view mode
	switch m.viewMode {
	case OverviewMode:
		s += "mtop - System Monitor (Overview)\n"
	case CPUDetailMode:
		s += "mtop - CPU Details\n"
	case MemoryDetailMode:
		s += "mtop - Memory Details\n"
	case GPUDetailMode:
		s += "mtop - GPU Details\n"
	}

	s += fmt.Sprintf("Last update: %s | Refresh rate: %v\n", 
		m.lastUpdate.Format("15:04:05"), m.refreshRate)
	s += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"

	// Render content based on view mode
	switch m.viewMode {
	case OverviewMode:
		s += m.renderOverview()
	case CPUDetailMode:
		s += m.renderCPUDetail()
	case MemoryDetailMode:
		s += m.renderMemoryDetail()
	case GPUDetailMode:
		s += m.renderGPUDetail()
	}

	// Footer with controls and error display
	s += "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	if m.lastError != "" {
		s += fmt.Sprintf("⚠ %s\n", m.lastError)
	}
	s += "1: Overview | 2: CPU | 3: Memory | 4: GPU | +/-: Refresh rate | q: Quit\n"

	return s
}

func (m model) renderOverview() string {
	s := fmt.Sprintf("CPU Usage:    %.1f%% | Temp: %.1f°C\n", m.stats.CPU.Usage, m.stats.CPU.Temp)
	s += fmt.Sprintf("Memory Usage: %.1f%% (%.1f GB / %.1f GB)\n", 
		m.stats.Memory.Usage, 
		float64(m.stats.Memory.Used)/(1024*1024*1024),
		float64(m.stats.Memory.Total)/(1024*1024*1024))
	s += fmt.Sprintf("GPU Usage:    %.1f%% | Memory: %.1f%%\n", m.stats.GPU.Usage, m.stats.GPU.MemoryUsage)
	s += fmt.Sprintf("Load Average: %.2f, %.2f, %.2f\n", 
		m.stats.CPU.LoadAvg[0], m.stats.CPU.LoadAvg[1], m.stats.CPU.LoadAvg[2])
	s += fmt.Sprintf("Uptime:       %v\n", m.stats.Uptime.Round(time.Second))
	
	return s
}

func (m model) renderCPUDetail() string {
	s := fmt.Sprintf("Overall CPU Usage: %.1f%%\n", m.stats.CPU.Usage)
	s += fmt.Sprintf("Temperature: %.1f°C\n\n", m.stats.CPU.Temp)
	
	s += "Per-Core Usage:\n"
	for i, usage := range m.stats.CPU.Cores {
		s += fmt.Sprintf("Core %2d: %.1f%%\n", i, usage)
	}
	
	s += fmt.Sprintf("\nLoad Average: %.2f, %.2f, %.2f\n", 
		m.stats.CPU.LoadAvg[0], m.stats.CPU.LoadAvg[1], m.stats.CPU.LoadAvg[2])
	
	return s
}

func (m model) renderMemoryDetail() string {
	s := fmt.Sprintf("Memory Usage: %.1f%% (%.2f GB used / %.2f GB total)\n",
		m.stats.Memory.Usage,
		float64(m.stats.Memory.Used)/(1024*1024*1024),
		float64(m.stats.Memory.Total)/(1024*1024*1024))
	s += fmt.Sprintf("Available: %.2f GB\n\n", float64(m.stats.Memory.Available)/(1024*1024*1024))
	
	s += fmt.Sprintf("Swap Usage: %.1f%% (%.2f GB used / %.2f GB total)\n",
		m.stats.Memory.Swap.Usage,
		float64(m.stats.Memory.Swap.Used)/(1024*1024*1024),
		float64(m.stats.Memory.Swap.Total)/(1024*1024*1024))
	
	return s
}

func (m model) renderGPUDetail() string {
	s := fmt.Sprintf("GPU Usage: %.1f%%\n", m.stats.GPU.Usage)
	s += fmt.Sprintf("Temperature: %.1f°C\n\n", m.stats.GPU.Temp)
	
	s += fmt.Sprintf("GPU Memory Usage: %.1f%% (%.2f GB used / %.2f GB total)\n",
		m.stats.GPU.MemoryUsage,
		float64(m.stats.GPU.MemoryUsed)/(1024*1024*1024),
		float64(m.stats.GPU.MemoryTotal)/(1024*1024*1024))
	
	return s
}