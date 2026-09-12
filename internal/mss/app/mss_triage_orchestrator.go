// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Compare high-churn disk hotspots with a persistent baseline.
package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mac-storage-scout/internal/mss/domain"
	"mac-storage-scout/internal/mss/ports"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

// MssTriageOrchestrator coordinates hotspot collection, process attribution, and baseline state.
//
// @purpose Turn repeated disk incidents into one comparable diagnostic workflow.
// @consumer cmd/mac-storage-scout/main.go
// @invariant A failed optional process probe never suppresses filesystem findings.
type MssTriageOrchestrator struct {
	Collector ports.MssTriageCollectorPort
	Anomalies ports.MssTriageAnomalyDetectorPort
	State     ports.MssTriageStatePort
	Processes ports.MssProcessUsageProbePort
	Usage     ports.MssVolumeUsageProviderPort
}

// Run executes one triage scan and optionally updates its baseline.
//
// @purpose Report current high-churn occupancy, stale bytes, active owners, and change since baseline.
// @consumer cmd/mac-storage-scout/main.go
// @pre Paths, Collector, State, Usage, threshold, and top-N are valid.
// @param ctx Cancellation context.
// @param cfg Triage paths, output policy, and baseline settings.
// @returns Complete triage report or fatal collection/state error.
// @post A successful saved run atomically replaces the baseline after collection completes.
func (o *MssTriageOrchestrator) Run(ctx context.Context, cfg domain.MssTriageConfig) (domain.MssTriageReport, error) {
	if err := mssValidateTriageConfig(o, cfg); err != nil {
		return domain.MssTriageReport{}, err
	}
	now := cfg.Now
	if now.IsZero() {
		now = time.Now()
	}
	usage, err := o.Usage.Measure(cfg.Paths[0])
	if err != nil {
		return domain.MssTriageReport{}, fmt.Errorf("[MssTriageOrchestrator.Run] measure volume: %w", err)
	}
	anomalyScan := domain.MssTriageAnomalyScan{}
	if o.Anomalies != nil && len(cfg.AnomalyPaths) > 0 {
		anomalyScan, err = o.Anomalies.Detect(ctx, cfg.AnomalyPaths, now)
		if err != nil {
			return domain.MssTriageReport{}, fmt.Errorf("[MssTriageOrchestrator.Run] detect anomalies: %w", err)
		}
	}
	hotspots, errorCount, err := o.Collector.Collect(ctx, cfg.Paths, now)
	if err != nil {
		return domain.MssTriageReport{}, fmt.Errorf("[MssTriageOrchestrator.Run] collect hotspots: %w", err)
	}
	targetPaths := mssTriageTargetedPaths(cfg.Paths, anomalyScan.Findings)
	if len(targetPaths) > 0 {
		targeted, targetedErrors, targetErr := o.Collector.Collect(ctx, targetPaths, now)
		if targetErr != nil {
			return domain.MssTriageReport{}, fmt.Errorf("[MssTriageOrchestrator.Run] collect targeted anomalies: %w", targetErr)
		}
		errorCount += targetedErrors
		hotspots = mssMergeTriageHotspots(hotspots, targeted)
	}
	errorCount += anomalyScan.Errors
	measuredByPath := make(map[string]domain.MssTriageHotspot, len(hotspots))
	for _, hotspot := range hotspots {
		measuredByPath[hotspot.Path] = hotspot
	}
	for i := range anomalyScan.Findings {
		if measured, ok := measuredByPath[anomalyScan.Findings[i].Path]; ok {
			anomalyScan.Findings[i].SizeBytes = measured.SizeBytes
			anomalyScan.Findings[i].Age = measured.Age
		}
	}

	previous, loadErr := o.State.Load(cfg.StatePath)
	if loadErr != nil && !errors.Is(loadErr, fs.ErrNotExist) {
		errorCount++
		previous = domain.MssTriageSnapshot{}
	}
	hasPrevious := !previous.CapturedAt.IsZero() && slices.Equal(previous.Roots, cfg.Paths)
	currentPaths := make(map[string]bool, len(hotspots))
	for i := range hotspots {
		currentPaths[hotspots[i].Path] = true
		if hasPrevious {
			hotspots[i].DeltaBytes = hotspots[i].SizeBytes - previous.Paths[hotspots[i].Path]
		}
	}
	if hasPrevious {
		for path, previousSize := range previous.Paths {
			if !currentPaths[path] && previousSize > 0 {
				hotspots = append(hotspots, domain.MssTriageHotspot{Path: path, DeltaBytes: -previousSize})
			}
		}
	}

	// START_ATTRIBUTE_TOP_HOTSPOTS_TO_PROCESSES
	// purpose: cap lsof matching work while prioritizing paths most likely to explain pressure.
	probePaths := mssTriageProbePaths(hotspots, 120)
	if o.Processes != nil {
		activity, probeErr := o.Processes.Probe(ctx, probePaths)
		if probeErr != nil {
			errorCount++
		} else {
			for i := range hotspots {
				if mssContainsString(probePaths, hotspots[i].Path) {
					hotspots[i].ProcessOK = true
					hotspots[i].Processes = activity[hotspots[i].Path]
				}
			}
		}
	}
	for i := range hotspots {
		hotspots[i].Safety, hotspots[i].Reason = mssClassifyTriageHotspot(hotspots[i])
	}
	// END_ATTRIBUTE_TOP_HOTSPOTS_TO_PROCESSES

	sort.SliceStable(hotspots, func(i, j int) bool {
		if hotspots[i].SizeBytes == hotspots[j].SizeBytes {
			return hotspots[i].Path < hotspots[j].Path
		}
		return hotspots[i].SizeBytes > hotspots[j].SizeBytes
	})

	previousAt := time.Time{}
	if hasPrevious {
		previousAt = previous.CapturedAt
	}
	report := domain.MssTriageReport{
		CapturedAt:   now,
		PreviousAt:   previousAt,
		Usage:        usage,
		Hotspots:     hotspots,
		AnomalyScan:  anomalyScan,
		Errors:       errorCount,
		BaselinePath: cfg.StatePath,
	}
	if hasPrevious {
		report.GrowthBytes = usage.OccupiedBytes - previous.OccupiedBytes
	}
	if cfg.SaveBaseline {
		paths := make(map[string]int64, len(hotspots))
		for _, hotspot := range hotspots {
			paths[hotspot.Path] = hotspot.SizeBytes
		}
		snapshot := domain.MssTriageSnapshot{Version: 1, CapturedAt: now, OccupiedBytes: usage.OccupiedBytes, Roots: append([]string(nil), cfg.Paths...), Paths: paths}
		if err := o.State.Save(cfg.StatePath, snapshot); err != nil {
			return domain.MssTriageReport{}, fmt.Errorf("[MssTriageOrchestrator.Run] save baseline: %w", err)
		}
		report.BaselineSaved = true
	}
	return report, nil
}

