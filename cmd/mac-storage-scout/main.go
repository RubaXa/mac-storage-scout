// @task spec/tasks/mac-storage-scout.task-05.md
// @purpose Provide CLI entrypoint for storage scout commands.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	iFS "io/fs"
	"mac-storage-scout/internal/mss/adapters/aggregate"
	fsadapter "mac-storage-scout/internal/mss/adapters/fs"
	"mac-storage-scout/internal/mss/adapters/progress"
	"mac-storage-scout/internal/mss/adapters/report"
	stateadapter "mac-storage-scout/internal/mss/adapters/state"
	"mac-storage-scout/internal/mss/app"
	"mac-storage-scout/internal/mss/domain"
	"mac-storage-scout/internal/mss/ports"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// main dispatches CLI subcommands.
//
// @purpose Route process execution to scan/delete command handlers.
// @consumer End users invoking mac-storage-scout binary.
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "scan":
		runScan(os.Args[2:])
	case "audit":
		runAudit(os.Args[2:])
	case "triage":
		runTriage(os.Args[2:])
	case "delete":
		runDelete(os.Args[2:])
	default:
		printUsage()
		os.Exit(2)
	}
}

// printUsage prints command help text.
//
// @purpose Describe available CLI commands and flags.
// @consumer End users invoking mac-storage-scout binary.
func printUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  mac-storage-scout scan [--threshold 500MB] [--top 5] [--size-mode logical|allocated] [--profile macos-core] [--plain] [paths...]")
	fmt.Fprintln(os.Stderr, "  mac-storage-scout audit [--volume /System/Volumes/Data] [--threshold 5GB] [--top 10] [--plain]")
	fmt.Fprintln(os.Stderr, "  mac-storage-scout triage [--broad] [--anomaly-budget 5s] [--anomaly-max-dirs 20000] [--state <path>] [--threshold 500MB] [--top 20] [--no-save] [paths...]")
	fmt.Fprintln(os.Stderr, "  mac-storage-scout delete [--dry-run] [--yes] <path> [path...]")
}

