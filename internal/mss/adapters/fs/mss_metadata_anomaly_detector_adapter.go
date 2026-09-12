// @task spec/tasks/mac-storage-scout.task-11.md
// @purpose Detect suspicious directory structures with bounded metadata reads.
package fs

import (
	"context"
	"errors"
	"fmt"
	"io"
	iofs "io/fs"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
)

const (
	mssDefaultAnomalyBudget         = 5 * time.Second
	mssDefaultAnomalyMaxDirs        = 20000
	mssDefaultAnomalyMaxDepth       = 8
	mssDefaultAnomalyMaxEntries     = 4096
	mssDefaultAnomalyMetadataSample = 64
)

var mssVersionDirectoryPattern = regexp.MustCompile(`^\d+(?:\.\d+){1,5}(?:[-_][A-Za-z0-9._-]+)?$`)

// MssMetadataAnomalyDetectorAdapter performs a broad but bounded metadata preflight.
//
// @purpose Find structural disk-risk signals before expensive recursive byte accounting.
// @consumer internal/mss/app/mss_triage_orchestrator.go
// @implements {MssTriageAnomalyDetectorPort} internal/mss/ports/mss_triage_anomaly_port.go
// @invariant File contents and symlink targets are never traversed.
type MssMetadataAnomalyDetectorAdapter struct {
	Budget           time.Duration
	MaxDirs          int
	MaxDepth         int
	MaxEntriesPerDir int
	MetadataSample   int
	OnFinding        func(domain.MssTriageAnomaly)
}

// mssAnomalyDirectory is one bounded breadth-first traversal item.
//
// @purpose Carry path depth without recursive call-stack growth.
// @consumer MssMetadataAnomalyDetectorAdapter.Detect.
type mssAnomalyDirectory struct {
	path  string
	depth int
}

// mssDirectoryMetadata contains bounded direct-child evidence for one directory.
//
// @purpose Keep structural anomaly policy independent from filesystem iteration.
// @consumer mssClassifyDirectoryAnomaly.
type mssDirectoryMetadata struct {
	entryCount    int64
	entryCountMin bool
	nlink         uint64
	versionCount  int
	staleVersions int
	activeTarget  string
	sampled       int
	staleSampled  int
	largeFiles    int
	largestFile   int64
	sampledFiles  int
	sampledBytes  int64
	children      []string
}

// Detect inspects broad roots under explicit time, depth, directory, and fanout budgets.
//
// @purpose Produce early product-independent evidence and targeted drill-down paths.
// @consumer internal/mss/app/mss_triage_orchestrator.go
// @pre roots contain existing directories or best-effort inaccessible paths.
// @param ctx Cancellation context.
// @param roots Broad metadata roots.
// @param now Stable age boundary.
// @returns Bounded anomaly scan and fatal caller cancellation or unexpected filesystem error.
// @post Findings are sorted by severity, structural magnitude, then path.
// @invariant Internal budget exhaustion returns partial evidence with Truncated=true, while caller cancellation is fatal.
func (a *MssMetadataAnomalyDetectorAdapter) Detect(ctx context.Context, roots []string, now time.Time) (domain.MssTriageAnomalyScan, error) {
	started := time.Now()
	budget, maxDirs, maxDepth, maxEntries, metadataSample := a.mssLimits()
	queue := make([]mssAnomalyDirectory, 0, len(roots))
	seen := make(map[string]bool, len(roots))
	for _, root := range roots {
		clean := filepath.Clean(root)
		if !seen[clean] {
			seen[clean] = true
			queue = append(queue, mssAnomalyDirectory{path: clean})
		}
	}
	scan := domain.MssTriageAnomalyScan{}

	// START_BOUNDED_METADATA_BREADTH_FIRST_DISCOVERY
	// invariant: broad discovery must terminate even when a directory contains millions of entries.
	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return scan, err
		}
		if scan.InspectedDirs >= maxDirs || time.Since(started) >= budget {
			scan.Truncated = true
			break
		}
		current := queue[0]
		queue = queue[1:]
		metadata, errorsSeen, err := mssInspectAnomalyDirectory(current.path, now, maxEntries, metadataSample)
		scan.InspectedDirs++
		scan.Errors += errorsSeen
		if err != nil {
			if errors.Is(err, iofs.ErrNotExist) || errors.Is(err, iofs.ErrPermission) {
				scan.Errors++
				continue
			}
			return scan, fmt.Errorf("[MssMetadataAnomalyDetectorAdapter.Detect] inspect %s: %w", current.path, err)
		}
		if finding, ok := mssClassifyDirectoryAnomaly(current.path, metadata); ok {
			scan.Findings = append(scan.Findings, finding)
			if a.OnFinding != nil {
				a.OnFinding(finding)
			}
		}
		if current.depth >= maxDepth {
			if len(metadata.children) > 0 {
				scan.Truncated = true
			}
			continue
		}
		sort.SliceStable(metadata.children, func(i, j int) bool {
			left, right := mssAnomalyTraversalPriority(metadata.children[i]), mssAnomalyTraversalPriority(metadata.children[j])
			if left == right {
				return metadata.children[i] < metadata.children[j]
			}
			return left > right
		})
		urgent := make([]mssAnomalyDirectory, 0, len(metadata.children))
		normal := make([]mssAnomalyDirectory, 0, len(metadata.children))
		for _, child := range metadata.children {
			if seen[child] || !mssShouldDescendForAnomalies(child) {
				continue
			}
			seen[child] = true
			item := mssAnomalyDirectory{path: child, depth: current.depth + 1}
			if mssAnomalyTraversalPriority(child) > 0 {
				urgent = append(urgent, item)
			} else {
				normal = append(normal, item)
			}
		}
		queue = append(urgent, queue...)
		queue = append(queue, normal...)
	}
	// END_BOUNDED_METADATA_BREADTH_FIRST_DISCOVERY

	scan.Elapsed = time.Since(started)
	sort.SliceStable(scan.Findings, func(i, j int) bool {
		left, right := mssAnomalySeverityRank(scan.Findings[i].Severity), mssAnomalySeverityRank(scan.Findings[j].Severity)
		if left != right {
			return left < right
		}
		if scan.Findings[i].EntryCount != scan.Findings[j].EntryCount {
			return scan.Findings[i].EntryCount > scan.Findings[j].EntryCount
		}
		return scan.Findings[i].Path < scan.Findings[j].Path
	})
	return scan, nil
}

