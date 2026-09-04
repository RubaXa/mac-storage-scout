// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Validate outcome-first volume audit rendering.
package report

import (
	"bytes"
	"io"
	"mac-storage-scout/internal/mss/domain"
	"strings"
	"testing"
)

func TestMssVolumeAuditTextReportAdapterRender(t *testing.T) {
	t.Run("renders reconciliation before the canonical tree", func(t *testing.T) {
		audit := domain.MssVolumeAudit{
			Usage: domain.MssVolumeUsage{
				Path:           "/volume",
				CapacityBytes:  1000,
				AvailableBytes: 150,
				OccupiedBytes:  850,
			},
			AccountedBytes:   600,
			UnaccountedBytes: 250,
			GrowthDuringScan: 50,
			Counters:         domain.MssCounters{Errors: 3},
			Roots: []*domain.MssNode{{
				Path:      "/volume",
				Name:      "/volume",
				Kind:      domain.MssEntryKindDir,
				SizeBytes: 600,
				Children:  []*domain.MssNode{{Name: "big.bin", Kind: domain.MssEntryKindFile, SizeBytes: 600}},
			}},
		}
		cfg := domain.MssScanConfig{ThresholdBytes: 500, TopN: 5, SizeMode: domain.MssSizeModeAllocated, PlainOutput: true}
		var out bytes.Buffer

		renderer := &MssVolumeAuditTextReportAdapter{}
		if err := renderer.Render(&out, audit, cfg); err != nil {
			t.Fatalf("Render(...) unexpected error: %v", err)
		}

		result := out.String()
		wantParts := []string{
			"mac-storage-scout audit",
			"occupied: 850B",
			"available: 150B",
			"accounted-by-readable-files: 600B (70.6%)",
			"unaccounted: 250B",
			"growth-during-scan: +50B",
			"scan-errors: 3",
			"access-warning:",
			"[/volume] 600B",
			"big.bin",
		}
		for _, part := range wantParts {
			if !strings.Contains(result, part) {
				t.Errorf("Render(...) output missing %q: %s", part, result)
			}
		}
		if strings.Index(result, "unaccounted: 250B") > strings.Index(result, "[/volume] 600B") {
			t.Errorf("Render(...) reconciliation appears after tree: %s", result)
		}
	})

	t.Run("rejects nil writer", func(t *testing.T) {
		renderer := &MssVolumeAuditTextReportAdapter{}
		if err := renderer.Render(io.Writer(nil), domain.MssVolumeAudit{}, domain.MssScanConfig{}); err == nil {
			t.Errorf("Render(nil, ...) error = nil, want non-nil")
		}
	})

	t.Run("warns when scan root is narrower than volume", func(t *testing.T) {
		audit := domain.MssVolumeAudit{
			ScanRoot: "/volume/subdir",
			Usage:    domain.MssVolumeUsage{Path: "/volume", CapacityBytes: 1000, OccupiedBytes: 800},
		}
		var out bytes.Buffer
		if err := (&MssVolumeAuditTextReportAdapter{}).Render(&out, audit, domain.MssScanConfig{TopN: 1}); err != nil {
			t.Fatalf("Render(...) unexpected error: %v", err)
		}
		if !strings.Contains(out.String(), "scope-warning:") {
			t.Errorf("Render(...) missing scope warning: %s", out.String())
		}
	})
}

func TestMssSignedHumanBytes(t *testing.T) {
	testCases := []struct {
		name  string
		value int64
		want  string
	}{
		{name: "growth", value: 1024, want: "+1.00KB"},
		{name: "reclaimed", value: -1024, want: "-1.00KB"},
		{name: "unchanged", value: 0, want: "+0B"},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if got := mssSignedHumanBytes(tc.value); got != tc.want {
				t.Errorf("mssSignedHumanBytes(%d) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}
