// @task spec/tasks/mac-storage-scout.task-05.md
// @purpose Orchestrate scan walk, aggregation, and progress lifecycle.
package app

import (
	"context"
	"mac-storage-scout/internal/mss/domain"
	"mac-storage-scout/internal/mss/ports"
	"sync"
	"sync/atomic"
	"time"
)

type MssScanOrchestrator struct {
	Walker     ports.MssFilesystemWalkerPort
	Aggregator ports.MssNodeAggregatorPort
	Progress   ports.MssProgressEmitterPort
}

// @implements {MssScanOrchestratorPort} internal/mss/ports/mss_scan_orchestrator_port.go
func (o *MssScanOrchestrator) Run(ctx context.Context, cfg domain.MssScanConfig) ([]*domain.MssNode, domain.MssCounters, error) {
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

	walkerCounters := o.Walker.Walk(ctx, cfg, emit)
	roots, err := o.Aggregator.BuildTree(events, cfg)
	if err != nil {
		return nil, walkerCounters, err
	}
	return roots, walkerCounters, nil
}
