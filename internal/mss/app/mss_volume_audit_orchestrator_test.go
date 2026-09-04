// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Validate volume reconciliation and unresolved-space accounting.
package app

import (
	"context"
	"errors"
	"mac-storage-scout/internal/mss/domain"
	"testing"
)

type auditTestScanner struct {
	roots    []*domain.MssNode
	counters domain.MssCounters
	err      error
}

func (s *auditTestScanner) Run(_ context.Context, _ domain.MssScanConfig) ([]*domain.MssNode, domain.MssCounters, error) {
	return s.roots, s.counters, s.err
}

type auditTestUsageProvider struct {
	results []domain.MssVolumeUsage
	err     error
	calls   int
}

func (p *auditTestUsageProvider) Measure(_ string) (domain.MssVolumeUsage, error) {
	if p.err != nil {
		return domain.MssVolumeUsage{}, p.err
	}
	result := p.results[p.calls]
	p.calls++
	return result, nil
}

func TestMssVolumeAuditOrchestratorRun(t *testing.T) {
	validCfg := domain.MssScanConfig{
		Paths:          []string{"/volume"},
		ThresholdBytes: 1,
		TopN:           1,
		SizeMode:       domain.MssSizeModeAllocated,
	}

	t.Run("reports accounted unaccounted and growth bytes", func(t *testing.T) {
		// START_AUDIT_RECONCILIATION_SETUP_FIXTURES
		scanner := &auditTestScanner{
			roots:    []*domain.MssNode{{Path: "/volume", SizeBytes: 600}},
			counters: domain.MssCounters{Errors: 3},
		}
		usage := &auditTestUsageProvider{results: []domain.MssVolumeUsage{
			{Path: "/volume", CapacityBytes: 1000, AvailableBytes: 200, OccupiedBytes: 800},
			{Path: "/volume", CapacityBytes: 1000, AvailableBytes: 150, OccupiedBytes: 850},
		}}
		auditor := &MssVolumeAuditOrchestrator{Scanner: scanner, Usage: usage}
		// END_AUDIT_RECONCILIATION_SETUP_FIXTURES

		audit, err := auditor.Run(context.Background(), "/volume", validCfg)
		if err != nil {
			t.Fatalf("Run(...) unexpected error: %v", err)
		}
		if audit.AccountedBytes != 600 {
			t.Errorf("Run(...) accounted = %d, want 600", audit.AccountedBytes)
		}
		if audit.ScanRoot != "/volume" {
			t.Errorf("Run(...) scan root = %q, want /volume", audit.ScanRoot)
		}
		if audit.UnaccountedBytes != 250 {
			t.Errorf("Run(...) unaccounted = %d, want 250", audit.UnaccountedBytes)
		}
		if audit.GrowthDuringScan != 50 {
			t.Errorf("Run(...) growth = %d, want 50", audit.GrowthDuringScan)
		}
		if audit.Counters.Errors != 3 {
			t.Errorf("Run(...) errors = %d, want 3", audit.Counters.Errors)
		}
	})

	t.Run("reports allocated size overcount separately", func(t *testing.T) {
		scanner := &auditTestScanner{roots: []*domain.MssNode{{Path: "/volume", SizeBytes: 900}}}
		usage := &auditTestUsageProvider{results: []domain.MssVolumeUsage{
			{Path: "/volume", CapacityBytes: 1000, AvailableBytes: 150, OccupiedBytes: 850},
			{Path: "/volume", CapacityBytes: 1000, AvailableBytes: 150, OccupiedBytes: 850},
		}}
		auditor := &MssVolumeAuditOrchestrator{Scanner: scanner, Usage: usage}

		audit, err := auditor.Run(context.Background(), "/volume", validCfg)
		if err != nil {
			t.Fatalf("Run(...) unexpected error: %v", err)
		}
		if audit.UnaccountedBytes != 0 {
			t.Errorf("Run(...) unaccounted = %d, want 0", audit.UnaccountedBytes)
		}
		if audit.OvercountBytes != 50 {
			t.Errorf("Run(...) overcount = %d, want 50", audit.OvercountBytes)
		}
	})

	t.Run("rejects invalid reconciliation preconditions", func(t *testing.T) {
		testCases := []struct {
			name    string
			auditor *MssVolumeAuditOrchestrator
			volume  string
			cfg     domain.MssScanConfig
		}{
			{name: "nil scanner", auditor: &MssVolumeAuditOrchestrator{Usage: &auditTestUsageProvider{}}, volume: "/volume", cfg: validCfg},
			{name: "nil usage provider", auditor: &MssVolumeAuditOrchestrator{Scanner: &auditTestScanner{}}, volume: "/volume", cfg: validCfg},
			{name: "empty volume", auditor: &MssVolumeAuditOrchestrator{Scanner: &auditTestScanner{}, Usage: &auditTestUsageProvider{}}, volume: "", cfg: validCfg},
			{name: "logical mode", auditor: &MssVolumeAuditOrchestrator{Scanner: &auditTestScanner{}, Usage: &auditTestUsageProvider{}}, volume: "/volume", cfg: domain.MssScanConfig{Paths: []string{"/volume"}, SizeMode: domain.MssSizeModeLogical}},
			{name: "multiple roots", auditor: &MssVolumeAuditOrchestrator{Scanner: &auditTestScanner{}, Usage: &auditTestUsageProvider{}}, volume: "/volume", cfg: domain.MssScanConfig{Paths: []string{"/a", "/b"}, SizeMode: domain.MssSizeModeAllocated}},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				if _, err := tc.auditor.Run(context.Background(), tc.volume, tc.cfg); err == nil {
					t.Errorf("Run(...) error = nil, want validation error")
				}
			})
		}
	})

	t.Run("preserves scanner failure cause", func(t *testing.T) {
		rootErr := errors.New("scan failed")
		auditor := &MssVolumeAuditOrchestrator{
			Scanner: &auditTestScanner{err: rootErr},
			Usage: &auditTestUsageProvider{results: []domain.MssVolumeUsage{
				{Path: "/volume", CapacityBytes: 1000, OccupiedBytes: 800},
			}},
		}

		_, err := auditor.Run(context.Background(), "/volume", validCfg)
		if !errors.Is(err, rootErr) {
			t.Errorf("Run(...) errors.Is(rootErr) = false, got %v", err)
		}
	})
}
