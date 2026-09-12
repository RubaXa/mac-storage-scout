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

// MssVolumeUsage describes capacity and availability reported by the mounted filesystem.
//
// @purpose Carry the authoritative volume-level baseline used for scan reconciliation.
// @consumer internal/mss/app/mss_volume_audit_orchestrator.go
type MssVolumeUsage struct {
	Path           string
	CapacityBytes  int64
	AvailableBytes int64
	OccupiedBytes  int64
}

// MssVolumeAudit reconciles readable file allocations with volume-level occupancy.
//
// @purpose Make unexplained disk usage explicit instead of silently omitting it.
// @consumer internal/mss/adapters/report/mss_volume_audit_text_report_adapter.go
type MssVolumeAudit struct {
	ScanRoot         string
	Usage            MssVolumeUsage
	AccountedBytes   int64
	UnaccountedBytes int64
	OvercountBytes   int64
	GrowthDuringScan int64
	Counters         MssCounters
	Roots            []*MssNode
}

// MssTriageAgeBytes groups allocated bytes by last-modification age.
//
// @purpose Distinguish current churn from stale cleanup candidates.
// @consumer internal/mss/adapters/report/mss_triage_text_report_adapter.go
type MssTriageAgeBytes struct {
	Today     int64 `json:"today"`
	Week      int64 `json:"week"`
	Older     int64 `json:"older"`
	NoModTime int64 `json:"no_mod_time"`
}

// MssTriageHotspot describes one root or shallow descendant measured by triage.
//
// @purpose Carry comparable hotspot size, recency, activity, and safety evidence.
// @consumer internal/mss/app/mss_triage_orchestrator.go
type MssTriageHotspot struct {
	Path       string            `json:"path"`
	SizeBytes  int64             `json:"size_bytes"`
	DeltaBytes int64             `json:"delta_bytes"`
	Age        MssTriageAgeBytes `json:"age"`
	Processes  []string          `json:"processes,omitempty"`
	ProcessOK  bool              `json:"process_evidence"`
	Safety     string            `json:"safety"`
	Reason     string            `json:"reason"`
}

// MssTriageSnapshot is the compact persisted baseline used by later triage runs.
//
// @purpose Persist path totals without retaining file names or file contents.
// @consumer internal/mss/adapters/state/mss_json_triage_state_adapter.go
type MssTriageSnapshot struct {
	Version       int              `json:"version"`
	CapturedAt    time.Time        `json:"captured_at"`
	OccupiedBytes int64            `json:"occupied_bytes"`
	Roots         []string         `json:"roots"`
	Paths         map[string]int64 `json:"paths"`
}

// MssTriageConfig defines one fast incident scan and baseline update.
//
// @purpose Carry triage policy without coupling it to CLI flag parsing.
// @consumer internal/mss/app/mss_triage_orchestrator.go
type MssTriageConfig struct {
	Paths          []string
	AnomalyPaths   []string
	StatePath      string
	ThresholdBytes int64
	TopN           int
	SaveBaseline   bool
	Now            time.Time
}

// MssTriageReport is the outcome-first result of one incident triage run.
//
// @purpose Explain current hotspot occupancy and change since the previous baseline.
// @consumer internal/mss/adapters/report/mss_triage_text_report_adapter.go
type MssTriageReport struct {
	CapturedAt    time.Time
	PreviousAt    time.Time
	Usage         MssVolumeUsage
	GrowthBytes   int64
	Hotspots      []MssTriageHotspot
	AnomalyScan   MssTriageAnomalyScan
	Errors        int64
	BaselinePath  string
	BaselineSaved bool
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
