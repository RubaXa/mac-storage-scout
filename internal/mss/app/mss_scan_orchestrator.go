// @task spec/tasks/mac-storage-scout.task-05.md
// @purpose Orchestrate scan walk, aggregation, and progress lifecycle.
package app

import (
	"context"
	"fmt"
	"mac-storage-scout/internal/mss/domain"
	"mac-storage-scout/internal/mss/ports"
	"sync"
	"sync/atomic"
	"time"
)

// MssScanOrchestrator wires walker, aggregator, and progress lifecycle.
//
// @purpose Coordinate full scan execution from walk events to aggregated roots.
// @consumer cmd/mss/main.go
// @invariant Non-fatal runtime errors do not crash orchestration when adapters follow contracts.
// @implements {MssScanOrchestratorPort} internal/mss/ports/mss_scan_orchestrator_port.go
type MssScanOrchestrator struct {
	Walker     ports.MssFilesystemWalkerPort
	Aggregator ports.MssNodeAggregatorPort
	Progress   ports.MssProgressEmitterPort
}

// @see {MssScanOrchestratorPort#Run} internal/mss/ports/mss_scan_orchestrator_port.go
// @purpose Run scan orchestration and return aggregated roots.
// @consumer cmd/mss/main.go
// @pre Walker and Aggregator are configured.
// @pre ThresholdBytes > 0, TopN >= 1, and at least one path is configured.
// @param ctx Execution context.
// @param cfg Scan configuration.
// @returns Aggregated roots, counters, and optional execution error.
// @post Returns roots aggregated from all emitted walk events.
// @post Progress goroutine is always stopped before return.
func (o *MssScanOrchestrator) Run(ctx context.Context, cfg domain.MssScanConfig) ([]*domain.MssNode, domain.MssCounters, error) {
	if err := mssValidateOrchestratorConfig(o, cfg); err != nil {
		return nil, domain.MssCounters{}, err
	}

	// START_COLLECT_WALK_EVENTS
	// invariant: every emitted event is appended exactly once and contributes to counters snapshot.
	events := make([]domain.MssWalkEvent, 0, 1024)
	var mu sync.Mutex

	var ptr atomic.Pointer[domain.MssCounters]
	base := &domain.MssCounters{StartedAt: time.Now()}
	ptr.Store(base)
	stopProgress := func() {}
	if cfg.Progress && o.Progress != nil {
		stopProgress = o.Progress.Start(&ptr)
	}
	defer stopProgress()

	emit := func(ev domain.MssWalkEvent) {
		mu.Lock()
		events = append(events, ev)
		mu.Unlock()

		cur := ptr.Load()
		if cur == nil {
			return
		}
		next := *cur
		if ev.Entry != nil {
			if ev.Entry.Kind == domain.MssEntryKindDir {
				next.DirsScanned++
			} else {
				next.FilesScanned++
				next.BytesSeen += ev.Entry.SizeBytes
			}
		}
		if ev.Err != nil {
			next.Errors++
		}
		ptr.Store(&next)
	}
	// END_COLLECT_WALK_EVENTS

	// START_EXECUTE_WALKER_AND_AGGREGATE
	// failure mode: aggregation failure must preserve walker counters for diagnostics.
	walkerCounters := o.Walker.Walk(ctx, cfg, emit)
	roots, err := o.Aggregator.BuildTree(events, cfg)
	if err != nil {
		return nil, walkerCounters, fmt.Errorf("[MssScanOrchestrator.Run] aggregate walk events: %w", err)
	}
	// END_EXECUTE_WALKER_AND_AGGREGATE
	return roots, walkerCounters, nil
}

// mssValidateOrchestratorConfig validates orchestrator dependencies and cfg.
//
// @purpose Fail fast on invalid orchestration preconditions.
// @consumer MssScanOrchestrator.Run.
// @param o Orchestrator instance.
// @param cfg Scan configuration.
// @returns Validation error when preconditions are broken.
func mssValidateOrchestratorConfig(o *MssScanOrchestrator, cfg domain.MssScanConfig) error {
	if o == nil {
		return fmt.Errorf("[MssScanOrchestrator.Run] orchestrator is nil")
	}
	if o.Walker == nil {
		return fmt.Errorf("[MssScanOrchestrator.Run] walker is nil")
	}
	if o.Aggregator == nil {
		return fmt.Errorf("[MssScanOrchestrator.Run] aggregator is nil")
	}
	if len(cfg.Paths) == 0 {
		return fmt.Errorf("[MssScanOrchestrator.Run] no scan paths configured")
	}
	if cfg.ThresholdBytes <= 0 {
		return fmt.Errorf("[MssScanOrchestrator.Run] threshold must be positive")
	}
	if cfg.TopN < 1 {
		return fmt.Errorf("[MssScanOrchestrator.Run] topN must be >= 1")
	}
	return nil
}
