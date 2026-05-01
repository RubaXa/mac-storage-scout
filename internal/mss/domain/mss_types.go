// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define core scan and report domain structures.
package domain

import "time"

// MssSizeMode defines how file size is accounted during scans.
//
// @purpose Declare supported file-size accounting modes.
// @consumer cmd/mac-storage-scout/main.go
type MssSizeMode string

const (
	// MssSizeModeLogical uses logical file size (`st_size`).
	MssSizeModeLogical MssSizeMode = "logical"
	// MssSizeModeAllocated uses allocated blocks (`st_blocks*512`).
	MssSizeModeAllocated MssSizeMode = "allocated"
)

// MssScanConfig defines scan execution parameters.
//
// @purpose Carry scan runtime parameters across adapters.
// @consumer cmd/mac-storage-scout/main.go
type MssScanConfig struct {
	Paths          []string
	ThresholdBytes int64
	TopN           int
	SizeMode       MssSizeMode
	Workers        int
	Progress       bool
	PlainOutput    bool
	OneFileSystem  bool
}

// MssCounters tracks scan progress and completion statistics.
//
// @purpose Carry runtime counters for progress and final summary.
// @consumer internal/mss/app/mss_scan_orchestrator.go
type MssCounters struct {
	DirsScanned  int64
	FilesScanned int64
	BytesSeen    int64
	Errors       int64
	QueueDepth   int64
	StartedAt    time.Time
}

// MssEntryKind describes filesystem entry kind.
//
// @purpose Classify entries for aggregation/reporting logic.
// @consumer internal/mss/adapters/fs/mss_go_fs_walker_adapter.go
type MssEntryKind string

const (
	MssEntryKindFile MssEntryKind = "file"
	MssEntryKindDir  MssEntryKind = "dir"
)

// MssWalkEntry is one filesystem entry emitted by walker.
//
// @purpose Represent normalized walker payload for aggregation.
// @consumer internal/mss/adapters/fs/mss_go_fs_walker_adapter.go
type MssWalkEntry struct {
	Path       string
	ParentPath string
	Name       string
	Kind       MssEntryKind
	SizeBytes  int64
	Ext        string
}

// MssWalkEvent wraps a walk entry or a non-fatal traversal error.
//
// @purpose Preserve event or error boundary between walker and aggregator flow.
// @consumer internal/mss/app/mss_scan_orchestrator.go
type MssWalkEvent struct {
	Entry *MssWalkEntry
	Err   error
}

// MssExtStat aggregates size and count per file extension.
//
// @purpose Carry extension aggregation stats for other/types block.
// @consumer internal/mss/adapters/aggregate/mss_tree_aggregator_adapter.go
type MssExtStat struct {
	Ext       string
	SizeBytes int64
	Count     int64
}

// MssOtherBucket summarizes items below threshold for one directory.
//
// @purpose Preserve small-item rollup structure for report contract.
// @consumer internal/mss/adapters/aggregate/mss_tree_aggregator_adapter.go
type MssOtherBucket struct {
	Count       int
	SizeBytes   int64
	TopItems    []*MssNode
	RestCount   int
	RestSize    int64
	TypeTop     []MssExtStat
	TypeRestCnt int64
	TypeRestSz  int64
}

// MssNode is an aggregated filesystem tree node used by report rendering.
//
// @purpose Represent report-ready aggregated tree nodes.
// @consumer internal/mss/adapters/report/mss_tree_text_report_adapter.go
type MssNode struct {
	Path      string
	Name      string
	Kind      MssEntryKind
	SizeBytes int64
	Ext       string
	Children  []*MssNode
	Other     *MssOtherBucket
}
