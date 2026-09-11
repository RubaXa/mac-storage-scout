// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Verify atomic JSON triage baseline persistence.
package state

import (
	"mac-storage-scout/internal/mss/domain"
	"path/filepath"
	"testing"
	"time"
)

func TestMssJSONTriageStateAdapterRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "triage.json")
	want := domain.MssTriageSnapshot{Version: 1, CapturedAt: time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC), OccupiedBytes: 99, Roots: []string{"/tmp"}, Paths: map[string]int64{"/tmp": 42}}
	adapter := &MssJSONTriageStateAdapter{}
	if err := adapter.Save(path, want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := adapter.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.Version != want.Version || !got.CapturedAt.Equal(want.CapturedAt) || got.OccupiedBytes != want.OccupiedBytes || len(got.Roots) != 1 || got.Roots[0] != "/tmp" || got.Paths["/tmp"] != 42 {
		t.Errorf("Load() = %+v, want %+v", got, want)
	}
}