// runTriage executes fast high-churn disk incident diagnosis.
//
// @purpose Compare common volatile macOS paths with a persistent baseline and process evidence.
// @consumer End users invoking mac-storage-scout triage.
// @param args Raw triage subcommand args.
func runTriage(args []string) {
	triageCmd := flag.NewFlagSet("triage", flag.ContinueOnError)
	statePath := triageCmd.String("state", mssDefaultTriageStatePath(), "persistent baseline JSON path")
	threshold := triageCmd.String("threshold", "500MB", "visible hotspot threshold")
	top := triageCmd.Int("top", 20, "maximum hotspots and candidates")
	noSave := triageCmd.Bool("no-save", false, "read baseline without updating it")
	noProcesses := triageCmd.Bool("no-processes", false, "skip best-effort lsof attribution")
	broad := triageCmd.Bool("broad", false, "also scan app data, containers, projects, and downloads")
	anomalyBudget := triageCmd.Duration("anomaly-budget", 5*time.Second, "maximum metadata-first anomaly discovery time")
	anomalyMaxDirs := triageCmd.Int("anomaly-max-dirs", 20000, "maximum directories inspected by anomaly preflight")
	if err := triageCmd.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	thresholdBytes, err := domain.MssParseBytes(*threshold)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid --threshold:", err)
		os.Exit(2)
	}
	if err := validateTopN(*top); err != nil {
		fmt.Fprintln(os.Stderr, "invalid --top:", err)
		os.Exit(2)
	}
	if *anomalyBudget <= 0 || *anomalyMaxDirs < 1 {
		fmt.Fprintln(os.Stderr, "triage: --anomaly-budget must be positive and --anomaly-max-dirs must be >= 1")
		os.Exit(2)
	}
	paths := triageCmd.Args()
	anomalyPaths := []string(nil)
	if len(paths) == 0 {
		paths = mssDefaultTriagePaths(*broad)
		anomalyPaths = mssDefaultTriageAnomalyPaths()
	} else {
		paths = mssExistingNonOverlappingPaths(paths)
		anomalyPaths = append([]string(nil), paths...)
	}
	if len(paths) == 0 {
		fmt.Fprintln(os.Stderr, "triage: no existing paths to scan")
		os.Exit(2)
	}
	processProbe := ports.MssProcessUsageProbePort(nil)
	if !*noProcesses {
		processProbe = &fsadapter.MssLsofProcessProbeAdapter{}
	}
	fmt.Fprintf(os.Stderr, "triage: metadata anomaly preflight (budget=%s, max-dirs=%d)\n", anomalyBudget.String(), *anomalyMaxDirs)
	orchestrator := &app.MssTriageOrchestrator{
		Collector: &fsadapter.MssTriageCollectorAdapter{OnRootComplete: func(path string, elapsed time.Duration, err error) {
			status := "done"
			if err != nil {
				status = "partial"
			}
			fmt.Fprintf(os.Stderr, "triage measured: %s elapsed=%s status=%s\n", path, elapsed.Round(time.Millisecond), status)
		}},
		Anomalies: &fsadapter.MssMetadataAnomalyDetectorAdapter{
			Budget:  *anomalyBudget,
			MaxDirs: *anomalyMaxDirs,
			OnFinding: func(finding domain.MssTriageAnomaly) {
				countPrefix := ""
				if finding.EntryCountMin {
					countPrefix = ">="
				}
				action := "reported for review"
				if finding.Targeted {
					action = "targeted measurement queued"
				}
				fmt.Fprintf(os.Stderr, "triage anomaly: [%s/%s] %s entries=%s%d dirs=%d; %s\n", finding.Severity, finding.Kind, finding.Path, countPrefix, finding.EntryCount, finding.DirectoryCount, action)
			},
		},
		State:     &stateadapter.MssJSONTriageStateAdapter{},
		Processes: processProbe,
		Usage:     &fsadapter.MssStatfsVolumeUsageAdapter{},
	}
	cfg := domain.MssTriageConfig{
		Paths:          paths,
		AnomalyPaths:   anomalyPaths,
		StatePath:      expandPath(*statePath),
		ThresholdBytes: thresholdBytes,
		TopN:           *top,
		SaveBaseline:   !*noSave,
		Now:            time.Now(),
	}
	triage, err := orchestrator.Run(context.Background(), cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "triage failed:", err)
		os.Exit(1)
	}
	if err := (&report.MssTriageTextReportAdapter{}).Render(os.Stdout, triage, cfg); err != nil {
		fmt.Fprintln(os.Stderr, "render failed:", err)
		os.Exit(1)
	}
}

// mssDefaultTriageAnomalyPaths returns broad metadata-only roots used even by fast triage.
//
// @purpose Find deeply nested structural anomalies without recursively sizing every broad root.
// @consumer runTriage anomaly detector wiring.
// @returns Existing, normalized, non-overlapping roots.
func mssDefaultTriageAnomalyPaths() []string {
	home, _ := os.UserHomeDir()
	paths := mssDefaultTriagePaths(true)
	paths = append(paths, "/Applications", filepath.Join(home, "Applications"))
	tempRoot := filepath.Clean(os.TempDir())
	tempContainer := filepath.Dir(tempRoot)
	if filepath.Base(tempRoot) == "T" && tempContainer != string(filepath.Separator) {
		paths = append(paths, tempContainer)
	}
	return mssExistingNonOverlappingPaths(paths)
}

// mssDefaultTriagePaths discovers common high-churn macOS paths without app-specific configuration.
//
// @purpose Cover agent state, package caches, app data, temp, swap, and update staging in one fast workflow.
// @consumer runTriage default path resolution.
// @param broad Include slower application, container, project, and download roots.
// @returns Existing, normalized, non-overlapping roots.
func mssDefaultTriagePaths(broad bool) []string {
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, "Library", "Caches"),
		"/private/tmp",
		os.TempDir(),
		"/private/var/vm",
		"/System/Volumes/Update",
	}
	for _, relative := range []string{
		".cache", ".local/share", ".npm", ".pnpm-store", ".yarn", ".gradle", ".m2",
		".cargo", ".rustup", ".docker", ".colima", ".codex", ".claude", ".gennady", ".Trash",
	} {
		paths = append(paths, filepath.Join(home, relative))
	}
	if broad {
		for _, relative := range []string{
			"Library/Application Support", "Library/Containers", "Library/Group Containers", "Developer", "Downloads",
		} {
			paths = append(paths, filepath.Join(home, relative))
		}
	}
	return mssExistingNonOverlappingPaths(paths)
}