// mssLimits resolves zero-valued adapter settings to bounded production defaults.
//
// @purpose Keep direct construction safe and deterministic.
// @consumer MssMetadataAnomalyDetectorAdapter.Detect.
// @returns Budget, directory limit, depth limit, entry limit, and metadata sample size.
func (a *MssMetadataAnomalyDetectorAdapter) mssLimits() (time.Duration, int, int, int, int) {
	budget := a.Budget
	if budget <= 0 {
		budget = mssDefaultAnomalyBudget
	}
	maxDirs := a.MaxDirs
	if maxDirs <= 0 {
		maxDirs = mssDefaultAnomalyMaxDirs
	}
	maxDepth := a.MaxDepth
	if maxDepth <= 0 {
		maxDepth = mssDefaultAnomalyMaxDepth
	}
	maxEntries := a.MaxEntriesPerDir
	if maxEntries <= 0 {
		maxEntries = mssDefaultAnomalyMaxEntries
	}
	metadataSample := a.MetadataSample
	if metadataSample <= 0 {
		metadataSample = mssDefaultAnomalyMetadataSample
	}
	return budget, maxDirs, maxDepth, maxEntries, metadataSample
}

// mssInspectAnomalyDirectory samples direct children without descending or following symlinks.
//
// @purpose Bound work per directory while preserving fanout, version, age, and large-file evidence.
// @consumer MssMetadataAnomalyDetectorAdapter.Detect.
// @param path Directory to inspect.
// @param now Stable age boundary.
// @param maxEntries Maximum direct entries to read.
// @param metadataSample Maximum entries to stat for age and size.
// @returns Metadata evidence, non-fatal metadata error count, and directory read error.
func mssInspectAnomalyDirectory(path string, now time.Time, maxEntries, metadataSample int) (mssDirectoryMetadata, int64, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return mssDirectoryMetadata{}, 0, err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return mssDirectoryMetadata{}, 0, nil
	}
	metadata := mssDirectoryMetadata{}
	if stat, ok := info.Sys().(*syscall.Stat_t); ok && stat != nil {
		metadata.nlink = uint64(stat.Nlink)
	}
	handle, err := os.Open(path)
	if err != nil {
		return metadata, 0, err
	}
	defer handle.Close()
	var errorsSeen int64
	for {
		entries, readErr := handle.ReadDir(256)
		for _, entry := range entries {
			metadata.entryCount++
			if metadata.entryCount > int64(maxEntries) {
				metadata.entryCount = int64(maxEntries)
				metadata.entryCountMin = true
				return metadata, errorsSeen, nil
			}
			entryPath := filepath.Join(path, entry.Name())
			if entry.Type()&os.ModeSymlink != 0 {
				if mssIsActiveVersionLink(entry.Name()) {
					if target, linkErr := os.Readlink(entryPath); linkErr == nil {
						metadata.activeTarget = filepath.Base(filepath.Clean(target))
					} else {
						errorsSeen++
					}
				}
				continue
			}
			if entry.IsDir() {
				metadata.children = append(metadata.children, entryPath)
			}
			needsInfo := metadata.sampled < metadataSample || (entry.IsDir() && mssVersionDirectoryPattern.MatchString(entry.Name()))
			if !needsInfo {
				continue
			}
			entryInfo, infoErr := entry.Info()
			if infoErr != nil {
				errorsSeen++
				continue
			}
			if metadata.sampled < metadataSample {
				metadata.sampled++
				allocated := mssTriageAllocatedSize(entryInfo)
				if !entryInfo.IsDir() {
					metadata.sampledFiles++
					metadata.sampledBytes += allocated
					if allocated > metadata.largestFile {
						metadata.largestFile = allocated
					}
					if allocated >= 2<<30 {
						metadata.largeFiles++
					}
				}
				if now.Sub(entryInfo.ModTime()) >= 14*24*time.Hour {
					metadata.staleSampled++
				}
			}
			if entryInfo.IsDir() && mssVersionDirectoryPattern.MatchString(entry.Name()) {
				metadata.versionCount++
				if now.Sub(entryInfo.ModTime()) >= 14*24*time.Hour {
					metadata.staleVersions++
				}
			}
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			return metadata, errorsSeen, readErr
		}
	}
	if metadata.activeTarget != "" && mssVersionDirectoryPattern.MatchString(metadata.activeTarget) {
		if activeInfo, activeErr := os.Stat(filepath.Join(path, metadata.activeTarget)); activeErr == nil && now.Sub(activeInfo.ModTime()) >= 14*24*time.Hour && metadata.staleVersions > 0 {
			metadata.staleVersions--
		}
	}
	return metadata, errorsSeen, nil
}

