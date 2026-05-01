// @task spec/tasks/mac-storage-scout.task-05.md
// @purpose Validate end-to-end scan orchestration with real walker, aggregator, and report adapters.
package integration

import (
	"bytes"
	"context"
	"mac-storage-scout/internal/mss/adapters/aggregate"
	fsadapter "mac-storage-scout/internal/mss/adapters/fs"
	"mac-storage-scout/internal/mss/adapters/report"
	"mac-storage-scout/internal/mss/app"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanPipelineRendersThresholdAndOther(t *testing.T) {
	root := t.TempDir()
	big := filepath.Join(root, "big.bin")
	small := filepath.Join(root, "small.txt")

	if err := os.WriteFile(big, bytes.Repeat([]byte("A"), 256), 0o644); err != nil {
		t.Fatalf("write big fixture: %v", err)
	}
	if err := os.WriteFile(small, bytes.Repeat([]byte("B"), 32), 0o644); err != nil {
		t.Fatalf("write small fixture: %v", err)
	}

	cfg := domain.MssScanConfig{
		Paths:          []string{root},
		ThresholdBytes: 128,
		TopN:           5,
		SizeMode:       domain.MssSizeModeLogical,
		Workers:        2,
		Progress:       false,
		PlainOutput:    true,
	}

	orch := &app.MssScanOrchestrator{
		Walker:     &fsadapter.MssGoFsWalkerAdapter{},
		Aggregator: &aggregate.MssTreeAggregatorAdapter{},
	}

	roots, counters, err := orch.Run(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Run(...) unexpected error: %v", err)
	}
	if counters.FilesScanned < 2 {
		t.Errorf("Run(...) files scanned = %d, want at least 2", counters.FilesScanned)
	}
	if len(roots) != 1 {
		t.Fatalf("Run(...) roots count = %d, want 1", len(roots))
	}

	renderer := &report.MssTreeTextReportAdapter{}
	var out bytes.Buffer
	if err := renderer.Render(&out, roots, cfg); err != nil {
		t.Fatalf("Render(...) unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "big.bin") {
		t.Errorf("Render(...) output missing explicit large item: %s", result)
	}
	if !strings.Contains(result, "other (<128B each") {
		t.Errorf("Render(...) output missing other bucket: %s", result)
	}
	if !strings.Contains(result, "top-1") {
		t.Errorf("Render(...) output missing top-N block: %s", result)
	}
}
