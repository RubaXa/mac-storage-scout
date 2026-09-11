// @task spec/tasks/mac-storage-scout.task-05.md
// @purpose Validate CLI safety guards and parameter validation.
package main

import (
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
