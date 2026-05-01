// @task spec/tasks/mac-storage-scout.task-01.md
// @purpose Define core scan and report domain structures.
package domain

import "time"

type MssSizeMode string

const (
	MssSizeModeLogical   MssSizeMode = "logical"
	MssSizeModeAllocated MssSizeMode = "allocated"
)

type MssScanConfig struct {
	Paths          []string
	ThresholdBytes int64
	TopN           int
	SizeMode       MssSizeMode
	Workers        int
	Progress       bool
	OneFileSystem  bool
}

type MssCounters struct {
	DirsScanned  int64
	FilesScanned int64
	BytesSeen    int64
	Errors       int64
	QueueDepth   int64
	StartedAt    time.Time
}

type MssEntryKind string

const (
	MssEntryKindFile MssEntryKind = "file"
	MssEntryKindDir  MssEntryKind = "dir"
)

type MssWalkEntry struct {
	Path       string
	ParentPath string
	Name       string
	Kind       MssEntryKind
	SizeBytes  int64
	Ext        string
}

type MssWalkEvent struct {
	Entry *MssWalkEntry
	Err   error
}

type MssExtStat struct {
	Ext       string
	SizeBytes int64
	Count     int64
}

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

type MssNode struct {
	Path      string
	Name      string
	Kind      MssEntryKind
	SizeBytes int64
	Ext       string
	Children  []*MssNode
	Other     *MssOtherBucket
}