// mssClassifyDirectoryAnomaly converts structural evidence into one strongest finding.
//
// @purpose Rank generalized signals without product-specific paths or names.
// @consumer MssMetadataAnomalyDetectorAdapter.Detect.
// @param path Inspected directory path.
// @param metadata Bounded direct-child evidence.
// @returns Strongest finding and whether evidence crossed a detection threshold.
func mssClassifyDirectoryAnomaly(path string, metadata mssDirectoryMetadata) (domain.MssTriageAnomaly, bool) {
	generated := mssLooksGeneratedDirectory(filepath.Base(path))
	extremeFanout := metadata.entryCountMin || metadata.entryCount >= 10000 || metadata.nlink >= 10000
	highFanout := metadata.entryCount >= 1000 || metadata.nlink >= 1000
	staleDense := metadata.entryCount >= 500 && metadata.sampled >= 16 && metadata.staleSampled*10 >= metadata.sampled*9
	finding := domain.MssTriageAnomaly{
		Path:           path,
		EntryCount:     metadata.entryCount,
		EntryCountMin:  metadata.entryCountMin,
		DirectoryCount: len(metadata.children),
		VersionCount:   metadata.versionCount,
		StaleCount:     metadata.staleVersions,
		ActiveTarget:   metadata.activeTarget,
	}
	switch {
	case metadata.largeFiles > 0 || metadata.sampledBytes >= 2<<30:
		finding.Kind = "large-direct-files"
		finding.Severity = "high"
		if metadata.sampledBytes >= 5<<30 {
			finding.Severity = "critical"
		}
		finding.Evidence = fmt.Sprintf("%d sampled direct files use %s; %d are >=2GB; largest=%s", metadata.sampledFiles, domain.MssHumanBytes(metadata.sampledBytes), metadata.largeFiles, domain.MssHumanBytes(metadata.largestFile))
		finding.Hint = "targeted size/age scan queued; verify whether these are temporary copies or required payloads"
		finding.Targeted = true
	case metadata.versionCount >= 5 || (metadata.versionCount >= 3 && metadata.activeTarget != ""):
		finding.Kind = "version-accumulation"
		finding.Severity = "medium"
		if metadata.versionCount >= 10 {
			finding.Severity = "high"
		}
		finding.Evidence = fmt.Sprintf("%d version-like sibling directories", metadata.versionCount)
		if metadata.activeTarget != "" {
			finding.Evidence += fmt.Sprintf("; active link -> %s", metadata.activeTarget)
		}
		if metadata.staleVersions > 0 {
			finding.Evidence += fmt.Sprintf("; %d older than 14d", metadata.staleVersions)
		}
		finding.Hint = "keep the active/current version; verify and review older siblings"
		finding.Targeted = true
	case extremeFanout || highFanout:
		finding.Kind = "extreme-fanout"
		finding.Severity = "high"
		if extremeFanout {
			finding.Severity = "critical"
		}
		countPrefix := ""
		if metadata.entryCountMin {
			countPrefix = ">="
		}
		finding.Evidence = fmt.Sprintf("%s%d direct entries; directory link-count=%d", countPrefix, metadata.entryCount, metadata.nlink)
		if generated {
			finding.Evidence += "; queue/cache/report-like name"
		}
		finding.Hint = "targeted size/age scan queued; identify the producer before cleanup"
		finding.Targeted = true
	case generated && metadata.entryCount >= 250:
		finding.Kind = "generated-queue"
		finding.Severity = "medium"
		finding.Evidence = fmt.Sprintf("%d direct entries in a queue/cache/report-like directory", metadata.entryCount)
		finding.Hint = "check live owners, then review stale generated files"
		finding.Targeted = true
	case staleDense:
		finding.Kind = "stale-dense"
		finding.Severity = "medium"
		finding.Evidence = fmt.Sprintf("%d direct entries; %d/%d sampled entries older than 14d", metadata.entryCount, metadata.staleSampled, metadata.sampled)
		finding.Targeted = len(metadata.children)*2 < int(metadata.entryCount)
		if finding.Targeted {
			finding.Hint = "targeted size/age scan queued; review ownership and regeneration cost"
		} else {
			finding.Hint = "dense directory tree reported without auto-scan; run explicit triage when its ownership is relevant"
		}
	default:
		return domain.MssTriageAnomaly{}, false
	}
	return finding, true
}

