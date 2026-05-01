// @task spec/tasks/mac-storage-scout.task-03.md
// @purpose Validate threshold and other bucket aggregation behavior.
package aggregate

import (
	"mac-storage-scout/internal/mss/domain"
	"testing"
)

func TestBuildTreeThresholdOther(t *testing.T) {
	t.Run("splits explicit and other items by threshold", func(t *testing.T) {
		a := &MssTreeAggregatorAdapter{}
		cfg := domain.MssScanConfig{Paths: []string{"/root"}, ThresholdBytes: 500, TopN: 2}
		events := []domain.MssWalkEvent{
			{Entry: &domain.MssWalkEntry{Path: "/root", ParentPath: "/", Name: "/root", Kind: domain.MssEntryKindDir}},
			{Entry: &domain.MssWalkEntry{Path: "/root/big.bin", ParentPath: "/root", Name: "big.bin", Kind: domain.MssEntryKindFile, SizeBytes: 700, Ext: ".bin"}},
			{Entry: &domain.MssWalkEntry{Path: "/root/small1.txt", ParentPath: "/root", Name: "small1.txt", Kind: domain.MssEntryKindFile, SizeBytes: 100, Ext: ".txt"}},
			{Entry: &domain.MssWalkEntry{Path: "/root/small2.log", ParentPath: "/root", Name: "small2.log", Kind: domain.MssEntryKindFile, SizeBytes: 200, Ext: ".log"}},
		}

		roots, err := a.BuildTree(events, cfg)
		if err != nil {
			t.Fatalf("BuildTree(...) unexpected error: %v", err)
		}
		if len(roots) != 1 {
			t.Fatalf("BuildTree(...) roots count = %d, want 1", len(roots))
		}
		r := roots[0]
		if len(r.Children) != 1 {
			t.Errorf("BuildTree(...) visible children count = %d, want 1", len(r.Children))
		}
		if r.Other == nil {
			t.Fatalf("BuildTree(...) other bucket = nil, want non-nil")
		}
		if r.Other.SizeBytes != 300 {
			t.Errorf("BuildTree(...) other size = %d, want 300", r.Other.SizeBytes)
		}
		if r.Other.Count != 2 {
			t.Errorf("BuildTree(...) other count = %d, want 2", r.Other.Count)
		}
	})
}

func TestBuildTreeRejectsInvalidTopN(t *testing.T) {
	a := &MssTreeAggregatorAdapter{}
	events := []domain.MssWalkEvent{
		{Entry: &domain.MssWalkEntry{Path: "/root", ParentPath: "/", Name: "/root", Kind: domain.MssEntryKindDir}},
	}

	t.Run("rejects non-positive topN", func(t *testing.T) {
		cfg := domain.MssScanConfig{Paths: []string{"/root"}, ThresholdBytes: 500, TopN: 0}
		if _, err := a.BuildTree(events, cfg); err == nil {
			t.Errorf("BuildTree(...) error = nil, want topN validation error")
		}
	})

	t.Run("rejects non-positive threshold", func(t *testing.T) {
		cfg := domain.MssScanConfig{Paths: []string{"/root"}, ThresholdBytes: 0, TopN: 2}
		if _, err := a.BuildTree(events, cfg); err == nil {
			t.Errorf("BuildTree(...) error = nil, want threshold validation error")
		}
	})
}
