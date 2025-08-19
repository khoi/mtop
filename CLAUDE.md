# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

mtop is a system monitor for macOS written in Go that provides both:
- Interactive TUI mode using Bubble Tea framework
- JSON output mode for scripting/integration

## Build and Development Commands

```bash
# Build the binary
go build -o mtop

# Run the application (TUI mode)
go run .

# Run with JSON output
go run . --json

# Build for distribution
go build -ldflags="-s -w" -o mtop

# Test the build
./mtop --json
```

## Architecture

### Core Components

The application follows a Model-View-Update (MVU) pattern using Bubble Tea:

1. **main.go**: Entry point, handles CLI flags (--json) and initializes either TUI or JSON output mode
2. **models.go**: Defines data structures and Bubble Tea model, handles UI state and rendering
3. **system.go**: Platform-specific system stats collection using syscalls
4. **mach.go**: CGO bindings to macOS Mach kernel APIs for memory statistics

### Key Data Flow

```
User Input → main.go 
              ↓
         [JSON mode] → collectSystemStats() → JSON output
              ↓
         [TUI mode] → Bubble Tea Program
                        ↓
                     model.Update() → collectSystemStats()
                        ↓                    ↓
                     model.View()  ← system.go/mach.go
```

### macOS-Specific Implementation

The app uses CGO to call Mach kernel APIs (`host_statistics64`) for accurate memory statistics. This is necessary because Go's syscall package doesn't expose these low-level macOS APIs directly.

Memory calculation formula:
- Used = active + inactive + wired + speculative + compressed - purgeable - external
- Available = free + inactive + purgeable

### View Modes

The TUI supports 4 view modes (switchable with keys 1-4):
- Overview: Summary of all metrics
- CPU Detail: Per-core usage and load averages  
- Memory Detail: RAM and swap usage breakdown
- GPU Detail: GPU usage and memory

### Dependencies

- github.com/charmbracelet/bubbletea: TUI framework
- CGO: Required for Mach kernel API access
- golang.org/x/sys/unix: Unix syscalls