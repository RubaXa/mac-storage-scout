// @task spec/tasks/mac-storage-scout.task-05.md
// @purpose Validate CLI safety guards and parameter validation.
package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateTopN(t *testing.T) {
	testCases := []struct {
		name    string
		top     int
		wantErr bool
	}{
		{name: "accepts minimum valid value", top: 1, wantErr: false},
		{name: "rejects zero", top: 0, wantErr: true},
		{name: "rejects negative", top: -2, wantErr: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := validateTopN(tc.top)
			gotErr := err != nil
			if gotErr != tc.wantErr {
				t.Errorf("validateTopN(%d) error = %v, wantErr %v", tc.top, err, tc.wantErr)
			}
		})
	}
}

func TestGuardDeletePathBlocksProtectedAndDescendants(t *testing.T) {
	cases := []string{
		"/",
		"/private/var/vm",
		"/private/var/vm/swapfile0",
		"/System",
	}

	for _, p := range cases {
		p := p
		t.Run(p, func(t *testing.T) {
			if err := guardDeletePath(p); err == nil {
				t.Errorf("guardDeletePath(%q) error = nil, want protected path error", p)
			}
		})
	}
}

func TestGuardDeletePathBlocksSymlinkAliasToProtected(t *testing.T) {
	tmp := t.TempDir()
	alias := filepath.Join(tmp, "vm-alias")
	if err := os.Symlink("/private/var/vm", alias); err != nil {
		t.Fatalf("failed to create symlink fixture: %v", err)
	}

	if err := guardDeletePath(alias); err == nil {
		t.Errorf("guardDeletePath(%q) error = nil, want protected path error", alias)
	}
}

func TestGuardDeletePathAllowsRegularPath(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "safe")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	if err := guardDeletePath(target); err != nil {
		t.Errorf("guardDeletePath(%q) error = %v, want nil", target, err)
	}
}

func TestMssDefaultAuditVolumeReturnsExistingPath(t *testing.T) {
	volume := mssDefaultAuditVolume()
	if _, err := os.Stat(volume); err != nil {
		t.Errorf("mssDefaultAuditVolume() = %q, stat error = %v", volume, err)
	}
}

func TestExecuteDeletePlanExcludesFailedTargetFromReclaimedTotal(t *testing.T) {
	root := t.TempDir()
	removed := filepath.Join(root, "removed")
	failed := filepath.Join(root, "failed")
	if err := os.Mkdir(removed, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(failed, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(removed, "payload"), []byte("1234567890"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(failed, "payload"), []byte("12345678901234567890"), 0o600); err != nil {
		t.Fatal(err)
	}
	var out, errOut bytes.Buffer
	remove := func(path string) error {
		if path == failed {
			return errors.New("permission denied")
		}
		return os.RemoveAll(path)
	}

	summary := executeDeletePlan(&out, &errOut, []string{removed, failed}, false, remove, measurePath)
	if summary.CandidateBytes != 30 || summary.ReclaimedBytes != 10 || summary.Deleted != 1 || summary.Failed != 1 {
		t.Errorf("executeDeletePlan() = %+v, want candidate=30 reclaimed=10 deleted=1 failed=1", summary)
	}
	if _, err := os.Stat(failed); err != nil {
		t.Errorf("failed target should remain: %v", err)
	}
	if !bytes.Contains(errOut.Bytes(), []byte("permission denied")) || !bytes.Contains(out.Bytes(), []byte("failed=1")) {
		t.Errorf("executeDeletePlan() output missing failure summary:\nstdout=%s\nstderr=%s", out.String(), errOut.String())
	}
}

func TestExecuteDeletePlanCountsOnlyPartialBytesActuallyRemoved(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "partial")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	first := filepath.Join(target, "first")
	if err := os.WriteFile(first, []byte("1234567890"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "second"), []byte("12345678901234567890"), 0o600); err != nil {
		t.Fatal(err)
	}
	remove := func(string) error {
		if err := os.Remove(first); err != nil {
			return err
		}
		return errors.New("permission denied")
	}

	summary := executeDeletePlan(&bytes.Buffer{}, &bytes.Buffer{}, []string{target}, false, remove, measurePath)
	if summary.ReclaimedBytes != 10 || summary.Failed != 1 || summary.Deleted != 0 {
		t.Errorf("executeDeletePlan() = %+v, want partial reclaimed=10 and failed=1", summary)
	}
}
