// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define report rendering contract.
package ports

import (
	"io"
	"mac-storage-scout/internal/mss/domain"
)

// MssReportComposerPort renders aggregated roots into user-facing output.
//
// @purpose Define report rendering boundary for CLI output.
// @consumer cmd/mss/main.go
type MssReportComposerPort interface {
	// Render writes sectioned report output from aggregated roots.
	//
	// @purpose Serialize contract-compliant report to output writer.
	// @consumer cmd/mss/main.go
	// @param w Output stream.
	// @param roots Aggregated roots to render.
	// @param cfg Scan configuration affecting report style.
	// @returns Optional render error.
	Render(w io.Writer, roots []*domain.MssNode, cfg domain.MssScanConfig) error
}
