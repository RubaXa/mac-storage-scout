// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Persist compact triage baselines as versioned JSON.
package state

import (
	"encoding/json"
	"fmt"
	"mac-storage-scout/internal/mss/domain"
	"os"
	"path/filepath"
)

// MssJSONTriageStateAdapter stores triage baselines atomically.
//
// @purpose Enable deterministic growth comparison across independent CLI runs.
// @consumer internal/mss/app/mss_triage_orchestrator.go
// @implements {MssTriageStatePort} internal/mss/ports/mss_triage_ports.go
type MssJSONTriageStateAdapter struct{}

// Load reads a baseline snapshot.
//
// @purpose Restore prior path totals for delta calculation.
// @consumer internal/mss/app/mss_triage_orchestrator.go
// @param path Baseline JSON path.
// @returns Decoded snapshot or read/decode error.
func (a *MssJSONTriageStateAdapter) Load(path string) (domain.MssTriageSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return domain.MssTriageSnapshot{}, err
	}
	var snapshot domain.MssTriageSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return domain.MssTriageSnapshot{}, fmt.Errorf("[MssJSONTriageStateAdapter.Load] decode baseline: %w", err)
	}
	return snapshot, nil
}

// Save atomically replaces a baseline snapshot.
//
// @purpose Avoid partial baseline state when a process is interrupted.
// @consumer internal/mss/app/mss_triage_orchestrator.go
// @param path Baseline JSON path.
// @param snapshot Completed triage snapshot.
// @returns Persistence error, if any.
func (a *MssJSONTriageStateAdapter) Save(path string, snapshot domain.MssTriageSnapshot) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("[MssJSONTriageStateAdapter.Save] encode baseline: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("[MssJSONTriageStateAdapter.Save] create state directory: %w", err)
	}
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o600); err != nil {
		return fmt.Errorf("[MssJSONTriageStateAdapter.Save] write temporary baseline: %w", err)
	}
	if err := os.Rename(temporary, path); err != nil {
		return fmt.Errorf("[MssJSONTriageStateAdapter.Save] replace baseline: %w", err)
	}
	return nil
}
