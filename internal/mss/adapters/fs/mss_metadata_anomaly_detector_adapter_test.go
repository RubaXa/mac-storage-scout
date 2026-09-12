// @task spec/tasks/mac-storage-scout.task-11.md
// @purpose Verify bounded product-independent metadata anomaly detection.
package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMssMetadataAnomalyDetectorCapsExtremeFanout(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 40; i++ {
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("item-%03d", i)), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	detector := &MssMetadataAnomalyDetectorAdapter{MaxDirs: 1, MaxDepth: 1, MaxEntriesPerDir: 32}

	scan, err := detector.Detect(context.Background(), []string{root}, time.Now())
	if err != nil {
		t.Fatalf("Detect() error = %v, want nil", err)
	}
	if len(scan.Findings) != 1 {
		t.Fatalf("Detect() findings = %d, want 1", len(scan.Findings))
	}
	finding := scan.Findings[0]
	if finding.Kind != "extreme-fanout" || finding.Severity != "critical" || !finding.EntryCountMin || finding.EntryCount != 32 {
		t.Errorf("Detect() finding = %+v, want capped critical fanout", finding)
	}
}

func TestMssMetadataAnomalyDetectorFindsVersionAccumulation(t *testing.T) {
	root := t.TempDir()
	old := time.Now().Add(-30 * 24 * time.Hour)
	for _, version := range []string{"1.2.3", "1.3.0", "2.0.0", "2.1.0"} {
		path := filepath.Join(root, version)
		if err := os.Mkdir(path, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink("2.1.0", filepath.Join(root, "Current")); err != nil {
		t.Fatal(err)
	}
	detector := &MssMetadataAnomalyDetectorAdapter{MaxDirs: 1, MaxDepth: 1}

	scan, err := detector.Detect(context.Background(), []string{root}, time.Now())
	if err != nil {
		t.Fatalf("Detect() error = %v, want nil", err)
	}
	if len(scan.Findings) != 1 {
		t.Fatalf("Detect() findings = %d, want 1", len(scan.Findings))
	}
	finding := scan.Findings[0]
	if finding.Kind != "version-accumulation" || finding.VersionCount != 4 || finding.ActiveTarget != "2.1.0" || finding.StaleCount != 3 {
		t.Errorf("Detect() finding = %+v, want generalized version evidence", finding)
	}
}

func TestMssMetadataAnomalyDetectorFindsGeneratedQueueAndDoesNotFollowSymlinks(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "reports")
	outside := filepath.Join(parent, "outside")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-30 * 24 * time.Hour)
	for i := 0; i < 260; i++ {
		path := filepath.Join(root, fmt.Sprintf("report-%03d", i))
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Symlink(outside, filepath.Join(root, "linked-tree")); err != nil {
		t.Fatal(err)
	}
	detector := &MssMetadataAnomalyDetectorAdapter{MaxDirs: 10, MaxDepth: 3, MaxEntriesPerDir: 512}

	scan, err := detector.Detect(context.Background(), []string{root}, time.Now())
	if err != nil {
		t.Fatalf("Detect() error = %v, want nil", err)
	}
	if len(scan.Findings) != 1 || scan.Findings[0].Kind != "generated-queue" || !scan.Findings[0].Targeted {
		t.Errorf("Detect() findings = %+v, want generated queue", scan.Findings)
	}
	if scan.InspectedDirs != 1 {
		t.Errorf("Detect() inspected dirs = %d, want 1 (symlink not followed)", scan.InspectedDirs)
	}
}

func TestMssMetadataAnomalyDetectorMakesDirectoryBudgetExplicit(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "child"), 0o755); err != nil {
		t.Fatal(err)
	}
	detector := &MssMetadataAnomalyDetectorAdapter{MaxDirs: 1, MaxDepth: 3}

	scan, err := detector.Detect(context.Background(), []string{root}, time.Now())
	if err != nil {
		t.Fatalf("Detect() error = %v, want nil", err)
	}
	if !scan.Truncated || scan.InspectedDirs != 1 {
		t.Errorf("Detect() = %+v, want explicit truncation at one directory", scan)
	}
}

func TestMssMetadataAnomalyDetectorFindsLargeTemporaryPayloadWithoutProductRule(t *testing.T) {
	root := filepath.Join(t.TempDir(), "staging")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	payload := filepath.Join(root, "unsigned-copy.bin")
	handle, err := os.Create(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := handle.Truncate(6 << 30); err != nil {
		handle.Close()
		t.Fatal(err)
	}
	if err := handle.Close(); err != nil {
		t.Fatal(err)
	}
	detector := &MssMetadataAnomalyDetectorAdapter{MaxDirs: 1}

	scan, err := detector.Detect(context.Background(), []string{root}, time.Now())
	if err != nil {
		t.Fatalf("Detect() error = %v, want nil", err)
	}
	if len(scan.Findings) != 1 || scan.Findings[0].Kind != "large-direct-files" || scan.Findings[0].Severity != "critical" {
		t.Errorf("Detect() findings = %+v, want critical large-direct-files", scan.Findings)
	}
}

func TestMssMetadataAnomalyDetectorRejectsWeakVersionLookalike(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"1.0", "2.0", "3.0", "4.0"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	detector := &MssMetadataAnomalyDetectorAdapter{MaxDirs: 1}

	scan, err := detector.Detect(context.Background(), []string{root}, time.Now())
	if err != nil {
		t.Fatalf("Detect() error = %v, want nil", err)
	}
	if len(scan.Findings) != 0 {
		t.Errorf("Detect() findings = %+v, want weak version-like layout ignored", scan.Findings)
	}
}

func TestMssMetadataAnomalyDetectorPrioritizesGeneratedChainsWithinDirectoryBudget(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 20; i++ {
		if err := os.Mkdir(filepath.Join(root, fmt.Sprintf("ordinary-%02d", i)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	pending := filepath.Join(root, "code-sign-clone-staging", "pending")
	if err := os.MkdirAll(pending, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 260; i++ {
		if err := os.WriteFile(filepath.Join(pending, fmt.Sprintf("payload-%03d", i)), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	detector := &MssMetadataAnomalyDetectorAdapter{MaxDirs: 3, MaxDepth: 4, MaxEntriesPerDir: 512}

	scan, err := detector.Detect(context.Background(), []string{root}, time.Now())
	if err != nil {
		t.Fatalf("Detect() error = %v, want nil", err)
	}
	if len(scan.Findings) != 1 || scan.Findings[0].Path != pending {
		t.Errorf("Detect() findings = %+v, want prioritized nested generated queue", scan.Findings)
	}
}

func TestMssClassifyDirectoryAnomalyDoesNotAutoScanDenseDirectoryTrees(t *testing.T) {
	children := make([]string, 300)
	metadata := mssDirectoryMetadata{entryCount: 500, sampled: 64, staleSampled: 64, children: children}

	finding, ok := mssClassifyDirectoryAnomaly("/broad/tree", metadata)
	if !ok || finding.Kind != "stale-dense" || finding.Targeted {
		t.Errorf("mssClassifyDirectoryAnomaly() = (%+v, %v), want reported non-targeted dense tree", finding, ok)
	}
}
