// @task spec/tasks/mac-storage-scout.task-04.md
// @purpose Validate ASCII output shape for other bucket rendering.
package report

import (
	"bytes"
	"io"
	"mac-storage-scout/internal/mss/domain"
	"strings"
	"testing"
)

func TestRenderContainsOtherBlock(t *testing.T) {
	r := &MssTreeTextReportAdapter{}
	root := &domain.MssNode{
		Path:      "/root",
		Name:      "/root",
		Kind:      domain.MssEntryKindDir,
		SizeBytes: 1000,
		Children: []*domain.MssNode{{
			Name:      "big.bin",
			Kind:      domain.MssEntryKindFile,
			SizeBytes: 700,
		}},
		Other: &domain.MssOtherBucket{
			Count:     2,
			SizeBytes: 300,
			TopItems:  []*domain.MssNode{{Name: "small1.txt", SizeBytes: 200}, {Name: "small2.log", SizeBytes: 100}},
			TypeTop:   []domain.MssExtStat{{Ext: ".txt", SizeBytes: 200, Count: 1}},
		},
	}
	cfg := domain.MssScanConfig{ThresholdBytes: 500}
	var buf bytes.Buffer
	if err := r.Render(&buf, []*domain.MssNode{root}, cfg); err != nil {
		t.Fatalf("Render(...) unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "mss :: mac-storage-scout") {
		t.Errorf("Render(...) output missing header: %s", out)
	}
	if !strings.Contains(out, "other (<500B each") {
		t.Errorf("Render(...) output missing other line: %s", out)
	}
	if !strings.Contains(out, "top-2") {
		t.Errorf("Render(...) output missing top block: %s", out)
	}
	if !strings.Contains(out, "│  ├─") && !strings.Contains(out, "│  └─") {
		t.Errorf("Render(...) output missing top branches: %s", out)
	}
	if !strings.Contains(out, "types:") {
		t.Errorf("Render(...) output missing types block: %s", out)
	}
}

func TestRenderPlainOutputUsesASCIIOnlyLabels(t *testing.T) {
	r := &MssTreeTextReportAdapter{}
	root := &domain.MssNode{
		Path:      "/root",
		Name:      "/root",
		Kind:      domain.MssEntryKindDir,
		SizeBytes: 700,
		Children: []*domain.MssNode{{
			Name:      "big.bin",
			Kind:      domain.MssEntryKindFile,
			SizeBytes: 700,
		}},
	}

	cfg := domain.MssScanConfig{ThresholdBytes: 500, TopN: 5, SizeMode: domain.MssSizeModeLogical, PlainOutput: true}
	var buf bytes.Buffer
	if err := r.Render(&buf, []*domain.MssNode{root}, cfg); err != nil {
		t.Fatalf("Render(...) unexpected error: %v", err)
	}
	out := buf.String()

	if strings.Contains(out, "🛰️") || strings.Contains(out, "📂") || strings.Contains(out, "📁") {
		t.Errorf("Render(...) plain output contains emoji: %s", out)
	}
	if !strings.Contains(out, "mss :: mac-storage-scout") {
		t.Errorf("Render(...) plain output missing header: %s", out)
	}
	if !strings.Contains(out, "threshold: 500B | top: 5 | size-mode: logical") {
		t.Errorf("Render(...) plain output missing config line: %s", out)
	}
}

func TestRenderRejectsNilWriter(t *testing.T) {
	r := &MssTreeTextReportAdapter{}
	cfg := domain.MssScanConfig{ThresholdBytes: 500, TopN: 5}
	if err := r.Render(io.Writer(nil), nil, cfg); err == nil {
		t.Errorf("Render(nil, ...) error = nil, want non-nil")
	}
}
