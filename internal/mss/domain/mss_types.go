// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define core scan and report domain structures.
package domain

import "time"

// MssSizeMode defines how file size is accounted during scans.
type MssSizeMode string

const (
	// MssSizeModeLogical uses logical file size (`st_size`).
	MssSizeModeLogical MssSizeMode = "logical"
	// MssSizeModeAllocated uses allocated blocks (`st_blocks*512`).
	MssSizeModeAllocated MssSizeMode = "allocated"
)

// MssScanConfig defines scan execution parameters.
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
type MssCounters struct {
	DirsScanned  int64
	FilesScanned int64
	BytesSeen    int64
	Errors       int64
	QueueDepth   int64
	StartedAt    time.Time
}

// MssEntryKind describes filesystem entry kind.
type MssEntryKind string

const (
	MssEntryKindFile MssEntryKind = "file"
	MssEntryKindDir  MssEntryKind = "dir"
)

// MssWalkEntry is one filesystem entry emitted by walker.
type MssWalkEntry struct {
	Path       string
	ParentPath string
	Name       string
	Kind       MssEntryKind
	SizeBytes  int64
	Ext        string
}

// MssWalkEvent wraps a walk entry or a non-fatal traversal error.
type MssWalkEvent struct {
	Entry *MssWalkEntry
	Err   error
}

// MssExtStat aggregates size and count per file extension.
type MssExtStat struct {
	Ext       string
	SizeBytes int64
	Count     int64
}

// MssOtherBucket summarizes items below threshold for one directory.
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
type MssNode struct {
	Path      string
	Name      string
	Kind      MssEntryKind
	SizeBytes int64
	Ext       string
	Children  []*MssNode
	Other     *MssOtherBucket
}
