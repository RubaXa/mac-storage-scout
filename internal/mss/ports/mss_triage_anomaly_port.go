// @task spec/tasks/mac-storage-scout.task-11.md
// @purpose Define the bounded metadata anomaly detection boundary.
package ports

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
	"time"
)

// MssTriageAnomalyDetectorPort finds suspicious directory structures before recursive sizing.
//
// @purpose Isolate cheap broad metadata discovery from expensive targeted byte accounting.
// @consumer internal/mss/app/mss_triage_orchestrator.go
type MssTriageAnomalyDetectorPort interface {
	// Detect performs a bounded metadata-only preflight.
	//
	// @purpose Find generalized fanout, queue, version, and stale-density signals.
	// @consumer internal/mss/app/mss_triage_orchestrator.go
	// @param ctx Cancellation context.
	// @param roots Broad roots eligible for metadata inspection.
	// @param now Stable age boundary.
	// @returns Partial-or-complete scan evidence and fatal caller cancellation error.
	Detect(ctx context.Context, roots []string, now time.Time) (domain.MssTriageAnomalyScan, error)
}
