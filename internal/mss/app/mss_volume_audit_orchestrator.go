// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Reconcile filesystem scan totals with mounted-volume occupancy.
package app

import (
	"context"
	"fmt"
	"mac-storage-scout/internal/mss/domain"
	"mac-storage-scout/internal/mss/ports"
)

// MssVolumeAuditOrchestrator coordinates volume snapshots and one allocated-size scan.
//
// @purpose Explain how much occupied space is attributable to readable files and how much remains unresolved.
// @consumer cmd/mac-storage-scout/main.go
// @invariant Accounted, unaccounted, and overcount values are non-negative.
type MssVolumeAuditOrchestrator struct {
	Scanner ports.MssScanOrchestratorPort
	Usage   ports.MssVolumeUsageProviderPort
}

// Run executes one volume reconciliation audit.
//
// @purpose Produce a truthful disk balance even when permissions or APFS semantics prevent full attribution.
// @consumer cmd/mac-storage-scout/main.go
// @pre volumePath is non-empty, Scanner and Usage are configured, and cfg uses allocated size.
// @param ctx Cancellation and deadline propagation context.
// @param volumePath Existing path on the audited mounted volume.
// @param cfg Scan configuration rooted at the audited volume.
// @returns Reconciled audit report or execution error.
// @post UnaccountedBytes explicitly contains occupied bytes not represented by readable scan roots.
func (o *MssVolumeAuditOrchestrator) Run(ctx context.Context, volumePath string, cfg domain.MssScanConfig) (domain.MssVolumeAudit, error) {
	if err := mssValidateVolumeAuditConfig(o, volumePath, cfg); err != nil {
		return domain.MssVolumeAudit{}, err
	}

	startUsage, err := o.Usage.Measure(volumePath)
	if err != nil {
		return domain.MssVolumeAudit{}, fmt.Errorf("[MssVolumeAuditOrchestrator.Run] measure volume before scan: %w", err)
	}

	roots, counters, err := o.Scanner.Run(ctx, cfg)
	if err != nil {
		return domain.MssVolumeAudit{}, fmt.Errorf("[MssVolumeAuditOrchestrator.Run] scan volume: %w", err)
	}

	endUsage, err := o.Usage.Measure(volumePath)
	if err != nil {
		return domain.MssVolumeAudit{}, fmt.Errorf("[MssVolumeAuditOrchestrator.Run] measure volume after scan: %w", err)
	}

	// START_RECONCILE_VOLUME_AND_READABLE_FILES
	// invariant: APFS clone overcount and unreadable/unattributed occupancy are reported separately.
	var accounted int64
	for _, root := range roots {
		if root != nil && root.SizeBytes > 0 {
			accounted += root.SizeBytes
		}
	}

	unaccounted := endUsage.OccupiedBytes - accounted
	overcount := int64(0)
	if unaccounted < 0 {
		overcount = -unaccounted
		unaccounted = 0
	}
	// END_RECONCILE_VOLUME_AND_READABLE_FILES

	return domain.MssVolumeAudit{
		ScanRoot:         cfg.Paths[0],
		Usage:            endUsage,
		AccountedBytes:   accounted,
		UnaccountedBytes: unaccounted,
		OvercountBytes:   overcount,
		GrowthDuringScan: endUsage.OccupiedBytes - startUsage.OccupiedBytes,
		Counters:         counters,
		Roots:            roots,
	}, nil
}

// mssValidateVolumeAuditConfig validates audit dependencies and scan semantics.
//
// @purpose Fail before expensive traversal when reconciliation preconditions are invalid.
// @consumer MssVolumeAuditOrchestrator.Run.
// @param o Audit orchestrator instance.
// @param volumePath Target mounted-volume path.
// @param cfg Scan configuration.
// @returns Validation error when a precondition is broken.
func mssValidateVolumeAuditConfig(o *MssVolumeAuditOrchestrator, volumePath string, cfg domain.MssScanConfig) error {
	if o == nil {
		return fmt.Errorf("[MssVolumeAuditOrchestrator.Run] orchestrator is nil")
	}
	if o.Scanner == nil {
		return fmt.Errorf("[MssVolumeAuditOrchestrator.Run] scanner is nil")
	}
	if o.Usage == nil {
		return fmt.Errorf("[MssVolumeAuditOrchestrator.Run] volume usage provider is nil")
	}
	if volumePath == "" {
		return fmt.Errorf("[MssVolumeAuditOrchestrator.Run] volume path is empty")
	}
	if cfg.SizeMode != domain.MssSizeModeAllocated {
		return fmt.Errorf("[MssVolumeAuditOrchestrator.Run] audit requires allocated size mode")
	}
	if len(cfg.Paths) != 1 {
		return fmt.Errorf("[MssVolumeAuditOrchestrator.Run] audit requires exactly one scan root")
	}
	return nil
}
