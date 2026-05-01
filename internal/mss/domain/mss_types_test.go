// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Validate threshold parsing and config constraints.
package domain

import "testing"

func TestMssParseBytesValid(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int64
	}{
		{name: "parses megabytes", in: "500MB", want: 500 * 1024 * 1024},
		{name: "parses gigabytes", in: "1GB", want: 1024 * 1024 * 1024},
		{name: "parses raw bytes", in: "2048", want: 2048},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got, err := MssParseBytes(tc.in)
			if err != nil {
				t.Fatalf("MssParseBytes(%q) unexpected error: %v", tc.in, err)
			}
			if got != tc.want {
				t.Errorf("MssParseBytes(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestMssParseBytesInvalid(t *testing.T) {
	bad := []string{"", "0", "-1", "abc"}
	for _, in := range bad {
		in := in
		t.Run(in, func(t *testing.T) {
			if _, err := MssParseBytes(in); err == nil {
				t.Errorf("MssParseBytes(%q) error = nil, want non-nil", in)
			}
		})
	}
}
