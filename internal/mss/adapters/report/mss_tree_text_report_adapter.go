// @task spec/tasks/mac-storage-scout.task-04.md
// @purpose Render compact readable ASCII tree report.
package report

import (
	"fmt"
	"io"
	"mac-storage-scout/internal/mss/domain"
	"strings"
	"unicode/utf8"
)

type MssTreeTextReportAdapter struct{}

// @implements {MssReportComposerPort} internal/mss/ports/mss_report_composer_port.go
func (a *MssTreeTextReportAdapter) Render(w io.Writer, roots []*domain.MssNode, cfg domain.MssScanConfig) error {
	fmt.Fprintf(w, "┌─ 🛰️  mss :: mac-storage-scout\n")
	fmt.Fprintf(w, "│  🎚️  threshold: %s | 🔝 top: %d | 📐 size-mode: %s\n", domain.MssHumanBytes(cfg.ThresholdBytes), cfg.TopN, cfg.SizeMode)
	fmt.Fprintf(w, "└─ 🧭 sections: %d\n\n", len(roots))

	for i, r := range roots {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "📂 [%s] %s\n", r.Path, domain.MssHumanBytes(r.SizeBytes))
		a.printNodeChildren(w, r, "", cfg.ThresholdBytes)
	}
	return nil
}

func (a *MssTreeTextReportAdapter) printNodeChildren(w io.Writer, n *domain.MssNode, indent string, threshold int64) {
	count := len(n.Children)
	for i, c := range n.Children {
		isLast := i == count-1 && n.Other == nil
		branch := "├─"
		nextIndent := indent + "│  "
		if isLast {
			branch = "└─"
			nextIndent = indent + "   "
		}
		icon := "📁"
		if c.Kind != domain.MssEntryKindDir {
			icon = "📄"
		}
		fmt.Fprintf(w, "%s%s %s %s %s\n", indent, branch, icon, padRight(c.Name, 50), domain.MssHumanBytes(c.SizeBytes))
		if c.Kind == domain.MssEntryKindDir {
			a.printNodeChildren(w, c, nextIndent, threshold)
		}
	}

	if n.Other == nil {
		return
	}

	topLabel := len(n.Other.TopItems)
	if topLabel == 0 {
		topLabel = 5
	}

	fmt.Fprintf(w, "%s└─ 📦 other (<%s each, %d items) %s\n", indent, domain.MssHumanBytes(threshold), n.Other.Count, domain.MssHumanBytes(n.Other.SizeBytes))

	otherIndent := indent + "   "
	fmt.Fprintf(w, "%s├─ 🔝 top-%d:\n", otherIndent, topLabel)
	for i, t := range n.Other.TopItems {
		b := "├─"
		if i == len(n.Other.TopItems)-1 {
			b = "└─"
		}
		fmt.Fprintf(w, "%s│  %s 📌 %s %s\n", otherIndent, b, padRight(t.Name, 50), domain.MssHumanBytes(t.SizeBytes))
	}
	if n.Other.RestCount > 0 {
		fmt.Fprintf(w, "%s├─ 🧩 rest (%d items) %s\n", otherIndent, n.Other.RestCount, domain.MssHumanBytes(n.Other.RestSize))
	}

	fmt.Fprintf(w, "%s└─ 🧪 types:\n", otherIndent)
	for i, ts := range n.Other.TypeTop {
		b := "├─"
		if i == len(n.Other.TypeTop)-1 && n.Other.TypeRestCnt == 0 {
			b = "└─"
		}
		fmt.Fprintf(w, "%s   %s 🏷️  %s %s (%d files)\n", otherIndent, b, padRight(ts.Ext, 18), domain.MssHumanBytes(ts.SizeBytes), ts.Count)
	}
	if n.Other.TypeRestCnt > 0 {
		fmt.Fprintf(w, "%s   └─ 🏷️  %s %s (%d files)\n", otherIndent, padRight("rest types", 18), domain.MssHumanBytes(n.Other.TypeRestSz), n.Other.TypeRestCnt)
	}
}

func padRight(s string, width int) string {
	if utf8.RuneCountInString(s) >= width {
		return s
	}
	return s + strings.Repeat(".", width-utf8.RuneCountInString(s))
}
