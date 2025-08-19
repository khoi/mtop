package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// vm_statistics64 structure from mach/vm_statistics.h
type vm_statistics64 struct {
	FreeCount                          uint32
	ActiveCount                        uint32
	InactiveCount                      uint32
	WireCount                          uint32
	ZeroFillCount                      uint64
	Reactivations                      uint64
	Pageins                            uint64
	Pageouts                           uint64
	Faults                             uint64
	CowFaults                          uint64
	Lookups                            uint64
	Hits                               uint64
	Purges                             uint64
	PurgeableCount                     uint32
	SpeculativeCount                   uint32
	Decompressions                     uint64
	Compressions                       uint64
	Swapins                            uint64
	Swapouts                           uint64
	CompressorPageCount                uint32
	ThrottledCount                     uint32
	ExternalPageCount                  uint32
	InternalPageCount                  uint32
	TotalUncompressedPagesInCompressor uint64
}

// getVMStatistics64 calls host_statistics64 to get detailed VM statistics
func getVMStatistics64() (*vm_statistics64, error) {
	return GetVMStatisticsCGO()
}

// getPageSize gets the system page size using sysconf(_SC_PAGESIZE)
func getPageSize() (uint64, error) {
	pageSize, err := unix.SysctlUint64("hw.pagesize")
	if err != nil {
		// Fallback to checking vm.pagesize or using 4096
		if vmPageSize, vmErr := unix.SysctlUint64("vm.pagesize"); vmErr == nil {
			return vmPageSize, nil
		}
		return 4096, nil
	}
	return pageSize, nil
}

// collectSystemStats gathers all system statistics
func collectSystemStats() (SystemStats, error) {
	var stats SystemStats
	var err error

	// Collect memory stats
	stats.Memory, err = collectMemoryStats()
	if err != nil {
		return stats, fmt.Errorf("failed to collect memory stats: %w", err)
	}

	// Return empty CPU and GPU stats
	stats.CPU = CPUStats{}
	stats.GPU = GPUStats{}
	stats.Uptime = 0

	return stats, nil
}

// collectMemoryStats collects memory usage information using syscalls
func collectMemoryStats() (MemoryStats, error) {
	var memStats MemoryStats

	// Get total physical memory using sysctl
	physmem, err := unix.SysctlUint64("hw.memsize")
	if err != nil {
		return memStats, fmt.Errorf("failed to get physical memory: %w", err)
	}

	// Get page size
	pageSize, err := getPageSize()
	if err != nil {
		return memStats, fmt.Errorf("failed to get page size: %w", err)
	}

	// Get VM statistics using host_statistics64
	vmStats, err := getVMStatistics64()
	if err != nil {
		return memStats, fmt.Errorf("failed to get VM statistics: %w", err)
	}

	// Calculate total pages for validation
	totalPages := physmem / pageSize

	// Calculate memory usage
	// Used memory = active + inactive + wired + speculative + compressed - purgeable - external
	usedPages := uint64(0) +
		uint64(vmStats.ActiveCount) +
		uint64(vmStats.InactiveCount) +
		uint64(vmStats.WireCount) +
		uint64(vmStats.SpeculativeCount) +
		uint64(vmStats.CompressorPageCount) -
		uint64(vmStats.PurgeableCount) -
		uint64(vmStats.ExternalPageCount)

	freePages := uint64(vmStats.FreeCount)

	// Ensure used + free doesn't exceed total (fix approximation errors)
	if usedPages+freePages > totalPages {
		// Adjust used pages to be more realistic
		usedPages = totalPages - freePages
	}

	// Available memory = free + inactive + purgeable (more accurate)
	availablePages := freePages +
		uint64(vmStats.InactiveCount) +
		uint64(vmStats.PurgeableCount)

	// Ensure available doesn't exceed total
	if availablePages > totalPages {
		availablePages = totalPages
	}

	memStats.Total = physmem
	memStats.Used = usedPages * pageSize
	memStats.Available = availablePages * pageSize
	memStats.Usage = float64(memStats.Used) / float64(memStats.Total) * 100

	// Get swap information
	memStats.Swap, _ = collectSwapStats()

	return memStats, nil
}

func collectSwapStats() (SwapStats, error) {
	var swapStats SwapStats
	return swapStats, nil
}
