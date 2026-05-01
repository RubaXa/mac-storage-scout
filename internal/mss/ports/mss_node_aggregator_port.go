// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define tree aggregation contract.
package ports

import "mac-storage-scout/internal/mss/domain"

// MssNodeAggregatorPort builds threshold-aware trees from walk events.
type MssNodeAggregatorPort interface {
	BuildTree(events []domain.MssWalkEvent, cfg domain.MssScanConfig) ([]*domain.MssNode, error)
}
