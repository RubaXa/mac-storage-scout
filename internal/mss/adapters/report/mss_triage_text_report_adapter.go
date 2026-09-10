// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Render outcome-first incident triage findings.
package report

import (
	"fmt"
	"io"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MssTriageTextReportAdapter renders comparable hotspot and cleanup evidence.
//
// @purpose Make disk growth, stale bytes, and active ownership scannable in one report.
// @consumer cmd/mac-storage-scout/main.go
type MssTriageTextReportAdapter struct{}

// Render writes one deterministic triage report.
//
// @purpose Lead with volume outcome and rank actionable hotspots by observed growth.
// @consumer cmd/mac-storage-scout/main.go
// @pre cfg.TopN >= 1 and cfg.ThresholdBytes > 0.
// @param w Destination writer.
// @param triage Completed triage report.
// @param cfg Visibility and baseline configuration.
// @returns Render validation or write-independent formatting error.
func (a *MssTriageTextReportAdapter) Render(w io.Writer, triage domain.MssTriageReport, cfg domain.MssTriageConfig) error {
	if w == nil || cfg.TopN < 1 || cfg.ThresholdBytes <= 0 {
		return fmt.Errorf("[MssTriageTextReportAdapter.Render] invalid writer or config")
	}
	fmt.Fprintln(w, "mac-storage-scout triage")
	fmt.Fprintf(w, "captured: %s\n", triage.CapturedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "occupied: %s\n", domain.MssHumanBytes(triage.Usage.OccupiedBytes))
	fmt.Fprintf(w, "available: %s\n", domain.MssHumanBytes(triage.Usage.AvailableBytes))
	if triage.PreviousAt.IsZero() {
		fmt.Fprintln(w, "since-baseline: first run (deltas start next run)")
	} else {
		fmt.Fprintf(w, "since-baseline: %s | volume-growth: %s\n", triage.PreviousAt.Format("2006-01-02 15:04:05"), mssSignedHumanBytes(triage.GrowthBytes))
	}
	fmt.Fprintf(w, "scan-errors: %d\n", triage.Errors)
	fmt.Fprintf(w, "baseline: %s (%s)\n\n", triage.BaselinePath, map[bool]string{true: "updated", false: "read-only"}[triage.BaselineSaved])

	if !triage.PreviousAt.IsZero() {
		fmt.Fprintln(w, "changes:")
		changes := mssVisibleTriageChanges(triage.Hotspots, cfg)
		if len(changes) == 0 {
			fmt.Fprintln(w, "  none above threshold")
		}
		for _, hotspot := range changes {
			fmt.Fprintf(w, "  %s  delta=%s now=%s old>7d=%s\n", mssTildePath(hotspot.Path), mssSignedHumanBytes(hotspot.DeltaBytes), domain.MssHumanBytes(hotspot.SizeBytes), domain.MssHumanBytes(hotspot.Age.Older))
		}
		fmt.Fprintln(w)
	}

	visible := mssVisibleTriageHotspots(triage.Hotspots, cfg)
	fmt.Fprintln(w, "hotspots:")
	if len(visible) == 0 {
		fmt.Fprintln(w, "  none above threshold")
		return nil
	}
	for _, hotspot := range visible {
		activity := mssTriageOwnerSummary(hotspot.Processes, 3)
		fmt.Fprintf(w, "  %s  size=%s delta=%s old>7d=%s status=%s owner=%s\n",
			mssTildePath(hotspot.Path),
			domain.MssHumanBytes(hotspot.SizeBytes),
			mssSignedHumanBytes(hotspot.DeltaBytes),
			domain.MssHumanBytes(hotspot.Age.Older),
			hotspot.Safety,
			activity,
		)
	}

	candidates := append([]domain.MssTriageHotspot(nil), visible...)
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Safety == candidates[j].Safety {
			return candidates[i].Age.Older > candidates[j].Age.Older
		}
		return mssSafetyRank(candidates[i].Safety) < mssSafetyRank(candidates[j].Safety)
	})
	fmt.Fprintln(w, "\ncandidates:")
	shown := 0
	selectedPaths := []string{}
	for _, candidate := range candidates {
		if candidate.Safety != "safe" && candidate.Safety != "review" {
			continue
		}
		if mssTriagePathOverlaps(candidate.Path, selectedPaths) {
			continue
		}
		fmt.Fprintf(w, "  [%s] %s old-bytes=%s — %s\n", candidate.Safety, mssTildePath(candidate.Path), domain.MssHumanBytes(candidate.Age.Older), candidate.Reason)
		selectedPaths = append(selectedPaths, candidate.Path)
		shown++
		if shown == cfg.TopN {
			break
		}
	}
	if shown == 0 {
		fmt.Fprintln(w, "  none; inspect large inactive hotspots before deletion")
	}
	return nil
}

