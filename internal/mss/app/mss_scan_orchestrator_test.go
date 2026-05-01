// @task spec/tasks/mac-storage-scout.task-05.md
// @purpose Validate orchestrator contract checks and lifecycle behavior.
package app

import (
	"context"
	"errors"
	"mac-storage-scout/internal/mss/domain"
	"sync/atomic"
	"testing"
)

type testWalker struct {
	events   []domain.MssWalkEvent
	counters domain.MssCounters
}

func (w *testWalker) Walk(_ context.Context, _ domain.MssScanConfig, emit func(domain.MssWalkEvent)) domain.MssCounters {
	for _, ev := range w.events {
		emit(ev)
	}
	return w.counters
}

type testAggregator struct {
	roots     []*domain.MssNode
	err       error
	seenEvent int
}

func (a *testAggregator) BuildTree(events []domain.MssWalkEvent, _ domain.MssScanConfig) ([]*domain.MssNode, error) {
	a.seenEvent = len(events)
	if a.err != nil {
		return nil, a.err
	}
	return a.roots, nil
}

type testProgress struct {
	startCalled int32
	stopCalled  int32
}

func (p *testProgress) Start(_ *atomic.Pointer[domain.MssCounters]) func() {
	atomic.StoreInt32(&p.startCalled, 1)
	return func() {
		atomic.StoreInt32(&p.stopCalled, 1)
	}
}

func TestMssScanOrchestratorRun(t *testing.T) {
	validCfg := domain.MssScanConfig{
		Paths:          []string{"/tmp"},
		ThresholdBytes: 1,
		TopN:           1,
	}

	t.Run("returns roots and walker counters on success", func(t *testing.T) {
		walker := &testWalker{
			events: []domain.MssWalkEvent{{
				Entry: &domain.MssWalkEntry{
					Path:       "/tmp",
					ParentPath: "/",
					Name:       "/tmp",
					Kind:       domain.MssEntryKindDir,
				},
			}},
			counters: domain.MssCounters{DirsScanned: 1},
		}
		agg := &testAggregator{roots: []*domain.MssNode{{Path: "/tmp", Kind: domain.MssEntryKindDir}}}
		progress := &testProgress{}
		orch := &MssScanOrchestrator{Walker: walker, Aggregator: agg, Progress: progress}

		roots, counters, err := orch.Run(context.Background(), validCfg)
		if err != nil {
			t.Fatalf("Run(...) unexpected error: %v", err)
		}
		if len(roots) != 1 {
			t.Errorf("Run(...) roots count = %d, want 1", len(roots))
		}
		if counters.DirsScanned != 1 {
			t.Errorf("Run(...) counters.DirsScanned = %d, want 1", counters.DirsScanned)
		}
		if agg.seenEvent != 1 {
			t.Errorf("Run(...) aggregator events = %d, want 1", agg.seenEvent)
		}
		if atomic.LoadInt32(&progress.startCalled) != 0 {
			t.Errorf("Run(...) progress should not start when cfg.Progress is false")
		}
	})

	t.Run("starts and stops progress when enabled", func(t *testing.T) {
		walker := &testWalker{}
		agg := &testAggregator{}
		progress := &testProgress{}
		orch := &MssScanOrchestrator{Walker: walker, Aggregator: agg, Progress: progress}

		cfg := validCfg
		cfg.Progress = true

		_, _, err := orch.Run(context.Background(), cfg)
		if err != nil {
			t.Fatalf("Run(...) unexpected error: %v", err)
		}
		if atomic.LoadInt32(&progress.startCalled) != 1 {
			t.Errorf("Run(...) progress start called = %d, want 1", atomic.LoadInt32(&progress.startCalled))
		}
		if atomic.LoadInt32(&progress.stopCalled) != 1 {
			t.Errorf("Run(...) progress stop called = %d, want 1", atomic.LoadInt32(&progress.stopCalled))
		}
	})

	t.Run("validates orchestrator contract preconditions", func(t *testing.T) {
		testCases := []struct {
			name string
			orch *MssScanOrchestrator
			cfg  domain.MssScanConfig
		}{
			{name: "nil walker", orch: &MssScanOrchestrator{Aggregator: &testAggregator{}}, cfg: validCfg},
			{name: "nil aggregator", orch: &MssScanOrchestrator{Walker: &testWalker{}}, cfg: validCfg},
			{name: "empty paths", orch: &MssScanOrchestrator{Walker: &testWalker{}, Aggregator: &testAggregator{}}, cfg: domain.MssScanConfig{ThresholdBytes: 1, TopN: 1}},
			{name: "non-positive threshold", orch: &MssScanOrchestrator{Walker: &testWalker{}, Aggregator: &testAggregator{}}, cfg: domain.MssScanConfig{Paths: []string{"/tmp"}, ThresholdBytes: 0, TopN: 1}},
			{name: "non-positive top", orch: &MssScanOrchestrator{Walker: &testWalker{}, Aggregator: &testAggregator{}}, cfg: domain.MssScanConfig{Paths: []string{"/tmp"}, ThresholdBytes: 1, TopN: 0}},
		}

		for _, tc := range testCases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				_, _, err := tc.orch.Run(context.Background(), tc.cfg)
				if err == nil {
					t.Errorf("Run(...) error = nil, want precondition error")
				}
			})
		}
	})

	t.Run("wraps aggregator errors with trace prefix", func(t *testing.T) {
		rootErr := errors.New("aggregate failed")
		orch := &MssScanOrchestrator{
			Walker:     &testWalker{},
			Aggregator: &testAggregator{err: rootErr},
		}
		_, _, err := orch.Run(context.Background(), validCfg)
		if err == nil {
			t.Fatalf("Run(...) error = nil, want wrapped aggregate error")
		}
		if !errors.Is(err, rootErr) {
			t.Errorf("Run(...) errors.Is(err, rootErr) = false, want true")
		}
	})
}
