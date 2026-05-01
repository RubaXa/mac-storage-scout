// @task spec/tasks/mac-storage-scout.task-02.md
// @purpose Validate runtime walker behavior, symlink policy, and non-fatal errors.
package fs

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
	"testing"
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
