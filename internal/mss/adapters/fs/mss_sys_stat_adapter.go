// @task spec/tasks/mac-storage-scout.task-02.md
// @purpose Compute logical and allocated sizes from stat metadata.
package fs

import (
	"mac-storage-scout/internal/mss/domain"
	"syscall"
)

// mssSizeFromStat selects logical or allocated byte size from syscall metadata.
//
// @purpose Resolve file size according to selected size mode.
// @consumer mss_go_fs_walker_adapter walk path processing.
// @param mode Selected size mode.
// @param info Sys metadata payload.
// @returns Non-negative resolved byte size.
func mssSizeFromStat(mode domain.MssSizeMode, info any, fallback int64) int64 {
	st, ok := info.(*syscall.Stat_t)
	if !ok || st == nil {
		if fallback < 0 {
			return 0
		}
		return fallback
	}
	if mode == domain.MssSizeModeAllocated {
		v := st.Blocks * 512
		if v < 0 {
			return 0
		}
		return v
	}
	if st.Size < 0 {
		return 0
	}
	return st.Size
}
