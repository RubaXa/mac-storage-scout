// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Define the mounted-volume capacity measurement boundary.
package ports

import "mac-storage-scout/internal/mss/domain"

// MssVolumeUsageProviderPort measures mounted-volume capacity and availability.
//
// @purpose Isolate OS-specific volume accounting from audit orchestration.
// @consumer internal/mss/app/mss_volume_audit_orchestrator.go
type MssVolumeUsageProviderPort interface {
	// Measure returns one volume-level usage snapshot for path.
	//
	// @purpose Provide the authoritative capacity baseline for reconciliation.
	// @consumer internal/mss/app/mss_volume_audit_orchestrator.go
	// @param path Existing path on the target mounted volume.
	// @returns Volume usage snapshot or measurement error.
	Measure(path string) (domain.MssVolumeUsage, error)
}
