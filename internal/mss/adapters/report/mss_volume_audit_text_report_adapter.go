// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Render volume reconciliation before the detailed filesystem tree.
package report

import (
	"fmt"
	"io"
	"mac-storage-scout/internal/mss/domain"
)

// MssVolumeAuditTextReportAdapter renders an outcome-first volume audit report.
//
// @purpose Show occupied, available, accounted, and unresolved bytes in one deterministic balance.
// @consumer cmd/mac-storage-scout/main.go
// @invariant The detailed tree preserves the standard threshold and other-bucket output contract.
type MssVolumeAuditTextReportAdapter struct{}

// Render writes a volume balance followed by the standard threshold-aware tree.
//
// @purpose Make missing attribution and permission failures visible to the operator.
// @consumer cmd/mac-storage-scout/main.go
// @pre w is non-nil.
// @param w Output writer.
// @param audit Reconciled volume audit.
// @param cfg Scan configuration controlling tree detail and style.
// @returns Render error.
// @post Volume reconciliation appears before detailed root sections.
func (a *MssVolumeAuditTextReportAdapter) Render(w io.Writer, audit domain.MssVolumeAudit, cfg domain.MssScanConfig) error {
	if w == nil {
		return fmt.Errorf("[MssVolumeAuditTextReportAdapter.Render] writer is nil")
	}

	coverage := float64(0)
	if audit.Usage.OccupiedBytes > 0 {
		coverage = float64(audit.AccountedBytes) / float64(audit.Usage.OccupiedBytes) * 100
	}

	// START_RENDER_VOLUME_RECONCILIATION
	// purpose: expose the complete volume balance before path-level detail.
	fmt.Fprintln(w, "mac-storage-scout audit")
	fmt.Fprintf(w, "volume: %s\n", audit.Usage.Path)
	if audit.ScanRoot != "" && audit.ScanRoot != audit.Usage.Path {
		fmt.Fprintf(w, "scan-root: %s\n", audit.ScanRoot)
		fmt.Fprintln(w, "scope-warning: scan root is narrower than the containing volume; unaccounted includes everything outside that root")
	}
	fmt.Fprintf(w, "capacity: %s\n", domain.MssHumanBytes(audit.Usage.CapacityBytes))
	fmt.Fprintf(w, "occupied: %s\n", domain.MssHumanBytes(audit.Usage.OccupiedBytes))
	fmt.Fprintf(w, "available: %s\n", domain.MssHumanBytes(audit.Usage.AvailableBytes))
	fmt.Fprintf(w, "accounted-by-readable-files: %s (%.1f%%)\n", domain.MssHumanBytes(audit.AccountedBytes), coverage)
	fmt.Fprintf(w, "unaccounted: %s\n", domain.MssHumanBytes(audit.UnaccountedBytes))
	if audit.OvercountBytes > 0 {
		fmt.Fprintf(w, "allocated-size-overcount: %s\n", domain.MssHumanBytes(audit.OvercountBytes))
	}
	fmt.Fprintf(w, "growth-during-scan: %s\n", mssSignedHumanBytes(audit.GrowthDuringScan))
	fmt.Fprintf(w, "scan-errors: %d\n", audit.Counters.Errors)
	if audit.UnaccountedBytes > 0 {
		fmt.Fprintln(w, "unaccounted-may-include: unreadable paths, APFS shared volumes/clones, snapshots, reserved or purgeable space, and deleted-open files")
	}
	if audit.Counters.Errors > 0 {
		fmt.Fprintln(w, "access-warning: rerun with sufficient macOS Full Disk Access or privileges to reduce unaccounted bytes")
	}
	fmt.Fprintf(w, "threshold: %s | top: %d | size-mode: %s\n", domain.MssHumanBytes(cfg.ThresholdBytes), cfg.TopN, cfg.SizeMode)
	fmt.Fprintf(w, "sections: %d\n\n", len(audit.Roots))
	// END_RENDER_VOLUME_RECONCILIATION

	tree := &MssTreeTextReportAdapter{}
	tree.mssRenderRootSections(w, audit.Roots, cfg)
	return nil
}

// mssSignedHumanBytes formats a signed byte delta without losing its sign.
//
// @purpose Distinguish growth from reclaimed space during a long-running audit.
// @consumer MssVolumeAuditTextReportAdapter.Render.
// @param value Signed byte delta.
// @returns Human-readable value prefixed with plus or minus.
func mssSignedHumanBytes(value int64) string {
	if value < 0 {
		return "-" + domain.MssHumanBytes(-value)
	}
	return "+" + domain.MssHumanBytes(value)
}