// mssExistingNonOverlappingPaths normalizes roots and removes duplicate nested scans.
//
// @purpose Prevent double counting and redundant traversal in triage.
// @consumer runTriage path resolution.
// @param paths Raw candidate roots.
// @returns Existing absolute roots sorted lexicographically.
func mssExistingNonOverlappingPaths(paths []string) []string {
	unique := map[string]bool{}
	for _, raw := range paths {
		absolute, err := filepath.Abs(expandPath(raw))
		if err != nil {
			continue
		}
		if info, err := os.Stat(absolute); err == nil && info.IsDir() {
			unique[filepath.Clean(absolute)] = true
		}
	}
	ordered := make([]string, 0, len(unique))
	for path := range unique {
		ordered = append(ordered, path)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		leftDepth := strings.Count(ordered[i], string(filepath.Separator))
		rightDepth := strings.Count(ordered[j], string(filepath.Separator))
		if leftDepth == rightDepth {
			return ordered[i] < ordered[j]
		}
		return leftDepth < rightDepth
	})
	selected := make([]string, 0, len(ordered))
	for _, candidate := range ordered {
		nested := false
		for _, parent := range selected {
			if strings.HasPrefix(candidate, parent+string(filepath.Separator)) {
				nested = true
				break
			}
		}
		if !nested {
			selected = append(selected, candidate)
		}
	}
	sort.Strings(selected)
	return selected
}

// mssDefaultTriageStatePath returns the per-user baseline location.
//
// @purpose Keep triage state outside scanned cache roots and project worktrees.
// @consumer runTriage flag defaults.
// @returns Absolute default baseline path.
func mssDefaultTriageStatePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "mac-storage-scout", "triage-v1.json")
}

// runScan executes scan command flow.
//
// @purpose Parse scan flags and run scan pipeline.
// @consumer End users invoking mac-storage-scout scan.
// @param args Raw scan subcommand args.
func runScan(args []string) {
	// START_PARSE_SCAN_FLAGS
	fsCmd := flag.NewFlagSet("scan", flag.ContinueOnError)
	threshold := fsCmd.String("threshold", "500MB", "detail threshold")
	top := fsCmd.Int("top", 5, "top items in other bucket")
	sizeMode := fsCmd.String("size-mode", "logical", "logical|allocated")
	profile := fsCmd.String("profile", "", "named profile, e.g. macos-core")
	workers := fsCmd.Int("workers", runtime.NumCPU()*2, "worker count")
	progressOn := fsCmd.Bool("progress", true, "show progress")
	noProgress := fsCmd.Bool("no-progress", false, "disable progress")
	plain := fsCmd.Bool("plain", false, "force ASCII-only report output")
	if err := fsCmd.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	// END_PARSE_SCAN_FLAGS

	// START_VALIDATE_SCAN_CONFIG
	thBytes, err := domain.MssParseBytes(*threshold)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid --threshold:", err)
		os.Exit(2)
	}

	mode := domain.MssSizeMode(strings.ToLower(*sizeMode))
	if mode != domain.MssSizeModeLogical && mode != domain.MssSizeModeAllocated {
		fmt.Fprintln(os.Stderr, "invalid --size-mode: expected logical|allocated")
		os.Exit(2)
	}
	if err := validateTopN(*top); err != nil {
		fmt.Fprintln(os.Stderr, "invalid --top:", err)
		os.Exit(2)
	}
	// END_VALIDATE_SCAN_CONFIG

	// START_RESOLVE_SCAN_PATHS
	paths := fsCmd.Args()
	if *profile == "macos-core" {
		home, _ := os.UserHomeDir()
		paths = []string{
			filepath.Join(home, "Library", "Containers"),
			filepath.Join(home, "Library", "Application Support"),
			"/private/var/vm",
			"/private/var/folders",
		}
	}
	if len(paths) == 0 {
		paths = []string{"."}
	}
	// END_RESOLVE_SCAN_PATHS

	cfg := domain.MssScanConfig{
		Paths:          paths,
		ThresholdBytes: thBytes,
		TopN:           *top,
		SizeMode:       mode,
		Workers:        *workers,
		Progress:       *progressOn && !*noProgress,
		PlainOutput:    *plain,
	}

	orch := &app.MssScanOrchestrator{
		Walker:     &fsadapter.MssGoFsWalkerAdapter{},
		Aggregator: &aggregate.MssTreeAggregatorAdapter{},
		Progress:   &progress.MssAnsiProgressAdapter{},
	}

	// START_RUN_SCAN_PIPELINE
	// purpose: walk -> aggregate -> render while preserving non-fatal scan behavior.
	roots, _, runErr := orch.Run(context.Background(), cfg)
	if runErr != nil && !errors.Is(runErr, os.ErrPermission) {
		fmt.Fprintln(os.Stderr, "scan failed:", runErr)
		os.Exit(1)
	}

	renderer := &report.MssTreeTextReportAdapter{}
	if err := renderer.Render(os.Stdout, roots, cfg); err != nil {
		fmt.Fprintln(os.Stderr, "render failed:", err)
		os.Exit(1)
	}
	// END_RUN_SCAN_PIPELINE
}

