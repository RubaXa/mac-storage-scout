// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Verify outcome-first triage report content and ordering.
package report

import (
	"bytes"
	"mac-storage-scout/internal/mss/domain"
	"strings"
	"testing"
	"time"
)

func TestMssTriageTextReportAdapterRender(t *testing.T) {
	report := domain.MssTriageReport{
		CapturedAt:   time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC),
		PreviousAt:   time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC),
		Usage:        domain.MssVolumeUsage{OccupiedBytes: 1000, AvailableBytes: 200},
		GrowthBytes:  300,
		BaselinePath: "/state.json",
		AnomalyScan: domain.MssTriageAnomalyScan{
			InspectedDirs: 42,
			Truncated:     true,
			Elapsed:       2 * time.Second,
			Findings: []domain.MssTriageAnomaly{
				{Path: "/tmp/reports", Kind: "extreme-fanout", Severity: "critical", Evidence: ">=4096 direct entries", Hint: "inspect producer", SizeBytes: 800, Age: domain.MssTriageAgeBytes{Older: 600}},
			},
		},
		Hotspots: []domain.MssTriageHotspot{
			{Path: "/tmp/cache", SizeBytes: 900, DeltaBytes: 400, Age: domain.MssTriageAgeBytes{Older: 700}, Safety: "safe", Reason: "stale"},
		},
	}
	var output bytes.Buffer
	err := (&MssTriageTextReportAdapter{}).Render(&output, report, domain.MssTriageConfig{ThresholdBytes: 1, TopN: 5})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, want := range []string{"mac-storage-scout triage", "volume-growth: +300B", "anomaly-preflight: dirs=42", "coverage=budget-limited", "[critical/extreme-fanout] /tmp/reports", "size=800B old>7d=600B", "next: inspect producer", "/tmp/cache", "old>7d=700B", "[safe]"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("Render() output missing %q:\n%s", want, output.String())
		}
	}
}

func TestMssTriageOwnerSummaryCapsLongLists(t *testing.T) {
	got := mssTriageOwnerSummary([]string{"a(1)", "b(2)", "c(3)", "d(4)"}, 2)
	if got != "a(1),b(2),+2 more" {
		t.Errorf("mssTriageOwnerSummary() = %q, want compact summary", got)
	}
}

func TestMssTriagePathOverlapsParentAndChild(t *testing.T) {
	if !mssTriagePathOverlaps("/tmp/cache/content", []string{"/tmp/cache"}) {
		t.Error("mssTriagePathOverlaps() = false for child candidate")
	}
	if mssTriagePathOverlaps("/tmp/other", []string{"/tmp/cache"}) {
		t.Error("mssTriagePathOverlaps() = true for disjoint candidate")
	}
}

func TestMssTriageReportOrdersMeasuredAnomaliesBySize(t *testing.T) {
	report := domain.MssTriageReport{
		CapturedAt:   time.Now(),
		BaselinePath: "/state",
		AnomalyScan: domain.MssTriageAnomalyScan{InspectedDirs: 2, Findings: []domain.MssTriageAnomaly{
			{Path: "/small-critical", Kind: "extreme-fanout", Severity: "critical", SizeBytes: 10, Hint: "small"},
			{Path: "/large-medium", Kind: "stale-dense", Severity: "medium", SizeBytes: 100, Hint: "large"},
		}},
	}
	var output bytes.Buffer
	if err := (&MssTriageTextReportAdapter{}).Render(&output, report, domain.MssTriageConfig{ThresholdBytes: 1, TopN: 5}); err != nil {
		t.Fatal(err)
	}
	largeAt := strings.Index(output.String(), "/large-medium")
	smallAt := strings.Index(output.String(), "/small-critical")
	if largeAt < 0 || smallAt < 0 || largeAt >= smallAt {
		t.Errorf("Render() anomaly order is not measured-size first:\n%s", output.String())
	}
}
