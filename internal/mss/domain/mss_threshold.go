// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Parse human threshold values into bytes.
package domain

import (
	"fmt"
	"strconv"
	"strings"
)

// MssParseBytes parses human-readable sizes like 500MB or 1GB into bytes.
//
// @purpose Normalize CLI threshold values into positive byte counts.
// @consumer cmd/mac-storage-scout/main.go
// @pre input is non-empty and contains a supported suffix or positive integer.
// @param input Human-readable or numeric size string.
// @returns Parsed bytes and optional parse/validation error.
// @post Returns bytes > 0 on success.
func MssParseBytes(input string) (int64, error) {
	raw := strings.TrimSpace(strings.ToUpper(input))
	if raw == "" {
		return 0, fmt.Errorf("[MssParseBytes] empty size value")
	}

	units := []struct {
		Suffix string
		Mul    int64
	}{
		{"GB", 1024 * 1024 * 1024},
		{"G", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
		{"M", 1024 * 1024},
		{"KB", 1024},
		{"K", 1024},
		{"B", 1},
	}

	for _, u := range units {
		if strings.HasSuffix(raw, u.Suffix) {
			num := strings.TrimSpace(strings.TrimSuffix(raw, u.Suffix))
			v, err := strconv.ParseFloat(num, 64)
			if err != nil {
				return 0, fmt.Errorf("[MssParseBytes] invalid size %q: %w", input, err)
			}
			out := int64(v * float64(u.Mul))
			if out <= 0 {
				return 0, fmt.Errorf("[MssParseBytes] size must be positive: %q", input)
			}
			return out, nil
		}
	}

	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("[MssParseBytes] invalid size %q", input)
	}
	if v <= 0 {
		return 0, fmt.Errorf("[MssParseBytes] size must be positive: %q", input)
	}
	return v, nil
}
