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

// MssTreeTextReportAdapter renders threshold-aware sections to text output.
//
// @purpose Serialize aggregated nodes into deterministic report sections.
// @consumer cmd/mss/main.go
// @invariant Other bucket shape stays consistent: top-N, rest, types.
// @implements {MssReportComposerPort} internal/mss/ports/mss_report_composer_port.go
type MssTreeTextReportAdapter struct{}

// @see {MssReportComposerPort#Render} internal/mss/ports/mss_report_composer_port.go
// @purpose Render roots into terminal report output.
// @consumer cmd/mss/main.go
// @pre w is non-nil.
// @param w Output writer.
// @param roots Aggregated roots.
// @param cfg Scan configuration.
// @returns Render error.
// @post Report header and every root section are emitted in deterministic order.
func (a *MssTreeTextReportAdapter) Render(w io.Writer, roots []*domain.MssNode, cfg domain.MssScanConfig) error {
	if w == nil {
		return fmt.Errorf("[MssTreeTextReportAdapter.Render] writer is nil")
	}

	// START_RENDER_HEADER
	style := renderStyle{emoji: !cfg.PlainOutput}
	if style.emoji {
		fmt.Fprintf(w, "┌─ 🛰️  mss :: mac-storage-scout\n")
		fmt.Fprintf(w, "│  🎚️  threshold: %s | 🔝 top: %d | 📐 size-mode: %s\n", domain.MssHumanBytes(cfg.ThresholdBytes), cfg.TopN, cfg.SizeMode)
		fmt.Fprintf(w, "└─ 🧭 sections: %d\n\n", len(roots))
	} else {
		fmt.Fprintf(w, "mss :: mac-storage-scout\n")
		fmt.Fprintf(w, "threshold: %s | top: %d | size-mode: %s\n", domain.MssHumanBytes(cfg.ThresholdBytes), cfg.TopN, cfg.SizeMode)
		fmt.Fprintf(w, "sections: %d\n\n", len(roots))
	}
	// END_RENDER_HEADER

	// START_RENDER_ROOT_SECTIONS
	// invariant: each root renders exactly one section body and optional separator.
	for i, r := range roots {
		if i > 0 {
			fmt.Fprintln(w)
		}
		if style.emoji {
			fmt.Fprintf(w, "📂 [%s] %s\n", r.Path, domain.MssHumanBytes(r.SizeBytes))
		} else {
			fmt.Fprintf(w, "[%s] %s\n", r.Path, domain.MssHumanBytes(r.SizeBytes))
		}
		a.printNodeChildren(w, r, "", cfg.ThresholdBytes, style)
	}
	// END_RENDER_ROOT_SECTIONS
	return nil
}

// renderStyle controls output visual mode.
//
// @purpose Keep style switch explicit between emoji and plain render.
// @consumer MssTreeTextReportAdapter rendering methods.
type renderStyle struct {
	emoji bool
}

