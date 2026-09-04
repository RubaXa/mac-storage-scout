// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Validate mounted-volume usage measurement and overflow safety.
package fs

import "testing"

func TestMssStatfsVolumeUsageAdapterMeasure(t *testing.T) {
	t.Run("returns a consistent snapshot for an existing path", func(t *testing.T) {
		adapter := &MssStatfsVolumeUsageAdapter{}
		usage, err := adapter.Measure(t.TempDir())
		if err != nil {
			t.Fatalf("Measure(...) unexpected error: %v", err)
		}
		if usage.CapacityBytes <= 0 {
			t.Errorf("Measure(...) capacity = %d, want > 0", usage.CapacityBytes)
		}
		if usage.AvailableBytes < 0 || usage.AvailableBytes > usage.CapacityBytes {
			t.Errorf("Measure(...) available = %d, want within [0,%d]", usage.AvailableBytes, usage.CapacityBytes)
		}
		if usage.OccupiedBytes != usage.CapacityBytes-usage.AvailableBytes {
			t.Errorf("Measure(...) occupied = %d, want %d", usage.OccupiedBytes, usage.CapacityBytes-usage.AvailableBytes)
		}
	})

	t.Run("returns an error for a missing path", func(t *testing.T) {
		adapter := &MssStatfsVolumeUsageAdapter{}
		if _, err := adapter.Measure(t.TempDir() + "/missing"); err == nil {
			t.Errorf("Measure(missing) error = nil, want non-nil")
		}
	})
}

func TestMssSaturatingBlockBytes(t *testing.T) {
	t.Run("multiplies normal block counts", func(t *testing.T) {
		if got := mssSaturatingBlockBytes(10, 4096); got != 40960 {
			t.Errorf("mssSaturatingBlockBytes(10, 4096) = %d, want 40960", got)
		}
	})

	t.Run("clamps overflowing values", func(t *testing.T) {
		const maxInt64 = int64(^uint64(0) >> 1)
		if got := mssSaturatingBlockBytes(^uint64(0), 4096); got != maxInt64 {
			t.Errorf("mssSaturatingBlockBytes(max, 4096) = %d, want %d", got, maxInt64)
		}
	})
}

func TestMssMountedPath(t *testing.T) {
	var raw [1024]int8
	for index, value := range []byte("/System/Volumes/Data") {
		raw[index] = int8(value)
	}
	if got := mssMountedPath(raw, "/fallback"); got != "/System/Volumes/Data" {
		t.Errorf("mssMountedPath(...) = %q, want /System/Volumes/Data", got)
	}
}
