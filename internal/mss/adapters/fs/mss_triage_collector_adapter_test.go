// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Verify shallow allocated-size and age aggregation.
package fs

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMssTriageCollectorAdapterCollect(t *testing.T) {
	root := t.TempDir()
	cache := filepath.Join(root, "cache", "nested")
	if err := os.MkdirAll(cache, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	oldPath := filepath.Join(cache, "old.bin")
	newPath := filepath.Join(cache, "new.bin")
	if err := os.WriteFile(oldPath, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(newPath, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(oldPath, now.Add(-8*24*time.Hour), now.Add(-8*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(newPath, now.Add(-time.Hour), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(oldPath, filepath.Join(root, "linked.bin")); err != nil {
		t.Fatal(err)
	}

	hotspots, errorsSeen, err := (&MssTriageCollectorAdapter{}).Collect(context.Background(), []string{root}, now)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if errorsSeen != 0 {
		t.Errorf("Collect() errors = %d, want 0", errorsSeen)
	}
	rootHotspot := findTriageHotspot(t, hotspots, root)
	if rootHotspot.Age.Older == 0 || rootHotspot.Age.Today == 0 {
		t.Errorf("Collect() root age = %+v, want old and today bytes", rootHotspot.Age)
	}
	if len(hotspots) != 3 {
		t.Errorf("Collect() hotspot count = %d, want 3 shallow paths", len(hotspots))
	}
}

func findTriageHotspot(t *testing.T, hotspots []domain.MssTriageHotspot, path string) domain.MssTriageHotspot {
	t.Helper()
	for _, hotspot := range hotspots {
		if hotspot.Path == path {
			return hotspot
		}
	}
	t.Fatalf("hotspot %q not found", path)
	return domain.MssTriageHotspot{}
}
