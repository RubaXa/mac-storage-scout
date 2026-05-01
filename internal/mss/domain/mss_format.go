// @task spec/tasks/mac-storage-scout.task-04.md
// @purpose Provide stable human-readable size formatting.
package domain

import "fmt"

// MssHumanBytes formats bytes in binary units for report readability.
//
// @purpose Keep size rendering compact and deterministic across all outputs.
// @consumer internal/mss/adapters/report/mss_tree_text_report_adapter.go
func MssHumanBytes(v int64) string {
	if v < 0 {
		v = 0
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	f := float64(v)
	i := 0
	for f >= 1024 && i < len(units)-1 {
		f /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d%s", v, units[i])
	}
	if f >= 100 {
		return fmt.Sprintf("%.0f%s", f, units[i])
	}
	if f >= 10 {
		return fmt.Sprintf("%.1f%s", f, units[i])
	}
	return fmt.Sprintf("%.2f%s", f, units[i])
}
