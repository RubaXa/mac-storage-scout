// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define live progress contract.
package ports

import (
	"mac-storage-scout/internal/mss/domain"
	"sync/atomic"
)

// MssProgressEmitterPort starts and stops live progress rendering.
//
// @purpose Define progress lifecycle boundary for scan orchestration.
// @consumer internal/mss/app/mss_scan_orchestrator.go
type MssProgressEmitterPort interface {
	// Start begins progress rendering and returns a stop callback.
	//
	// @purpose Provide non-blocking lifecycle hook for scan progress output.
	// @consumer internal/mss/app/mss_scan_orchestrator.go
	// @param counters Atomic counters pointer updated during scan.
	// @returns Stop callback for deterministic shutdown.
	Start(counters *atomic.Pointer[domain.MssCounters]) func()
}
