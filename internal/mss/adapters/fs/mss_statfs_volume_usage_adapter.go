// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Measure mounted-volume capacity with the macOS statfs syscall.
package fs

import (
	"fmt"
	"mac-storage-scout/internal/mss/domain"
	"path/filepath"
	"syscall"
)

// MssStatfsVolumeUsageAdapter reads capacity and available blocks from statfs.
//
// @purpose Provide volume-level accounting without parsing localized shell output.
// @consumer cmd/mac-storage-scout/main.go
// @invariant OccupiedBytes equals CapacityBytes minus AvailableBytes.
// @implements {MssVolumeUsageProviderPort} internal/mss/ports/mss_volume_usage_provider_port.go
type MssStatfsVolumeUsageAdapter struct{}

// @see {MssVolumeUsageProviderPort#Measure} internal/mss/ports/mss_volume_usage_provider_port.go
// @purpose Measure capacity, available bytes, and non-available bytes for a mounted volume.
// @consumer internal/mss/app/mss_volume_audit_orchestrator.go
// @pre path exists on a mounted filesystem.
// @param path Existing path on the target volume.
// @returns Normalized usage snapshot or statfs error.
func (a *MssStatfsVolumeUsageAdapter) Measure(path string) (domain.MssVolumeUsage, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return domain.MssVolumeUsage{}, fmt.Errorf("[MssStatfsVolumeUsageAdapter.Measure] resolve path: %w", err)
	}

	var stat syscall.Statfs_t
	if err := syscall.Statfs(absPath, &stat); err != nil {
		return domain.MssVolumeUsage{}, fmt.Errorf("[MssStatfsVolumeUsageAdapter.Measure] statfs %s: %w", absPath, err)
	}

	// START_NORMALIZE_VOLUME_BLOCK_COUNTS
	// invariant: arithmetic saturates at MaxInt64 and never wraps signed report values.
	capacity := mssSaturatingBlockBytes(stat.Blocks, uint64(stat.Bsize))
	available := mssSaturatingBlockBytes(stat.Bavail, uint64(stat.Bsize))
	if available > capacity {
		available = capacity
	}
	// END_NORMALIZE_VOLUME_BLOCK_COUNTS

	return domain.MssVolumeUsage{
		Path:           mssMountedPath(stat.Mntonname, absPath),
		CapacityBytes:  capacity,
		AvailableBytes: available,
		OccupiedBytes:  capacity - available,
	}, nil
}

// mssMountedPath decodes the null-terminated mount path returned by statfs.
//
// @purpose Distinguish the filesystem mount from a narrower user-selected scan root.
// @consumer MssStatfsVolumeUsageAdapter.Measure.
// @param raw Null-terminated mount path bytes from statfs.
// @param fallback Path returned when statfs does not provide a mount name.
// @returns Decoded mount path or fallback.
func mssMountedPath(raw [1024]int8, fallback string) string {
	bytes := make([]byte, 0, len(raw))
	for _, value := range raw {
		if value == 0 {
			break
		}
		bytes = append(bytes, byte(value))
	}
	if len(bytes) == 0 {
		return fallback
	}
	return string(bytes)
}

// mssSaturatingBlockBytes multiplies block count and size without signed overflow.
//
// @purpose Keep filesystem counter conversion safe on malformed or very large values.
// @consumer MssStatfsVolumeUsageAdapter.Measure.
// @param blocks Filesystem block count.
// @param blockSize Bytes per filesystem block.
// @returns Byte count clamped to the maximum signed 64-bit value.
func mssSaturatingBlockBytes(blocks uint64, blockSize uint64) int64 {
	const maxInt64 = uint64(^uint64(0) >> 1)
	if blockSize != 0 && blocks > maxInt64/blockSize {
		return int64(maxInt64)
	}
	return int64(blocks * blockSize)
}
