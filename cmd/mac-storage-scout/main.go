// @task spec/tasks/mac-storage-scout.task-05.md
// @purpose Provide CLI entrypoint for storage scout commands.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	iFS "io/fs"
	"mac-storage-scout/internal/mss/adapters/aggregate"
	fsadapter "mac-storage-scout/internal/mss/adapters/fs"
	"mac-storage-scout/internal/mss/adapters/progress"
	"mac-storage-scout/internal/mss/adapters/report"
	"mac-storage-scout/internal/mss/app"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
	"runtime"
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
	fmt.Fprintln(os.Stderr, "  mac-storage-scout delete [--dry-run] [--yes] <path> [path...]")
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

	// START_EXECUTE_DELETE_PLAN
	// invariant: protected paths are never deleted, and dry-run always reports candidate totals.
	var total int64
	for _, raw := range targets {
		p := expandPath(raw)
		ap, err := filepath.Abs(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", raw, err)
			continue
		}
		if err := guardDeletePath(ap); err != nil {
			fmt.Fprintf(os.Stderr, "skip %s: %v\n", ap, err)
			continue
		}
		sz, err := measurePath(ap)
		if err != nil && !errors.Is(err, iFS.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "warn: size estimate failed for %s: %v\n", ap, err)
		}
		total += sz

		if *dryRun {
			fmt.Printf("DRY-RUN delete %s (%s)\n", ap, domain.MssHumanBytes(sz))
			continue
		}

		if err := os.RemoveAll(ap); err != nil {
			fmt.Fprintf(os.Stderr, "failed delete %s: %v\n", ap, err)
			continue
		}
		fmt.Printf("DELETED %s (%s)\n", ap, domain.MssHumanBytes(sz))
	}

	if *dryRun {
		fmt.Printf("DRY-RUN total candidate: %s\n", domain.MssHumanBytes(total))
	} else {
		fmt.Printf("Deleted total estimated: %s\n", domain.MssHumanBytes(total))
	}
	// END_EXECUTE_DELETE_PLAN
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