// runAudit executes whole-volume scan reconciliation.
//
// @purpose Explain occupied volume bytes with readable paths and an explicit unaccounted remainder.
// @consumer End users invoking mac-storage-scout audit.
// @param args Raw audit subcommand args.
func runAudit(args []string) {
	// START_PARSE_AUDIT_FLAGS
	auditCmd := flag.NewFlagSet("audit", flag.ContinueOnError)
	volume := auditCmd.String("volume", mssDefaultAuditVolume(), "root to scan and reconcile against its containing volume")
	threshold := auditCmd.String("threshold", "5GB", "detail threshold")
	top := auditCmd.Int("top", 10, "top items in other bucket")
	workers := auditCmd.Int("workers", runtime.NumCPU()*2, "worker count")
	progressOn := auditCmd.Bool("progress", true, "show progress")
	noProgress := auditCmd.Bool("no-progress", false, "disable progress")
	plain := auditCmd.Bool("plain", false, "force ASCII-only report output")
	if err := auditCmd.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if len(auditCmd.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "audit: positional paths are not supported; use --volume")
		os.Exit(2)
	}
	// END_PARSE_AUDIT_FLAGS

	thBytes, err := domain.MssParseBytes(*threshold)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid --threshold:", err)
		os.Exit(2)
	}
	if err := validateTopN(*top); err != nil {
		fmt.Fprintln(os.Stderr, "invalid --top:", err)
		os.Exit(2)
	}

	volumePath, err := filepath.Abs(expandPath(*volume))
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid --volume:", err)
		os.Exit(2)
	}
	cfg := domain.MssScanConfig{
		Paths:          []string{volumePath},
		ThresholdBytes: thBytes,
		TopN:           *top,
		SizeMode:       domain.MssSizeModeAllocated,
		Workers:        *workers,
		Progress:       *progressOn && !*noProgress,
		PlainOutput:    *plain,
		OneFileSystem:  true,
	}

	scanner := &app.MssScanOrchestrator{
		Walker:     &fsadapter.MssGoFsWalkerAdapter{},
		Aggregator: &aggregate.MssTreeAggregatorAdapter{},
		Progress:   &progress.MssAnsiProgressAdapter{},
	}
	auditor := &app.MssVolumeAuditOrchestrator{
		Scanner: scanner,
		Usage:   &fsadapter.MssStatfsVolumeUsageAdapter{},
	}

	// START_RUN_VOLUME_AUDIT
	// invariant: volume accounting is sampled before and after the same allocated-size scan.
	audit, err := auditor.Run(context.Background(), volumePath, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "audit failed:", err)
		os.Exit(1)
	}
	renderer := &report.MssVolumeAuditTextReportAdapter{}
	if err := renderer.Render(os.Stdout, audit, cfg); err != nil {
		fmt.Fprintln(os.Stderr, "render failed:", err)
		os.Exit(1)
	}
	// END_RUN_VOLUME_AUDIT
}

