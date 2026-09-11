// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Measure high-churn filesystem roots with recency buckets.
package fs

import (
	"context"
	"errors"
	"io/fs"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

// MssTriageCollectorAdapter measures each root plus two descendant levels.
//
// @purpose Produce actionable hotspot totals without materializing a full report tree.
// @consumer internal/mss/app/mss_triage_orchestrator.go
// @implements {MssTriageCollectorPort} internal/mss/ports/mss_triage_ports.go
// @invariant Symlinks are not followed and allocated bytes are counted once per configured root.
type MssTriageCollectorAdapter struct{}

// Collect measures configured roots and shallow descendants.
//
// @purpose Gather allocated size and file-age evidence for incident triage.
// @consumer internal/mss/app/mss_triage_orchestrator.go
// @pre roots are absolute, existing, and non-overlapping.
// @param ctx Cancellation context.
// @param roots Existing non-overlapping scan roots.
// @param now Stable age boundary for all files.
// @returns Hotspots, non-fatal error count, and fatal cancellation error.
// @post Returned hotspots are sorted by size descending then path ascending.
func (a *MssTriageCollectorAdapter) Collect(ctx context.Context, roots []string, now time.Time) ([]domain.MssTriageHotspot, int64, error) {
	byPath := map[string]*domain.MssTriageHotspot{}
	var errorCount int64
	type rootResult struct {
		hotspots map[string]*domain.MssTriageHotspot
		errors   int64
		err      error
	}
	workerLimit := runtime.NumCPU()
	if workerLimit < 2 {
		workerLimit = 2
	}
	if workerLimit > 8 {
		workerLimit = 8
	}
	semaphore := make(chan struct{}, workerLimit)
	results := make(chan rootResult, len(roots))
	var wait sync.WaitGroup
	for _, configuredRoot := range roots {
		root := filepath.Clean(configuredRoot)
		wait.Add(1)
		go func() {
			defer wait.Done()
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				results <- rootResult{err: ctx.Err()}
				return
			}
			measured, errorsSeen, err := mssCollectTriageRoot(ctx, root, now)
			results <- rootResult{hotspots: measured, errors: errorsSeen, err: err}
		}()
	}
	go func() {
		wait.Wait()
		close(results)
	}()

	// START_MERGE_PARALLEL_ROOT_RESULTS
	// invariant: configured roots do not overlap, so each path total has one authoritative producer.
	for result := range results {
		errorCount += result.errors
		if result.err != nil {
			if errors.Is(result.err, context.Canceled) {
				return nil, errorCount, result.err
			}
			errorCount++
		}
		for path, hotspot := range result.hotspots {
			byPath[path] = hotspot
		}
	}
	// END_MERGE_PARALLEL_ROOT_RESULTS

	hotspots := make([]domain.MssTriageHotspot, 0, len(byPath))
	for _, hotspot := range byPath {
		if hotspot.SizeBytes > 0 {
			hotspots = append(hotspots, *hotspot)
		}
	}
	sort.SliceStable(hotspots, func(i, j int) bool {
		if hotspots[i].SizeBytes == hotspots[j].SizeBytes {
			return hotspots[i].Path < hotspots[j].Path
		}
		return hotspots[i].SizeBytes > hotspots[j].SizeBytes
	})
	return hotspots, errorCount, nil
}

// mssCollectTriageRoot measures one root without shared mutable state.
//
// @purpose Enable bounded root-level parallelism while preserving deterministic aggregation.
// @consumer MssTriageCollectorAdapter.Collect.
// @param ctx Cancellation context.
// @param root One configured scan root.
// @param now Stable age boundary.
// @returns Hotspots by path, non-fatal error count, and fatal cancellation error.
func mssCollectTriageRoot(ctx context.Context, root string, now time.Time) (map[string]*domain.MssTriageHotspot, int64, error) {
	byPath := map[string]*domain.MssTriageHotspot{root: {Path: root}}
	var errorCount int64
	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if walkErr != nil {
			errorCount++
			if entry != nil && entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			errorCount++
			return nil
		}
		size := mssTriageAllocatedSize(info)
		for _, aggregatePath := range mssTriageAggregatePaths(root, path) {
			hotspot := byPath[aggregatePath]
			if hotspot == nil {
				hotspot = &domain.MssTriageHotspot{Path: aggregatePath}
				byPath[aggregatePath] = hotspot
			}
			hotspot.SizeBytes += size
			mssAddTriageAge(&hotspot.Age, size, info.ModTime(), now)
		}
		return nil
	})
	return byPath, errorCount, walkErr
}

// mssTriageAllocatedSize reads allocated blocks with a logical-size fallback.
//
// @purpose Match triage totals to real disk pressure more closely than logical bytes.
// @consumer MssTriageCollectorAdapter.Collect.
// @param info File metadata.
// @returns Non-negative allocated byte estimate.
func mssTriageAllocatedSize(info os.FileInfo) int64 {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat != nil && stat.Blocks > 0 {
		return stat.Blocks * 512
	}
	if info.Size() > 0 {
		return info.Size()
	}
	return 0
}

// mssTriageAggregatePaths selects root and at most two descendant levels.
//
// @purpose Bound report cardinality while retaining useful drill-down.
// @consumer MssTriageCollectorAdapter.Collect.
// @param root Configured scan root.
// @param path File below root.
// @returns Ordered unique aggregation paths.
func mssTriageAggregatePaths(root, path string) []string {
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == "." || strings.HasPrefix(relative, "..") {
		return []string{root}
	}
	parts := strings.Split(relative, string(filepath.Separator))
	paths := []string{root}
	if len(parts) >= 1 {
		paths = append(paths, filepath.Join(root, parts[0]))
	}
	if len(parts) >= 2 {
		paths = append(paths, filepath.Join(root, parts[0], parts[1]))
	}
	return paths
}

// mssAddTriageAge assigns bytes to one mutually exclusive age bucket.
//
// @purpose Quantify recent churn and stale reclaim potential.
// @consumer MssTriageCollectorAdapter.Collect.
// @param age Mutable aggregate.
// @param size Allocated bytes.
// @param modified File modification time.
// @param now Stable scan time.
func mssAddTriageAge(age *domain.MssTriageAgeBytes, size int64, modified, now time.Time) {
	if modified.IsZero() {
		age.NoModTime += size
		return
	}
	ageDuration := now.Sub(modified)
	switch {
	case ageDuration < 24*time.Hour:
		age.Today += size
	case ageDuration < 7*24*time.Hour:
		age.Week += size
	default:
		age.Older += size
	}
}
