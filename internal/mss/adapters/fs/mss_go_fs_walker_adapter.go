// @task spec/tasks/mac-storage-scout.task-02.md
// @purpose Walk filesystem quickly and emit metadata events.
package fs

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// MssGoFsWalkerAdapter traverses filesystem entries and emits metadata events.
//
// @purpose Provide runtime filesystem walking for scan orchestration.
// @consumer internal/mss/app/mss_scan_orchestrator.go
// @invariant Symlinks are not followed in default traversal mode.
// @implements {MssFilesystemWalkerPort} internal/mss/ports/mss_filesystem_walker_port.go
type MssGoFsWalkerAdapter struct{}

// mssQueueItem carries pending traversal item state.
//
// @purpose Transport path/root pair through worker queue.
// @consumer MssGoFsWalkerAdapter.Walk queue processing.
type mssQueueItem struct {
	path string
	root string
}

// @see {MssFilesystemWalkerPort#Walk} internal/mss/ports/mss_filesystem_walker_port.go
// @purpose Walk configured roots and emit normalized scan events.
// @consumer internal/mss/app/mss_scan_orchestrator.go
// @pre cfg.Paths contains at least one path.
// @param ctx Cancellation context.
// @param cfg Scan configuration.
// @param emit Event sink callback.
// @returns Traversal counters.
// @post Returns counters collected from walk lifecycle.
func (a *MssGoFsWalkerAdapter) Walk(ctx context.Context, cfg domain.MssScanConfig, emit func(domain.MssWalkEvent)) domain.MssCounters {
	counters := domain.MssCounters{StartedAt: time.Now()}
	if len(cfg.Paths) == 0 {
		return counters
	}

	if emit == nil {
		return counters
	}

	workers := cfg.Workers
	if workers <= 0 {
		workers = runtime.NumCPU() * 2
		if workers < 4 {
			workers = 4
		}
		if workers > 32 {
			workers = 32
		}
	}

	// START_INITIALIZE_WORKER_POOL
	// invariant: queue depth tracks pending jobs (+1 accepted, -1 before processing).
	// The scheduler owns the pending queue so workers never block while publishing
	// descendants into a full jobs channel.
	jobs := make(chan mssQueueItem)
	discovered := make(chan mssQueueItem, workers)
	done := make(chan struct{}, workers)
	var workersWG sync.WaitGroup
	var schedulerWG sync.WaitGroup
	var dirs, files, bytesSeen, errs, qDepth int64

	enqueue := func(item mssQueueItem) {
		atomic.AddInt64(&qDepth, 1)
		select {
		case discovered <- item:
		case <-ctx.Done():
			atomic.AddInt64(&qDepth, -1)
		}
	}

	workersWG.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer workersWG.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-jobs:
					if !ok {
						return
					}
					atomic.AddInt64(&qDepth, -1)
					a.walkPath(ctx, cfg, item.path, item.root, emit, &dirs, &files, &bytesSeen, &errs, enqueue)
					done <- struct{}{}
				}
			}
		}()
	}

	initial := make([]mssQueueItem, 0, len(cfg.Paths))
	for _, p := range cfg.Paths {
		ap := mssExpandPath(p)
		initial = append(initial, mssQueueItem{path: ap, root: ap})
		atomic.AddInt64(&qDepth, 1)
	}
	// END_INITIALIZE_WORKER_POOL

	// START_SCHEDULE_WORK_AFTER_DISCOVERY
	// purpose: serialize queue ownership and close workers after all active and discovered work completes.
	schedulerWG.Add(1)
	go func() {
		defer schedulerWG.Done()
		queue := initial
		pending := int64(len(queue))
		for {
			// A worker publishes descendants before its completion signal. Drain
			// already-published discoveries before deciding that work is finished.
			for {
				select {
				case item := <-discovered:
					queue = append(queue, item)
					pending++
				default:
					goto discoveriesDrained
				}
			}

		discoveriesDrained:
			if pending == 0 {
				close(jobs)
				return
			}

			var next mssQueueItem
			var dispatch chan<- mssQueueItem
			if len(queue) > 0 {
				next = queue[0]
				dispatch = jobs
			}

			select {
			case item := <-discovered:
				queue = append(queue, item)
				pending++
			case dispatch <- next:
				queue = queue[1:]
			case <-done:
				pending--
			case <-ctx.Done():
				close(jobs)
				return
			}
		}
	}()
	// END_SCHEDULE_WORK_AFTER_DISCOVERY

	schedulerWG.Wait()
	workersWG.Wait()

	counters.DirsScanned = atomic.LoadInt64(&dirs)
	counters.FilesScanned = atomic.LoadInt64(&files)
	counters.BytesSeen = atomic.LoadInt64(&bytesSeen)
	counters.Errors = atomic.LoadInt64(&errs)
	counters.QueueDepth = atomic.LoadInt64(&qDepth)
	return counters
}

