// @task spec/tasks/mac-storage-scout.task-03.md
// @purpose Validate threshold and other bucket aggregation behavior.
package aggregate

import (
	"mac-storage-scout/internal/mss/domain"
	"testing"
)

func TestBuildTreeThresholdOther(t *testing.T) {
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
		t.Fatal(err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	r := roots[0]
	if len(r.Children) != 1 {
		t.Fatalf("expected one visible child, got %d", len(r.Children))
	}
	if r.Other == nil || r.Other.SizeBytes != 300 || r.Other.Count != 2 {
		t.Fatalf("unexpected other bucket: %+v", r.Other)
	}
}
