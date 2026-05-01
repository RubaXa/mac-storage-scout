package contractlint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunDetectsRequiredTags(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, filepath.Join(root, "a.go"), `package sample

// Thing represents demo type.
// @purpose Demo type purpose.
// @consumer tests
// @param not used here
// @returns not used here
type Thing struct{}

// Build does work.
// @purpose Build purpose.
func Build(name string) (string, error) { return name, nil }
`)

	report, err := Run(Config{Root: root})
	if err != nil {
		t.Fatalf("Run(...) unexpected error: %v", err)
	}
	if report.ErrorCount == 0 {
		t.Fatalf("Run(...) errors = %d, want > 0", report.ErrorCount)
	}

	foundBuild := false
	for _, e := range report.Entities {
		if e.Entity.Name != "Build" {
			continue
		}
		foundBuild = true
		if len(e.RequiredMiss) == 0 {
			t.Errorf("Build required misses = none, want missing tags")
		}
	}
	if !foundBuild {
		t.Fatalf("Build entity not found in report")
	}
}

func TestRunSkipsOptionalWhenRequiredPresent(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, filepath.Join(root, "b.go"), `package sample

// Sum sums numbers.
// @purpose Calculate sum.
// @consumer callers
// @param x first
// @param y second
// @returns computed value
func Sum(x int, y int) int { return x + y }
`)

	report, err := Run(Config{Root: root})
	if err != nil {
		t.Fatalf("Run(...) unexpected error: %v", err)
	}
	if report.ErrorCount != 0 {
		t.Fatalf("Run(...) errors = %d, want 0", report.ErrorCount)
	}
	for _, e := range report.Entities {
		if e.Entity.Name != "Sum" {
			continue
		}
		if len(e.Findings) != 0 {
			t.Errorf("Sum findings = %v, want none", e.Findings)
		}
	}
}

func TestRunChecksInterfaceMethods(t *testing.T) {
	root := t.TempDir()
	writeGoFile(t, filepath.Join(root, "c.go"), `package sample

// Worker performs work.
// @purpose Interface purpose.
// @consumer app
type Worker interface {
	// Do executes action.
	// @purpose method purpose
	// @consumer app
	Do(id string) (string, error)
}
`)

	report, err := Run(Config{Root: root})
	if err != nil {
		t.Fatalf("Run(...) unexpected error: %v", err)
	}

	foundMethod := false
	for _, e := range report.Entities {
		if e.Entity.Name == "Worker.Do" {
			foundMethod = true
		}
	}
	if !foundMethod {
		t.Fatalf("expected interface method Worker.Do in entity list")
	}
}

func writeGoFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}
