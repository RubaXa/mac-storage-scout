// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Define incident-triage runtime boundaries.
package ports

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
	"time"
)

// MssTriageCollectorPort measures shallow hotspot totals and age buckets.
//
// @purpose Isolate filesystem traversal from triage policy and reporting.
// @consumer internal/mss/app/mss_triage_orchestrator.go
type MssTriageCollectorPort interface {
	// Collect measures roots and two shallow descendant levels.
	//
	// @purpose Produce comparable size and age evidence.
	// @consumer internal/mss/app/mss_triage_orchestrator.go
	// @param ctx Cancellation context.
	// @param roots Existing non-overlapping roots.
	// @param now Stable age boundary.
	// @returns Hotspots, non-fatal error count, and fatal error.
	Collect(ctx context.Context, roots []string, now time.Time) ([]domain.MssTriageHotspot, int64, error)
}

// MssTriageStatePort loads and saves compact triage baselines.
//
// @purpose Isolate persistent baseline storage from triage comparison logic.
// @consumer internal/mss/app/mss_triage_orchestrator.go
type MssTriageStatePort interface {
	// Load restores one prior snapshot.
	//
	// @purpose Provide path totals for delta comparison.
	// @consumer internal/mss/app/mss_triage_orchestrator.go
	// @param path Baseline file path.
	// @returns Decoded snapshot or read/decode error.
	Load(path string) (domain.MssTriageSnapshot, error)
	// Save atomically persists one snapshot.
	//
	// @purpose Make a completed run the next comparison baseline.
	// @consumer internal/mss/app/mss_triage_orchestrator.go
	// @param path Baseline file path.
	// @param snapshot Completed triage snapshot.
	// @returns Persistence error, if any.
	Save(path string, snapshot domain.MssTriageSnapshot) error
}

// MssProcessUsageProbePort reports processes with open files below hotspot paths.
//
// @purpose Keep optional macOS process attribution outside domain logic.
// @consumer internal/mss/app/mss_triage_orchestrator.go
type MssProcessUsageProbePort interface {
	// Probe attributes open files to candidate paths.
	//
	// @purpose Prevent active data from being presented as a safe candidate.
	// @consumer internal/mss/app/mss_triage_orchestrator.go
	// @param ctx Cancellation context.
	// @param paths Candidate hotspot paths.
	// @returns Process labels keyed by path and optional probe error.
	Probe(ctx context.Context, paths []string) (map[string][]string, error)
}
