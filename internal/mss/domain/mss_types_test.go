// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Validate threshold parsing and config constraints.
package domain

import "testing"

func TestMssParseBytesValid(t *testing.T) {
	tests := []struct {
		in   string
		want int64
	}{
		{"500MB", 500 * 1024 * 1024},
		{"1GB", 1024 * 1024 * 1024},
		{"2048", 2048},
	}
	for _, tc := range tests {
		got, err := MssParseBytes(tc.in)
		if err != nil {
			t.Fatalf("unexpected err for %s: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("want %d got %d", tc.want, got)
		}
	}
}

func TestMssParseBytesInvalid(t *testing.T) {
	bad := []string{"", "0", "-1", "abc"}
	for _, in := range bad {
		if _, err := MssParseBytes(in); err == nil {
			t.Fatalf("expected error for %q", in)
		}
	}
}
