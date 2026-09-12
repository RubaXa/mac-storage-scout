// @task spec/tasks/mac-storage-scout.task-11.md
// @purpose Carry bounded metadata-first anomaly evidence.
package domain

import "time"

// MssTriageAnomaly describes one suspicious filesystem structure found without a full size walk.
//
// @purpose Explain why a directory deserves targeted measurement before product-specific knowledge exists.
// @consumer internal/mss/app/mss_triage_orchestrator.go
type MssTriageAnomaly struct {
	Path           string
	Kind           string
	Severity       string
	Evidence       string
	Hint           string
	EntryCount     int64
	EntryCountMin  bool
	DirectoryCount int
	VersionCount   int
	StaleCount     int
	ActiveTarget   string
	SizeBytes      int64
	Age            MssTriageAgeBytes
	Targeted       bool
}

// MssTriageAnomalyScan summarizes bounded preflight coverage and findings.
//
// @purpose Make partial metadata coverage explicit instead of presenting it as a complete disk scan.
// @consumer internal/mss/adapters/report/mss_triage_text_report_adapter.go
type MssTriageAnomalyScan struct {
	Findings      []MssTriageAnomaly
	InspectedDirs int
	Errors        int64
	Truncated     bool
	Elapsed       time.Duration
}
