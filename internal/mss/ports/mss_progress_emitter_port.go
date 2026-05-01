// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define live progress contract.
package ports

import (
	"mac-storage-scout/internal/mss/domain"
	"sync/atomic"
)

// MssProgressEmitterPort starts and stops live progress rendering.
//
// @consumer internal/mss/app/mss_scan_orchestrator.go
type MssProgressEmitterPort interface {
	Start(counters *atomic.Pointer[domain.MssCounters]) func()
}
