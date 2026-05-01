// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define report rendering contract.
package ports

import (
	"io"
	"mac-storage-scout/internal/mss/domain"
)

type MssReportComposerPort interface {
	Render(w io.Writer, roots []*domain.MssNode, cfg domain.MssScanConfig) error
}
