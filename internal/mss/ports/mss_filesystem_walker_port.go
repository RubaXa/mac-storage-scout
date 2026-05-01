// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define filesystem walker contract.
package ports

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
)

// MssFilesystemWalkerPort walks filesystem entries and emits scan events.
//
// @purpose Define traversal contract between orchestrator and walker adapter.
// @consumer internal/mss/app/mss_scan_orchestrator.go
type MssFilesystemWalkerPort interface {
	// Walk traverses configured roots and emits filesystem events.
	//
	// @purpose Execute bounded traversal and emit normalized scan events.
	// @consumer internal/mss/app/mss_scan_orchestrator.go
	// @param ctx Cancellation and deadline propagation context.
	// @param cfg Scan configuration.
	// @param emit Event sink callback for walk entries and non-fatal errors.
	// @returns Final traversal counters.
	Walk(ctx context.Context, cfg domain.MssScanConfig, emit func(domain.MssWalkEvent)) domain.MssCounters
}
