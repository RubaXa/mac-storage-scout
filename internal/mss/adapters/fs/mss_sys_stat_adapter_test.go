// @task spec/tasks/mac-storage-scout.task-02.md
// @purpose Validate logical and allocated size selection from stat metadata.
package fs

import (
	"mac-storage-scout/internal/mss/domain"
	"syscall"
	"testing"
)

func TestMssSizeFromStat(t *testing.T) {
	t.Run("uses logical size in logical mode", func(t *testing.T) {
		st := &syscall.Stat_t{Size: 1234, Blocks: 99}
		got := mssSizeFromStat(domain.MssSizeModeLogical, st, 0)
		if got != 1234 {
			t.Errorf("mssSizeFromStat(logical, ...) = %d, want 1234", got)
		}
	})

	t.Run("uses allocated blocks in allocated mode", func(t *testing.T) {
		st := &syscall.Stat_t{Size: 1234, Blocks: 8}
		got := mssSizeFromStat(domain.MssSizeModeAllocated, st, 0)
		want := int64(8 * 512)
		if got != want {
			t.Errorf("mssSizeFromStat(allocated, ...) = %d, want %d", got, want)
		}
	})

	t.Run("uses fallback when syscall stat is unavailable", func(t *testing.T) {
		got := mssSizeFromStat(domain.MssSizeModeLogical, struct{}{}, 77)
		if got != 77 {
			t.Errorf("mssSizeFromStat(..., non-stat, 77) = %d, want 77", got)
		}
	})

	t.Run("clamps negative values to zero", func(t *testing.T) {
		logical := mssSizeFromStat(domain.MssSizeModeLogical, &syscall.Stat_t{Size: -1}, 10)
		if logical != 0 {
			t.Errorf("mssSizeFromStat(logical, negative size) = %d, want 0", logical)
		}

		allocated := mssSizeFromStat(domain.MssSizeModeAllocated, &syscall.Stat_t{Blocks: -1}, 10)
		if allocated != 0 {
			t.Errorf("mssSizeFromStat(allocated, negative blocks) = %d, want 0", allocated)
		}

		fallback := mssSizeFromStat(domain.MssSizeModeLogical, struct{}{}, -1)
		if fallback != 0 {
			t.Errorf("mssSizeFromStat(..., fallback=-1) = %d, want 0", fallback)
		}
	})
}