// mssLooksGeneratedDirectory detects generic queue/cache/report naming tokens.
//
// @purpose Add context to structural evidence without coupling to an application.
// @consumer mssClassifyDirectoryAnomaly.
// @param name Directory base name.
// @returns True for generic generated-data tokens.
func mssLooksGeneratedDirectory(name string) bool {
	tokens := strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	for _, token := range tokens {
		switch token {
		case "cache", "caches", "crash", "crashes", "crashpad", "dump", "dumps", "log", "logs", "pending", "report", "reports", "temp", "tmp":
			return true
		}
	}
	return false
}

// mssIsActiveVersionLink identifies conventional active-version symlink names.
//
// @purpose Distinguish a version store from unrelated numeric directories.
// @consumer mssInspectAnomalyDirectory.
// @param name Directory entry name.
// @returns True for conventional active-version labels.
func mssIsActiveVersionLink(name string) bool {
	switch strings.ToLower(name) {
	case "current", "latest", "stable", "active":
		return true
	default:
		return false
	}
}

// mssShouldDescendForAnomalies prunes known high-cardinality dependency internals after inspecting their root.
//
// @purpose Preserve broad-root budget for independent structures.
// @consumer MssMetadataAnomalyDetectorAdapter.Detect.
// @param path Candidate child directory.
// @returns False when deeper metadata adds little anomaly value.
func mssShouldDescendForAnomalies(path string) bool {
	switch strings.ToLower(filepath.Base(path)) {
	case ".git", ".svn", "node_modules", "vendor":
		return false
	default:
		return true
	}
}

// mssAnomalyTraversalPriority prioritizes conventional generated/version containers at equal breadth.
//
// @purpose Reach likely anomaly evidence earlier within a fixed time budget.
// @consumer MssMetadataAnomalyDetectorAdapter.Detect.
// @param path Candidate child directory.
// @returns Higher integer for earlier traversal.
func mssAnomalyTraversalPriority(path string) int {
	name := strings.ToLower(filepath.Base(path))
	if mssLooksGeneratedDirectory(name) || mssIsActiveVersionLink(name) || strings.Contains(name, "clone") || strings.Contains(name, "staging") || name == "versions" || name == "frameworks" || name == "contents" {
		return 1
	}
	return 0
}

// mssAnomalySeverityRank maps severity labels to deterministic output order.
//
// @purpose Lead with structures most likely to represent runaway disk usage.
// @consumer MssMetadataAnomalyDetectorAdapter.Detect.
// @param severity Finding severity label.
// @returns Ascending sort rank.
func mssAnomalySeverityRank(severity string) int {
	switch severity {
	case "critical":
		return 0
	case "high":
		return 1
	default:
		return 2
	}
}
