// @task spec/tasks/mac-storage-scout.task-02.md
// @purpose Walk filesystem quickly and emit metadata events.
package fs

import (
	"context"
	"errors"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type MssGoFsWalkerAdapter struct{}

type mssQueueItem struct {
	path string
	root string
}

// @implements {MssFilesystemWalkerPort} internal/mss/ports/mss_filesystem_walker_port.go
func (a *MssGoFsWalkerAdapter) Walk(ctx context.Context, cfg domain.MssScanConfig, emit func(domain.MssWalkEvent)) domain.MssCounters {
	counters := domain.MssCounters{StartedAt: time.Now()}
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

	jobs := make(chan mssQueueItem, 65536)
	var taskWG sync.WaitGroup
	var workersWG sync.WaitGroup
	var dirs, files, bytesSeen, errs, qDepth int64

	enqueue := func(item mssQueueItem) {
		taskWG.Add(1)
		atomic.AddInt64(&qDepth, 1)
		select {
		case jobs <- item:
		case <-ctx.Done():
			taskWG.Done()
			atomic.AddInt64(&qDepth, -1)
		}
	}

	workersWG.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer workersWG.Done()
			for item := range jobs {
				atomic.AddInt64(&qDepth, -1)
				a.walkPath(ctx, cfg, item.path, item.root, emit, &dirs, &files, &bytesSeen, &errs, enqueue)
				taskWG.Done()
			}
		}()
	}

	for _, p := range cfg.Paths {
		ap := mssExpandPath(p)
		enqueue(mssQueueItem{path: ap, root: ap})
	}

	go func() {
		taskWG.Wait()
		close(jobs)
	}()

	workersWG.Wait()

	counters.DirsScanned = atomic.LoadInt64(&dirs)
	counters.FilesScanned = atomic.LoadInt64(&files)
	counters.BytesSeen = atomic.LoadInt64(&bytesSeen)
	counters.Errors = atomic.LoadInt64(&errs)
	counters.QueueDepth = atomic.LoadInt64(&qDepth)
	return counters
}

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
	select {
	case <-ctx.Done():
		return
	default:
	}

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

func mssExtOf(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return "no-ext"
	}
	return ext
}

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

func mssIsNonFatal(err error) bool {
	return errors.Is(err, os.ErrPermission) || errors.Is(err, os.ErrNotExist)
}
