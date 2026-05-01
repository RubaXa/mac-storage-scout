// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define tree aggregation contract.
package ports

import "mac-storage-scout/internal/mss/domain"

// MssNodeAggregatorPort builds threshold-aware trees from walk events.
//
// @purpose Define contract boundary for deterministic threshold aggregation.
// @consumer internal/mss/app/mss_scan_orchestrator.go
type MssNodeAggregatorPort interface {
	// BuildTree aggregates walk events into threshold-aware roots.
	//
	// @purpose Convert raw walk stream into report-ready tree roots.
	// @consumer internal/mss/app/mss_scan_orchestrator.go
	// @pre Events are produced by filesystem walker contract.
	// @param events Walk events from walker boundary.
	// @param cfg Scan configuration with threshold and top-N settings.
	// @returns Aggregated roots ready for rendering and optional contract error.
	// @post Every root preserves explicit-vs-other threshold split.
	BuildTree(events []domain.MssWalkEvent, cfg domain.MssScanConfig) ([]*domain.MssNode, error)
}
