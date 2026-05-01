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
	Run(ctx context.Context, cfg domain.MssScanConfig) ([]*domain.MssNode, domain.MssCounters, error)
}
