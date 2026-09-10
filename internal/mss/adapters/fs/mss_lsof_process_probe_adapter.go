// @task spec/tasks/mac-storage-scout.task-10.md
// @purpose Attribute open files to triage hotspot paths on macOS.
package fs

import (
	"bufio"
	"context"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// MssLsofProcessProbeAdapter maps lsof machine output to hotspot paths.
//
// @purpose Mark cleanup candidates active when a live process has an open file below them.
// @consumer internal/mss/app/mss_triage_orchestrator.go
// @implements {MssProcessUsageProbePort} internal/mss/ports/mss_triage_ports.go
type MssLsofProcessProbeAdapter struct{}

// Probe runs lsof in machine-readable mode and returns compact process labels.
//
// @purpose Add best-effort process evidence without making lsof a hard dependency.
// @consumer internal/mss/app/mss_triage_orchestrator.go
// @param ctx Cancellation context.
// @param paths Candidate hotspot paths.
// @returns Process labels by hotspot and optional lsof error.
// @post Missing lsof or denied process access returns an error and no false activity labels.
func (a *MssLsofProcessProbeAdapter) Probe(ctx context.Context, paths []string) (map[string][]string, error) {
	cmd := exec.CommandContext(ctx, "lsof", "-F", "pcn")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	sets := map[string]map[string]bool{}
	var pid int
	var command string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 2 {
			continue
		}
		switch line[0] {
		case 'p':
			pid, _ = strconv.Atoi(line[1:])
		case 'c':
			command = line[1:]
		case 'n':
			openPath := filepath.Clean(line[1:])
			for _, candidate := range paths {
				candidate = filepath.Clean(candidate)
				if openPath != candidate && !strings.HasPrefix(openPath, candidate+string(filepath.Separator)) {
					continue
				}
				if sets[candidate] == nil {
					sets[candidate] = map[string]bool{}
				}
				sets[candidate][command+"("+strconv.Itoa(pid)+")"] = true
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	result := map[string][]string{}
	for path, labels := range sets {
		for label := range labels {
			result[path] = append(result[path], label)
		}
		sort.Strings(result[path])
	}
	return result, nil
}
