// @task spec/tasks/mac-storage-scout.task-02.md
// @purpose Validate runtime walker behavior, symlink policy, and non-fatal errors.
package fs

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestMssGoFsWalkerAdapterWalk(t *testing.T) {
	t.Run("walks regular files and directories without following symlinks", func(t *testing.T) {
		root := t.TempDir()
		nested := filepath.Join(root, "nested")
		if err := os.Mkdir(nested, 0o755); err != nil {
			t.Fatalf("mkdir nested: %v", err)
		}

		rootFile := filepath.Join(root, "root.txt")
		if err := os.WriteFile(rootFile, []byte("root-data"), 0o644); err != nil {
			t.Fatalf("write root file: %v", err)
		}
		nestedFile := filepath.Join(nested, "nested.log")
		if err := os.WriteFile(nestedFile, []byte("nested-data"), 0o644); err != nil {
			t.Fatalf("write nested file: %v", err)
		}

		linkPath := filepath.Join(root, "nested-link")
		if err := os.Symlink(nested, linkPath); err != nil {
			t.Fatalf("create symlink: %v", err)
		}

		cfg := domain.MssScanConfig{
			Paths:          []string{root},
			ThresholdBytes: 1,
			TopN:           1,
			SizeMode:       domain.MssSizeModeLogical,
			Workers:        2,
		}

		adapter := &MssGoFsWalkerAdapter{}
		var events []domain.MssWalkEvent
		counters := adapter.Walk(context.Background(), cfg, func(ev domain.MssWalkEvent) {
			events = append(events, ev)
		})

		if counters.Errors != 0 {
			t.Errorf("Walk(...) errors = %d, want 0", counters.Errors)
		}
		if counters.DirsScanned < 2 {
			t.Errorf("Walk(...) dirs scanned = %d, want at least 2", counters.DirsScanned)
		}
		if counters.FilesScanned != 2 {
			t.Errorf("Walk(...) files scanned = %d, want 2", counters.FilesScanned)
		}
		if counters.BytesSeen <= 0 {
			t.Errorf("Walk(...) bytes seen = %d, want > 0", counters.BytesSeen)
		}

		for _, ev := range events {
			if ev.Entry == nil {
				continue
			}
			if ev.Entry.Path == linkPath {
				t.Errorf("Walk(...) emitted symlink entry %q, expected symlink skip", linkPath)
			}
		}
	})

	t.Run("reports missing root path as non-fatal error", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "missing")
		cfg := domain.MssScanConfig{
			Paths:          []string{root},
			ThresholdBytes: 1,
			TopN:           1,
			SizeMode:       domain.MssSizeModeLogical,
			Workers:        1,
		}

		adapter := &MssGoFsWalkerAdapter{}
		var errorEvents int
		counters := adapter.Walk(context.Background(), cfg, func(ev domain.MssWalkEvent) {
			if ev.Err != nil {
				errorEvents++
			}
		})

		if counters.Errors == 0 {
			t.Errorf("Walk(...) errors = %d, want > 0", counters.Errors)
		}
		if errorEvents == 0 {
			t.Errorf("Walk(...) error events = %d, want > 0", errorEvents)
		}
	})
}

func TestMssGoFsWalkerAdapterWalkDoesNotDeadlockWhenQueueBackpressureBuilds(t *testing.T) {
	root := t.TempDir()

	// START_SETUP_LARGE_DIRECTORY_QUEUE
	// purpose: fill the historical worker-to-worker jobs buffer with directory work.
	const directoryCount = 65537
	for i := 0; i < directoryCount; i++ {
		path := filepath.Join(root, "dir-"+strconv.Itoa(i))
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
	}
	// END_SETUP_LARGE_DIRECTORY_QUEUE

	cfg := domain.MssScanConfig{
		Paths:          []string{root},
		ThresholdBytes: 1,
		TopN:           1,
		SizeMode:       domain.MssSizeModeLogical,
		Workers:        1,
	}

	// START_TRIGGER_WALK_WITH_QUEUE_BACKPRESSURE
	// purpose: run the single-worker case that previously blocked on a full jobs channel.
	adapter := &MssGoFsWalkerAdapter{}
	finished := make(chan domain.MssCounters, 1)
	go func() {
		finished <- adapter.Walk(context.Background(), cfg, func(domain.MssWalkEvent) {})
	}()
	// END_TRIGGER_WALK_WITH_QUEUE_BACKPRESSURE

	// START_ASSERT_WALK_TERMINATES_AFTER_QUEUE_BACKPRESSURE
	select {
	case counters := <-finished:
		if counters.Errors != 0 {
			t.Errorf("Walk(...) errors = %d, want 0", counters.Errors)
		}
		if counters.DirsScanned != directoryCount+1 {
			t.Errorf("Walk(...) dirs scanned = %d, want %d", counters.DirsScanned, directoryCount+1)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("Walk(...) did not terminate after queue backpressure")
	}
}

func TestMssGoFsWalkerAdapterSkipsDifferentFilesystemDevice(t *testing.T) {
	root := t.TempDir()
	fi, err := os.Lstat(root)
	if err != nil {
		t.Fatalf("Lstat(%q) unexpected error: %v", root, err)
	}
	device, ok := mssDeviceFromFileInfo(fi)
	if !ok {
		t.Skip("syscall device metadata unavailable")
	}

	var emitted int64
	var dirs, files, bytesSeen, errs int64
	adapter := &MssGoFsWalkerAdapter{}
	adapter.walkPath(
		context.Background(),
		domain.MssScanConfig{OneFileSystem: true},
		mssQueueItem{path: root, root: root, rootDevice: device + 1, hasRootDevice: true},
		func(domain.MssWalkEvent) { emitted++ },
		&dirs,
		&files,
		&bytesSeen,
		&errs,
		func(mssQueueItem) {},
	)

	if emitted != 0 {
		t.Errorf("walkPath(different device) emitted = %d, want 0", emitted)
	}
	if dirs != 0 || files != 0 || errs != 0 {
		t.Errorf("walkPath(different device) counters = dirs:%d files:%d errors:%d, want all zero", dirs, files, errs)
	}
}