// printNodeChildren renders node subtree and optional other block.
//
// @purpose Serialize nested children with deterministic tree formatting.
// @consumer MssTreeTextReportAdapter.Render.
// @param w Output writer.
// @param n Node to render.
func (a *MssTreeTextReportAdapter) printNodeChildren(w io.Writer, n *domain.MssNode, indent string, threshold int64, style renderStyle) {
	count := len(n.Children)
	for i, c := range n.Children {
		isLast := i == count-1 && n.Other == nil
		branch := "├─"
		nextIndent := indent + "│  "
		if isLast {
			branch = "└─"
			nextIndent = indent + "   "
		}
		icon := ""
		if style.emoji {
			if c.Kind == domain.MssEntryKindDir {
				icon = "📁 "
			} else {
				icon = "📄 "
			}
		}
		fmt.Fprintf(w, "%s%s %s%s %s\n", indent, branch, icon, padRight(c.Name, 50), domain.MssHumanBytes(c.SizeBytes))
		if c.Kind == domain.MssEntryKindDir {
			a.printNodeChildren(w, c, nextIndent, threshold, style)
		}
	}

	if n.Other == nil {
		return
	}

	// START_RENDER_OTHER_BLOCK
	// purpose: preserve canonical small-items detail shape across all sections.
	topLabel := len(n.Other.TopItems)
	if topLabel == 0 {
		topLabel = 5
	}

	if style.emoji {
		fmt.Fprintf(w, "%s└─ 📦 other (<%s each, %d items) %s\n", indent, domain.MssHumanBytes(threshold), n.Other.Count, domain.MssHumanBytes(n.Other.SizeBytes))
	} else {
		fmt.Fprintf(w, "%s└─ other (<%s each, %d items) %s\n", indent, domain.MssHumanBytes(threshold), n.Other.Count, domain.MssHumanBytes(n.Other.SizeBytes))
	}

	otherIndent := indent + "   "
	if style.emoji {
		fmt.Fprintf(w, "%s├─ 🔝 top-%d:\n", otherIndent, topLabel)
	} else {
		fmt.Fprintf(w, "%s├─ top-%d:\n", otherIndent, topLabel)
	}
	for i, t := range n.Other.TopItems {
		b := "├─"
		if i == len(n.Other.TopItems)-1 {
			b = "└─"
		}
		if style.emoji {
			fmt.Fprintf(w, "%s│  %s 📌 %s %s\n", otherIndent, b, padRight(t.Name, 50), domain.MssHumanBytes(t.SizeBytes))
		} else {
			fmt.Fprintf(w, "%s│  %s %s %s\n", otherIndent, b, padRight(t.Name, 50), domain.MssHumanBytes(t.SizeBytes))
		}
	}
	if n.Other.RestCount > 0 {
		if style.emoji {
			fmt.Fprintf(w, "%s├─ 🧩 rest (%d items) %s\n", otherIndent, n.Other.RestCount, domain.MssHumanBytes(n.Other.RestSize))
		} else {
			fmt.Fprintf(w, "%s├─ rest (%d items) %s\n", otherIndent, n.Other.RestCount, domain.MssHumanBytes(n.Other.RestSize))
		}
	}

	if style.emoji {
		fmt.Fprintf(w, "%s└─ 🧪 types:\n", otherIndent)
	} else {
		fmt.Fprintf(w, "%s└─ types:\n", otherIndent)
	}
	for i, ts := range n.Other.TypeTop {
		b := "├─"
		if i == len(n.Other.TypeTop)-1 && n.Other.TypeRestCnt == 0 {
			b = "└─"
		}
		if style.emoji {
			fmt.Fprintf(w, "%s   %s 🏷️  %s %s (%d files)\n", otherIndent, b, padRight(ts.Ext, 18), domain.MssHumanBytes(ts.SizeBytes), ts.Count)
		} else {
			fmt.Fprintf(w, "%s   %s %s %s (%d files)\n", otherIndent, b, padRight(ts.Ext, 18), domain.MssHumanBytes(ts.SizeBytes), ts.Count)
		}
	}
	if n.Other.TypeRestCnt > 0 {
		if style.emoji {
			fmt.Fprintf(w, "%s   └─ 🏷️  %s %s (%d files)\n", otherIndent, padRight("rest types", 18), domain.MssHumanBytes(n.Other.TypeRestSz), n.Other.TypeRestCnt)
		} else {
			fmt.Fprintf(w, "%s   └─ %s %s (%d files)\n", otherIndent, padRight("rest types", 18), domain.MssHumanBytes(n.Other.TypeRestSz), n.Other.TypeRestCnt)
		}
	}
	// END_RENDER_OTHER_BLOCK
}

// padRight fills strings with dots for aligned text output.
//
// @purpose Improve human scan readability of report lines.
// @consumer MssTreeTextReportAdapter printing helpers.
// @param s Input label.
// @param width Target width.
// @returns Padded or unchanged label.
func padRight(s string, width int) string {
	if utf8.RuneCountInString(s) >= width {
		return s
	}
	return s + strings.Repeat(".", width-utf8.RuneCountInString(s))
}
