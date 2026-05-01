// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define orchestrator contract.
package ports

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
)

// MssScanOrchestratorPort coordinates walk, aggregation, and final report inputs.
//
// @consumer cmd/mss/main.go
type MssScanOrchestratorPort interface {
	// Run executes scan lifecycle from traversal through aggregation.
	//
	// @purpose Provide one contract entrypoint for scan orchestration.
	// @consumer cmd/mss/main.go
	// @param ctx Cancellation and deadline propagation context.
	// @param cfg Scan configuration.
	// @returns Aggregated roots, counters snapshot, and optional execution error.
	Run(ctx context.Context, cfg domain.MssScanConfig) ([]*domain.MssNode, domain.MssCounters, error)
}
