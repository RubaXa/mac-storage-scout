// @task spec/tasks/mac-storage-scout.task-04.md
// @purpose Validate ASCII output shape for other bucket rendering.
package report

import (
	"bytes"
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
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "mss :: mac-storage-scout") {
		t.Fatalf("missing header: %s", out)
	}
	if !strings.Contains(out, "other (<500B each") {
		t.Fatalf("missing other threshold line: %s", out)
	}
	if !strings.Contains(out, "top-2") {
		t.Fatalf("missing top block: %s", out)
	}
	if !strings.Contains(out, "│  ├─") && !strings.Contains(out, "│  └─") {
		t.Fatalf("missing top item branches: %s", out)
	}
	if !strings.Contains(out, "types:") {
		t.Fatalf("missing types block: %s", out)
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
		t.Fatal(err)
	}
	out := buf.String()

	if strings.Contains(out, "🛰️") || strings.Contains(out, "📂") || strings.Contains(out, "📁") {
		t.Fatalf("plain output must not contain emoji: %s", out)
	}
	if !strings.Contains(out, "mss :: mac-storage-scout") {
		t.Fatalf("missing plain header: %s", out)
	}
	if !strings.Contains(out, "threshold: 500B | top: 5 | size-mode: logical") {
		t.Fatalf("missing plain config line: %s", out)
	}
}
