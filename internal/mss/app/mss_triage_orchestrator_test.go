// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Verify triage deltas, disappeared paths, activity, and conservative safety.
package app

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
	"testing"
	"time"
)

type triageTestCollector struct{ hotspots []domain.MssTriageHotspot }

func (c *triageTestCollector) Collect(context.Context, []string, time.Time) ([]domain.MssTriageHotspot, int64, error) {
	return append([]domain.MssTriageHotspot(nil), c.hotspots...), 0, nil
}

type triageTestState struct {
	loaded domain.MssTriageSnapshot
	saved  domain.MssTriageSnapshot
}

func (s *triageTestState) Load(string) (domain.MssTriageSnapshot, error) { return s.loaded, nil }
func (s *triageTestState) Save(_ string, snapshot domain.MssTriageSnapshot) error {
	s.saved = snapshot
	return nil
}

type triageTestProcesses struct{ activity map[string][]string }

func (p *triageTestProcesses) Probe(context.Context, []string) (map[string][]string, error) {
	return p.activity, nil
}

func TestMssTriageOrchestratorRun(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	state := &triageTestState{loaded: domain.MssTriageSnapshot{Version: 1, CapturedAt: now.Add(-time.Hour), OccupiedBytes: 800, Roots: []string{"/tmp"}, Paths: map[string]int64{"/tmp/cache": 50, "/tmp/active": 100, "/tmp/gone": 40}}}
	orchestrator := &MssTriageOrchestrator{
		Collector: &triageTestCollector{hotspots: []domain.MssTriageHotspot{
			{Path: "/tmp/cache", SizeBytes: 100, Age: domain.MssTriageAgeBytes{Older: 80}},
			{Path: "/tmp/active", SizeBytes: 100, Age: domain.MssTriageAgeBytes{Older: 80}},
		}},
		State:     state,
		Processes: &triageTestProcesses{activity: map[string][]string{"/tmp/active": {"node(7)"}}},
		Usage:     &auditTestUsageProvider{results: []domain.MssVolumeUsage{{Path: "/", OccupiedBytes: 1000}}},
	}
	report, err := orchestrator.Run(context.Background(), domain.MssTriageConfig{Paths: []string{"/tmp"}, StatePath: "/state", ThresholdBytes: 1, TopN: 5, SaveBaseline: true, Now: now})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if report.GrowthBytes != 200 {
		t.Errorf("Run() growth = %d, want 200", report.GrowthBytes)
	}
	cache := findAppTriageHotspot(t, report.Hotspots, "/tmp/cache")
	if cache.DeltaBytes != 50 || cache.Safety != "safe" {
		t.Errorf("Run() cache = %+v, want delta 50 and safe", cache)
	}
	active := findAppTriageHotspot(t, report.Hotspots, "/tmp/active")
	if active.Safety != "active" {
		t.Errorf("Run() active safety = %q, want active", active.Safety)
	}
	gone := findAppTriageHotspot(t, report.Hotspots, "/tmp/gone")
	if gone.DeltaBytes != -40 {
		t.Errorf("Run() gone delta = %d, want -40", gone.DeltaBytes)
	}
	if state.saved.Paths["/tmp/gone"] != 0 {
		t.Errorf("Run() saved gone = %d, want 0", state.saved.Paths["/tmp/gone"])
	}
}

func TestMssTriageOrchestratorIgnoresBaselineFromDifferentScope(t *testing.T) {
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	orchestrator := &MssTriageOrchestrator{
		Collector: &triageTestCollector{hotspots: []domain.MssTriageHotspot{{Path: "/new", SizeBytes: 100}}},
		State:     &triageTestState{loaded: domain.MssTriageSnapshot{CapturedAt: now.Add(-time.Hour), Roots: []string{"/old"}, Paths: map[string]int64{"/old": 500}}},
		Usage:     &auditTestUsageProvider{results: []domain.MssVolumeUsage{{Path: "/", OccupiedBytes: 1000}}},
	}
	report, err := orchestrator.Run(context.Background(), domain.MssTriageConfig{Paths: []string{"/new"}, StatePath: "/state", ThresholdBytes: 1, TopN: 5, Now: now})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !report.PreviousAt.IsZero() || report.Hotspots[0].DeltaBytes != 0 {
		t.Errorf("Run() used mismatched baseline: %+v", report)
	}
}

func TestMssClassifyTriageHotspotRequiresProcessEvidenceForSafe(t *testing.T) {
	safety, _ := mssClassifyTriageHotspot(domain.MssTriageHotspot{Path: "/tmp/cache", Age: domain.MssTriageAgeBytes{Older: 100}})
	if safety == "safe" {
		t.Errorf("mssClassifyTriageHotspot() = safe without process evidence")
	}
}

func findAppTriageHotspot(t *testing.T, hotspots []domain.MssTriageHotspot, path string) domain.MssTriageHotspot {
	t.Helper()
	for _, hotspot := range hotspots {
		if hotspot.Path == path {
			return hotspot
		}
	}
	t.Fatalf("hotspot %q not found", path)
	return domain.MssTriageHotspot{}
}