// mssTriagePathOverlaps detects parent/child overlap with selected candidates.
//
// @purpose Prevent reclaim estimates from counting the same bytes through multiple tree levels.
// @consumer MssTriageTextReportAdapter.Render.
// @param candidate Candidate path.
// @param selected Previously selected non-overlapping paths.
// @returns True when either path contains the other.
func mssTriagePathOverlaps(candidate string, selected []string) bool {
	candidate = filepath.Clean(candidate)
	for _, path := range selected {
		path = filepath.Clean(path)
		if candidate == path || strings.HasPrefix(candidate, path+string(filepath.Separator)) || strings.HasPrefix(path, candidate+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

// mssTriageOwnerSummary bounds process evidence rendered on one line.
//
// @purpose Keep broad active roots readable while preserving owner examples and total count.
// @consumer MssTriageTextReportAdapter.Render.
// @param processes Sorted process labels.
// @param limit Maximum explicit labels.
// @returns Inactive marker or compact owner summary.
func mssTriageOwnerSummary(processes []string, limit int) string {
	if len(processes) == 0 {
		return "inactive"
	}
	if limit < 1 || len(processes) <= limit {
		return strings.Join(processes, ",")
	}
	return strings.Join(processes[:limit], ",") + fmt.Sprintf(",+%d more", len(processes)-limit)
}

// mssVisibleTriageHotspots applies threshold and top-N report limits.
//
// @purpose Keep large or materially changed paths explicit and output bounded.
// @consumer MssTriageTextReportAdapter.Render.
// @param hotspots Ranked triage hotspots.
// @param cfg Visibility settings.
// @returns Visible hotspot subset.
func mssVisibleTriageHotspots(hotspots []domain.MssTriageHotspot, cfg domain.MssTriageConfig) []domain.MssTriageHotspot {
	ordered := append([]domain.MssTriageHotspot(nil), hotspots...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].SizeBytes == ordered[j].SizeBytes {
			return ordered[i].Path < ordered[j].Path
		}
		return ordered[i].SizeBytes > ordered[j].SizeBytes
	})
	visible := make([]domain.MssTriageHotspot, 0, cfg.TopN)
	for _, hotspot := range ordered {
		if hotspot.SizeBytes < cfg.ThresholdBytes {
			continue
		}
		visible = append(visible, hotspot)
		if len(visible) == cfg.TopN {
			break
		}
	}
	return visible
}

// mssVisibleTriageChanges selects material growth and shrinkage by absolute delta.
//
// @purpose Keep disappeared paths visible instead of losing them below current large hotspots.
// @consumer MssTriageTextReportAdapter.Render.
// @param hotspots Triage hotspots including prior-only paths.
// @param cfg Visibility settings.
// @returns Material changes ordered by absolute delta descending.
func mssVisibleTriageChanges(hotspots []domain.MssTriageHotspot, cfg domain.MssTriageConfig) []domain.MssTriageHotspot {
	ordered := append([]domain.MssTriageHotspot(nil), hotspots...)
	sort.SliceStable(ordered, func(i, j int) bool {
		left := mssAbs64(ordered[i].DeltaBytes)
		right := mssAbs64(ordered[j].DeltaBytes)
		if left == right {
			return ordered[i].Path < ordered[j].Path
		}
		return left > right
	})
	visible := make([]domain.MssTriageHotspot, 0, cfg.TopN)
	for _, hotspot := range ordered {
		if mssAbs64(hotspot.DeltaBytes) < cfg.ThresholdBytes {
			continue
		}
		visible = append(visible, hotspot)
		if len(visible) == cfg.TopN {
			break
		}
	}
	return visible
}

// mssSafetyRank maps safety labels to deterministic candidate order.
//
// @purpose Show safest candidates before review-required data.
// @consumer MssTriageTextReportAdapter.Render.
// @param safety Safety label.
// @returns Ascending sort rank.
func mssSafetyRank(safety string) int {
	switch safety {
	case "safe":
		return 0
	case "review":
		return 1
	case "inactive":
		return 2
	default:
		return 3
	}
}

// mssAbs64 returns an absolute signed-byte delta.
//
// @purpose Apply one threshold to growth and shrinkage.
// @consumer mssVisibleTriageHotspots.
// @param value Signed byte value.
// @returns Absolute byte value.
func mssAbs64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}

// mssTildePath shortens paths under the current user's home.
//
// @purpose Keep triage lines compact without making paths ambiguous.
// @consumer MssTriageTextReportAdapter.Render.
// @param path Absolute path.
// @returns Home-relative display path when possible.
func mssTildePath(path string) string {
	home, _ := os.UserHomeDir()
	if home != "" && (path == home || strings.HasPrefix(path, home+string(filepath.Separator))) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