// walkPath processes one queued path.
//
// @purpose Emit one path and enqueue descendants under walker policy.
// @consumer MssGoFsWalkerAdapter.Walk worker loop.
// @param ctx Cancellation context.
// @param cfg Scan configuration.
func (a *MssGoFsWalkerAdapter) walkPath(
	ctx context.Context,
	cfg domain.MssScanConfig,
	path string,
	root string,
	emit func(domain.MssWalkEvent),
	dirs *int64,
	files *int64,
	bytesSeen *int64,
	errs *int64,
	enqueue func(mssQueueItem),
) {
	// START_ABORT_ON_CONTEXT_CANCEL
	select {
	case <-ctx.Done():
		return
	default:
	}
	// END_ABORT_ON_CONTEXT_CANCEL

	fi, err := os.Lstat(path)
	if err != nil {
		emit(domain.MssWalkEvent{Err: err})
		atomic.AddInt64(errs, 1)
		return
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return
	}

	name := filepath.Base(path)
	if path == root {
		name = path
	}

	if fi.IsDir() {
		atomic.AddInt64(dirs, 1)
		emit(domain.MssWalkEvent{Entry: &domain.MssWalkEntry{
			Path:       path,
			ParentPath: filepath.Dir(path),
			Name:       name,
			Kind:       domain.MssEntryKindDir,
			SizeBytes:  0,
		}})

		entries, readErr := os.ReadDir(path)
		if readErr != nil {
			emit(domain.MssWalkEvent{Err: readErr})
			atomic.AddInt64(errs, 1)
			return
		}

		// START_EMIT_DIRECTORY_CHILDREN
		// failure mode: per-entry metadata errors are non-fatal and do not abort sibling processing.
		for _, de := range entries {
			child := filepath.Join(path, de.Name())
			if de.IsDir() {
				enqueue(mssQueueItem{path: child, root: root})
				continue
			}
			childInfo, infoErr := de.Info()
			if infoErr != nil {
				emit(domain.MssWalkEvent{Err: infoErr})
				atomic.AddInt64(errs, 1)
				continue
			}
			if childInfo.Mode()&os.ModeSymlink != 0 {
				continue
			}
			sz := mssSizeFromStat(cfg.SizeMode, childInfo.Sys(), childInfo.Size())
			atomic.AddInt64(files, 1)
			atomic.AddInt64(bytesSeen, sz)
			emit(domain.MssWalkEvent{Entry: &domain.MssWalkEntry{
				Path:       child,
				ParentPath: path,
				Name:       de.Name(),
				Kind:       domain.MssEntryKindFile,
				SizeBytes:  sz,
				Ext:        mssExtOf(de.Name()),
			}})
		}
		// END_EMIT_DIRECTORY_CHILDREN
		return
	}

	if fi.Mode().IsRegular() {
		sz := mssSizeFromStat(cfg.SizeMode, fi.Sys(), fi.Size())
		atomic.AddInt64(files, 1)
		atomic.AddInt64(bytesSeen, sz)
		emit(domain.MssWalkEvent{Entry: &domain.MssWalkEntry{
			Path:       path,
			ParentPath: filepath.Dir(path),
			Name:       name,
			Kind:       domain.MssEntryKindFile,
			SizeBytes:  sz,
			Ext:        mssExtOf(name),
		}})
	}
}

// mssExtOf extracts normalized file extension.
//
// @purpose Classify files for type aggregation.
// @consumer walkPath file event emission.
// @param name File name.
// @returns Extension or no-ext marker.
func mssExtOf(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return "no-ext"
	}
	return ext
}

// mssExpandPath resolves home and relative paths.
//
// @purpose Normalize input roots before traversal.
// @consumer MssGoFsWalkerAdapter.Walk root setup.
// @param p Input path.
// @returns Expanded absolute or original path.
func mssExpandPath(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		h, err := os.UserHomeDir()
		if err == nil {
			if p == "~" {
				return h
			}
			return filepath.Join(h, strings.TrimPrefix(p, "~/"))
		}
	}
	if filepath.IsAbs(p) {
		return p
	}
	ap, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return ap
}
