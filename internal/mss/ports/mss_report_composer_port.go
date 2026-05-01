// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define report rendering contract.
package ports

import (
	"io"
	"mac-storage-scout/internal/mss/domain"
)

// MssReportComposerPort renders aggregated roots into user-facing output.
type MssReportComposerPort interface {
	Render(w io.Writer, roots []*domain.MssNode, cfg domain.MssScanConfig) error
}