// mssTriageTargetedPaths selects precise anomaly drill-down roots for a separate recursive scan.
//
// @purpose Convert cheap broad signals into precise allocated-size and age measurements.
// @consumer MssTriageOrchestrator.Run.
// @param roots Stable configured baseline roots.
// @param findings Metadata anomaly paths outside those roots.
// @returns Deterministically ordered roots that do not overlap each other and are not already visible within two base levels.
func mssTriageTargetedPaths(roots []string, findings []domain.MssTriageAnomaly) []string {
	selected := make([]string, 0, len(findings))
	appendIfDisjoint := func(candidate string) {
		candidate = filepath.Clean(candidate)
		covered := false
		for _, existing := range selected {
			if candidate == existing || strings.HasPrefix(candidate, existing+string(filepath.Separator)) || strings.HasPrefix(existing, candidate+string(filepath.Separator)) {
				covered = true
				break
			}
		}
		if !covered {
			selected = append(selected, candidate)
		}
	}
	// Findings arrive severity-ranked, so a precise critical path wins over a later broad low-severity parent.
	for _, finding := range findings {
		if finding.Path != "" && finding.Targeted && !mssTriagePathVisibleFromRoots(finding.Path, roots) {
			appendIfDisjoint(finding.Path)
		}
	}
	sort.Strings(selected)
	return selected
}

// mssTriagePathVisibleFromRoots checks whether shallow base aggregation already emits an exact path.
//
// @purpose Avoid redundant targeted scans for roots and their first two descendant levels.
// @consumer mssTriageTargetedPaths.
// @param path Anomaly path.
// @param roots Base collection roots.
// @returns True when base collection produces the exact anomaly hotspot.
func mssTriagePathVisibleFromRoots(path string, roots []string) bool {
	path = filepath.Clean(path)
	for _, root := range roots {
		root = filepath.Clean(root)
		if path == root {
			return true
		}
		relative, err := filepath.Rel(root, path)
		if err != nil || relative == "." || strings.HasPrefix(relative, "..") {
			continue
		}
		if len(strings.Split(relative, string(filepath.Separator))) <= 2 {
			return true
		}
	}
	return false
}

