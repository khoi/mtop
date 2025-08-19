package main

/*
#include <mach/mach.h>
#include <mach/mach_host.h>
#include <mach/host_info.h>
#include <mach/vm_statistics.h>

int getVMStats(struct vm_statistics64 *stats) {
    mach_port_t host_port = mach_host_self();
    mach_msg_type_number_t count = HOST_VM_INFO64_COUNT;
    
    kern_return_t kr = host_statistics64(
        host_port,
        HOST_VM_INFO64,
        (host_info64_t)stats,
        &count
    );
    
    return kr;
}
*/
import "C"
import (
	"fmt"
)

// GetVMStatisticsCGO gets VM statistics using CGO
func GetVMStatisticsCGO() (*vm_statistics64, error) {
	var cStats C.struct_vm_statistics64
	
	ret := C.getVMStats(&cStats)
	if ret != 0 {
		return nil, fmt.Errorf("host_statistics64 failed with error code: %d", ret)
	}
	
	// Convert C struct to Go struct
	stats := &vm_statistics64{
		FreeCount:                          uint32(cStats.free_count),
		ActiveCount:                        uint32(cStats.active_count),
		InactiveCount:                      uint32(cStats.inactive_count),
		WireCount:                          uint32(cStats.wire_count),
		ZeroFillCount:                      uint64(cStats.zero_fill_count),
		Reactivations:                      uint64(cStats.reactivations),
		Pageins:                            uint64(cStats.pageins),
		Pageouts:                           uint64(cStats.pageouts),
		Faults:                             uint64(cStats.faults),
		CowFaults:                          uint64(cStats.cow_faults),
		Lookups:                            uint64(cStats.lookups),
		Hits:                               uint64(cStats.hits),
		Purges:                             uint64(cStats.purges),
		PurgeableCount:                     uint32(cStats.purgeable_count),
		SpeculativeCount:                   uint32(cStats.speculative_count),
		Decompressions:                     uint64(cStats.decompressions),
		Compressions:                       uint64(cStats.compressions),
		Swapins:                            uint64(cStats.swapins),
		Swapouts:                           uint64(cStats.swapouts),
		CompressorPageCount:                uint32(cStats.compressor_page_count),
		ThrottledCount:                     uint32(cStats.throttled_count),
		ExternalPageCount:                  uint32(cStats.external_page_count),
		InternalPageCount:                  uint32(cStats.internal_page_count),
		TotalUncompressedPagesInCompressor: uint64(cStats.total_uncompressed_pages_in_compressor),
	}
	
	return stats, nil
}