// @task spec/tasks/mac-storage-scout.task-05.md
// @purpose Validate CLI safety guards and parameter validation.
package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateTopN(t *testing.T) {
	if err := validateTopN(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := validateTopN(0); err == nil {
		t.Fatalf("expected error for top=0")
	}
	if err := validateTopN(-2); err == nil {
		t.Fatalf("expected error for negative top")
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
		if err := guardDeletePath(p); err == nil {
			t.Fatalf("expected protected path to be blocked: %s", p)
		}
	}
}

func TestGuardDeletePathBlocksSymlinkAliasToProtected(t *testing.T) {
	tmp := t.TempDir()
	alias := filepath.Join(tmp, "vm-alias")
	if err := os.Symlink("/private/var/vm", alias); err != nil {
		t.Fatalf("failed to create symlink fixture: %v", err)
	}

	if err := guardDeletePath(alias); err == nil {
		t.Fatalf("expected alias to protected path to be blocked")
	}
}

func TestGuardDeletePathAllowsRegularPath(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "safe")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	if err := guardDeletePath(target); err != nil {
		t.Fatalf("expected regular path to be allowed, got: %v", err)
	}
}
