// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define filesystem walker contract.
package ports

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
)

// MssFilesystemWalkerPort walks filesystem entries and emits scan events.
type MssFilesystemWalkerPort interface {
	Walk(ctx context.Context, cfg domain.MssScanConfig, emit func(domain.MssWalkEvent)) domain.MssCounters
}
