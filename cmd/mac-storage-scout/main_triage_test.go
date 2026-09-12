// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Verify triage root discovery and overlap prevention.
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMssExistingNonOverlappingPaths(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatal(err)
	}
	got := mssExistingNonOverlappingPaths([]string{child, root, root, filepath.Join(root, "missing")})
	if len(got) != 1 || got[0] != root {
		t.Errorf("mssExistingNonOverlappingPaths() = %v, want [%s]", got, root)
	}
}

func TestMssDefaultTriageAnomalyPathsCoverUserTempContainer(t *testing.T) {
	paths := mssDefaultTriageAnomalyPaths()
	tempRoot := filepath.Clean(os.TempDir())
	want := tempRoot
	if filepath.Base(tempRoot) == "T" {
		want = filepath.Dir(tempRoot)
	}
	covered := false
	for _, path := range paths {
		if want == path || strings.HasPrefix(want, path+string(filepath.Separator)) {
			covered = true
			break
		}
	}
	if !covered {
		t.Errorf("mssDefaultTriageAnomalyPaths() = %v, want coverage for %s", paths, want)
	}
}

func TestMssDefaultTriagePathsAreExistingAndNonOverlapping(t *testing.T) {
	paths := mssDefaultTriagePaths(false)
	if len(paths) == 0 {
		t.Fatal("mssDefaultTriagePaths(false) returned no paths")
	}
	for i, path := range paths {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("mssDefaultTriagePaths(false)[%d] = %q, stat error = %v", i, path, err)
		}
		for j, other := range paths {
			if i != j && len(other) < len(path) && filepath.Dir(path) == other {
				t.Errorf("triage roots overlap: %q below %q", path, other)
			}
		}
	}
}