// mssDefaultAuditVolume returns the macOS writable Data volume when present.
//
// @purpose Make whole-disk audit useful without requiring operators to know APFS mount paths.
// @consumer runAudit flag defaults.
// @returns Existing Data volume path on macOS, otherwise filesystem root.
func mssDefaultAuditVolume() string {
	const dataVolume = "/System/Volumes/Data"
	if _, err := os.Stat(dataVolume); err == nil {
		return dataVolume
	}
	return string(filepath.Separator)
}

// runDelete executes delete command flow.
//
// @purpose Parse delete flags and run guarded delete workflow.
// @consumer End users invoking mac-storage-scout delete.
// @param args Raw delete subcommand args.
func runDelete(args []string) {
	// START_PARSE_DELETE_FLAGS
	delCmd := flag.NewFlagSet("delete", flag.ContinueOnError)
	dryRun := delCmd.Bool("dry-run", false, "show deletion plan without deleting")
	yes := delCmd.Bool("yes", false, "confirm deletion without prompt")
	if err := delCmd.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	targets := delCmd.Args()
	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "delete: at least one target path is required")
		os.Exit(2)
	}
	if !*dryRun && !*yes {
		fmt.Fprintln(os.Stderr, "delete: pass --yes to execute deletion (or use --dry-run)")
		os.Exit(2)
	}
	// END_PARSE_DELETE_FLAGS

	summary := executeDeletePlan(os.Stdout, os.Stderr, targets, *dryRun, os.RemoveAll, measurePath)
	if summary.Failed > 0 || summary.Skipped > 0 {
		os.Exit(1)
	}
}

// mssDeleteSummary separates planned, confirmed, failed, and skipped delete outcomes.
//
// @purpose Prevent candidate bytes from being mistaken for successfully reclaimed bytes.
// @consumer executeDeletePlan.
type mssDeleteSummary struct {
	CandidateBytes int64
	ReclaimedBytes int64
	Deleted        int
	Failed         int
	Skipped        int
}

// executeDeletePlan applies guarded deletion and reports conservative before/after byte estimates.
//
// @purpose Keep failed or skipped targets out of the successful reclaimed total.
// @consumer runDelete.
// @pre Real deletion is authorized by the caller; remove and measure dependencies are non-nil.
// @param out Standard result writer.
// @param errOut Warning and failure writer.
// @param targets Raw deletion targets.
// @param dryRun Whether to plan without mutation.
// @param remove Injected recursive removal operation.
// @param measure Injected size measurement operation.
// @returns Candidate and confirmed-reclamation summary.
// @post Every failed removal is remeasured when possible and contributes only bytes no longer present.
// @invariant Protected paths are never passed to remove.
func executeDeletePlan(out, errOut io.Writer, targets []string, dryRun bool, remove func(string) error, measure func(string) (int64, error)) mssDeleteSummary {
	summary := mssDeleteSummary{}
	// START_EXECUTE_DELETE_PLAN_WITH_TRUTHFUL_TOTALS
	for _, raw := range targets {
		ap, err := filepath.Abs(expandPath(raw))
		if err != nil {
			fmt.Fprintf(errOut, "skip %s: %v\n", raw, err)
			summary.Skipped++
			continue
		}
		if err := guardDeletePath(ap); err != nil {
			fmt.Fprintf(errOut, "skip %s: %v\n", ap, err)
			summary.Skipped++
			continue
		}
		before, measureErr := measure(ap)
		if errors.Is(measureErr, iFS.ErrNotExist) {
			fmt.Fprintf(errOut, "skip %s: path does not exist\n", ap)
			summary.Skipped++
			continue
		}
		if measureErr != nil {
			fmt.Fprintf(errOut, "warn: size estimate failed for %s: %v\n", ap, measureErr)
			before = 0
		}
		summary.CandidateBytes += before
		if dryRun {
			fmt.Fprintf(out, "DRY-RUN delete %s (%s)\n", ap, domain.MssHumanBytes(before))
			continue
		}

		removeErr := remove(ap)
		after, afterErr := measure(ap)
		absent := errors.Is(afterErr, iFS.ErrNotExist)
		if absent {
			after = 0
		}
		reclaimed := int64(0)
		if afterErr == nil || absent {
			reclaimed = before - after
			if reclaimed < 0 {
				reclaimed = 0
			}
		}
		summary.ReclaimedBytes += reclaimed
		if removeErr != nil || !absent {
			if removeErr == nil {
				removeErr = fmt.Errorf("target still exists after removal")
			}
			fmt.Fprintf(errOut, "failed delete %s: %v; reclaimed estimate=%s\n", ap, removeErr, domain.MssHumanBytes(reclaimed))
			summary.Failed++
			continue
		}
		fmt.Fprintf(out, "DELETED %s (%s)\n", ap, domain.MssHumanBytes(reclaimed))
		summary.Deleted++
	}
	if dryRun {
		fmt.Fprintf(out, "DRY-RUN total candidate: %s; skipped=%d\n", domain.MssHumanBytes(summary.CandidateBytes), summary.Skipped)
	} else {
		fmt.Fprintf(out, "Reclaimed estimate: %s; deleted=%d failed=%d skipped=%d\n", domain.MssHumanBytes(summary.ReclaimedBytes), summary.Deleted, summary.Failed, summary.Skipped)
	}
	// END_EXECUTE_DELETE_PLAN_WITH_TRUTHFUL_TOTALS
	return summary
}

