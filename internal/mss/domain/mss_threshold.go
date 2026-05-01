// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Parse human threshold values into bytes.
package domain

import (
	"fmt"
	"strconv"
	"strings"
)

func MssParseBytes(input string) (int64, error) {
	raw := strings.TrimSpace(strings.ToUpper(input))
	if raw == "" {
		return 0, fmt.Errorf("empty size value")
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
				return 0, fmt.Errorf("invalid size %q: %w", input, err)
			}
			out := int64(v * float64(u.Mul))
			if out <= 0 {
				return 0, fmt.Errorf("size must be positive: %q", input)
			}
			return out, nil
		}
	}

	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q", input)
	}
	if v <= 0 {
		return 0, fmt.Errorf("size must be positive: %q", input)
	}
	return v, nil
}