// mssMergeTriageHotspots combines base and targeted measurements by exact path.
//
// @purpose Add deep evidence without duplicate report nodes when a path is visible in both scans.
// @consumer MssTriageOrchestrator.Run.
// @param base Base-root hotspots.
// @param targeted Exact anomaly-root hotspots.
// @returns Deterministic de-duplicated hotspot slice.
func mssMergeTriageHotspots(base, targeted []domain.MssTriageHotspot) []domain.MssTriageHotspot {
	byPath := make(map[string]domain.MssTriageHotspot, len(base)+len(targeted))
	for _, hotspot := range base {
		byPath[hotspot.Path] = hotspot
	}
	for _, hotspot := range targeted {
		byPath[hotspot.Path] = hotspot
	}
	merged := make([]domain.MssTriageHotspot, 0, len(byPath))
	for _, hotspot := range byPath {
		merged = append(merged, hotspot)
	}
	return merged
}

// mssValidateTriageConfig rejects incomplete runtime wiring before filesystem work.
//
// @purpose Fail fast on invalid triage preconditions.
// @consumer MssTriageOrchestrator.Run.
// @param o Orchestrator dependencies.
// @param cfg Triage configuration.
// @returns Validation error, if any.
func mssValidateTriageConfig(o *MssTriageOrchestrator, cfg domain.MssTriageConfig) error {
	if o == nil || o.Collector == nil || o.State == nil || o.Usage == nil {
		return fmt.Errorf("[MssTriageOrchestrator.Run] required dependency is nil")
	}
	if len(cfg.Paths) == 0 || cfg.StatePath == "" {
		return fmt.Errorf("[MssTriageOrchestrator.Run] paths and state path are required")
	}
	if cfg.ThresholdBytes <= 0 || cfg.TopN < 1 {
		return fmt.Errorf("[MssTriageOrchestrator.Run] threshold must be positive and topN >= 1")
	}
	return nil
}

// mssTriageProbePaths selects the largest bounded process-attribution set.
//
// @purpose Prevent lsof matching cost from scaling with every shallow node.
// @consumer MssTriageOrchestrator.Run.
// @param hotspots Collected hotspot totals.
// @param limit Maximum probe paths.
// @returns Paths sorted by descending hotspot size.
func mssTriageProbePaths(hotspots []domain.MssTriageHotspot, limit int) []string {
	copyOf := append([]domain.MssTriageHotspot(nil), hotspots...)
	sort.SliceStable(copyOf, func(i, j int) bool { return copyOf[i].SizeBytes > copyOf[j].SizeBytes })
	if len(copyOf) > limit {
		copyOf = copyOf[:limit]
	}
	paths := make([]string, 0, len(copyOf))
	for _, hotspot := range copyOf {
		paths = append(paths, hotspot.Path)
	}
	return paths
}

// mssClassifyTriageHotspot applies conservative generic cleanup policy.
//
// @purpose Separate inactive stale cache/log/temp from active or review-required data.
// @consumer MssTriageOrchestrator.Run.
// @param hotspot Measured hotspot with optional process evidence.
// @returns Safety class and user-facing reason.
func mssClassifyTriageHotspot(hotspot domain.MssTriageHotspot) (string, string) {
	if len(hotspot.Processes) > 0 {
		return "active", "open files owned by a live process"
	}
	lower := strings.ToLower(filepath.ToSlash(hotspot.Path))
	stale := hotspot.Age.Older > 0
	cacheOrLog := strings.Contains(lower, "/cache") || strings.Contains(lower, "_cacache") || strings.Contains(lower, "/caches/") || strings.Contains(lower, "/logs/") || strings.HasSuffix(lower, "/log") || strings.Contains(lower, "/crashpad/pending")
	temporary := strings.Contains(lower, "/tmp/") || strings.HasSuffix(lower, "/tmp") || strings.Contains(lower, "/temporaryitems/")
	if hotspot.ProcessOK && stale && (cacheOrLog || temporary) {
		return "safe", "inactive cache/log/temp with bytes older than one week"
	}
	if strings.Contains(lower, "node_modules") || strings.Contains(lower, "/worktree") || strings.Contains(lower, "/snapshot") || strings.Contains(lower, "/sessions") {
		return "review", "generated or historical data; inspect ownership before deletion"
	}
	return "inspect", "large or growing data without a generic safe-delete rule"
}

// mssContainsString checks membership in a small bounded list.
//
// @purpose Mark only paths actually covered by the capped process probe as evidenced.
// @consumer MssTriageOrchestrator.Run.
// @param values Candidate path list.
// @param target Path to find.
// @returns True when target is present.
func mssContainsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
