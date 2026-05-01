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

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "scan":
		runScan(os.Args[2:])
	case "delete":
		runDelete(os.Args[2:])
	default:
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  mss scan [--threshold 500MB] [--top 5] [--size-mode logical|allocated] [--profile macos-core] [paths...]")
	fmt.Fprintln(os.Stderr, "  mss delete [--dry-run] [--yes] <path> [path...]")
}

func runScan(args []string) {
	fsCmd := flag.NewFlagSet("scan", flag.ContinueOnError)
	threshold := fsCmd.String("threshold", "500MB", "detail threshold")
	top := fsCmd.Int("top", 5, "top items in other bucket")
	sizeMode := fsCmd.String("size-mode", "logical", "logical|allocated")
	profile := fsCmd.String("profile", "", "named profile, e.g. macos-core")
	workers := fsCmd.Int("workers", runtime.NumCPU()*2, "worker count")
	progressOn := fsCmd.Bool("progress", true, "show progress")
	noProgress := fsCmd.Bool("no-progress", false, "disable progress")
	if err := fsCmd.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

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

	cfg := domain.MssScanConfig{
		Paths:          paths,
		ThresholdBytes: thBytes,
		TopN:           *top,
		SizeMode:       mode,
		Workers:        *workers,
		Progress:       *progressOn && !*noProgress,
	}

	orch := &app.MssScanOrchestrator{
		Walker:     &fsadapter.MssGoFsWalkerAdapter{},
		Aggregator: &aggregate.MssTreeAggregatorAdapter{},
		Progress:   &progress.MssAnsiProgressAdapter{},
	}

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
}

func runDelete(args []string) {
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
}

func guardDeletePath(path string) error {
	clean := filepath.Clean(path)
	if clean == "/" {
		return fmt.Errorf("refusing to delete root")
	}
	for _, banned := range []string{"/System", "/usr", "/bin", "/sbin", "/private/var/vm"} {
		if clean == banned {
			return fmt.Errorf("refusing protected path: %s", banned)
		}
	}
	return nil
}

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