// guardDeletePath enforces protected-path policy.
//
// @purpose Prevent destructive delete against protected roots and aliases.
// @consumer runDelete delete pipeline.
// @param path Absolute candidate path.
// @returns Validation error when path is blocked.
func guardDeletePath(path string) error {
	clean := filepath.Clean(path)
	canonical := clean
	if resolved, err := filepath.EvalSymlinks(clean); err == nil {
		canonical = filepath.Clean(resolved)
	}

	if isProtectedDeletePath(clean) || isProtectedDeletePath(canonical) {
		return fmt.Errorf("refusing protected path: %s", path)
	}
	return nil
}

// isProtectedDeletePath checks whether path is protected.
//
// @purpose Determine whether path belongs to forbidden delete set.
// @consumer guardDeletePath safety policy.
// @param path Candidate cleaned path.
// @returns True when path is protected.
func isProtectedDeletePath(path string) bool {
	if path == "/" {
		return true
	}
	lp := strings.ToLower(path)
	for _, banned := range []string{"/System", "/usr", "/bin", "/sbin", "/private/var/vm"} {
		lb := strings.ToLower(banned)
		if lp == lb || strings.HasPrefix(lp, lb+"/") {
			return true
		}
	}
	return false
}

// validateTopN validates top list size.
//
// @purpose Enforce minimum top-N value for report contracts.
// @consumer runScan flag validation.
// @param top Requested top value.
// @returns Validation error when value is out of range.
func validateTopN(top int) error {
	if top < 1 {
		return fmt.Errorf("[validateTopN] top must be >= 1")
	}
	return nil
}

// measurePath estimates logical size for delete preview.
//
// @purpose Provide best-effort byte estimate for delete candidate reporting.
// @consumer runDelete dry-run and delete summaries.
// @param path Candidate path.
// @returns Estimated bytes and optional traversal error.
func measurePath(path string) (int64, error) {
	fi, err := os.Lstat(path)
	if err != nil {
		return 0, err
	}
	if !fi.IsDir() {
		return fi.Size(), nil
	}
	var total int64
	err = filepath.WalkDir(path, func(p string, d iFS.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > 0 {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

// expandPath expands user home shortcut.
//
// @purpose Normalize ~ paths before path validation and traversal.
// @consumer runDelete delete pipeline.
// @param p Raw user path.
// @returns Expanded path.
func expandPath(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		h, err := os.UserHomeDir()
		if err == nil {
			if p == "~" {
				return h
			}
			return filepath.Join(h, strings.TrimPrefix(p, "~/"))
		}
	}
	return p
}

var _ = time.Now
