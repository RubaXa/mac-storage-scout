# Task: [TSK-09] - Performance Benchmark Suite & Engine Optimization Research

## 1. Meta & Traceability
- **Purpose**: Prove that `mac-storage-scout` is maximally fast on macOS/APFS by establishing a reproducible benchmark baseline, profiling the hot path, identifying concrete optimization candidates, and documenting which wins are worth taking vs. which add unacceptable complexity.
- **Dependencies**: TSK-02 (walker), TSK-05 (orchestrator), TSK-08 (git-flow).
- **Supporting Artifacts**:
  - `internal/mss/adapters/fs/mss_go_fs_walker_adapter.go`
  - `internal/mss/app/mss_scan_orchestrator.go`
  - `spec/mac-storage-scout.macos-performance.reference.md`
  - `spec/evidence/` (benchmark results land here)
- **Runtime Fidelity**: `benchmark + potential runtime change` (optimization PRs may follow as separate TSK-10+).
- **Target Files** (research + bench phase):
  - `internal/mss/adapters/fs/mss_walker_bench_test.go` (new — Go benchmark suite)
  - `spec/mac-storage-scout.macos-performance.reference.md` (update with measured data)
  - `spec/evidence/benchmark-baseline.txt` (new — captured `go test -bench` output)
  - `spec/evidence/pprof-cpu-profile.txt` (new — top20 pprof flamegraph summary)

## 2. Research Goals

### 2.1 Baseline Measurement
Capture wall-time, allocs/op, and ns/op for the current worker-pool walker across tree sizes:
- Small: ~10k entries (e.g. `~/go/src` or similar)
- Medium: ~100k entries (e.g. `~/Library`)
- Large: ~500k+ entries (e.g. `/Users/<user>` or `/usr/local`)

Metrics to capture per run:
- `ns/op` (normalized per directory entry)
- `allocs/op`
- `B/op`
- Peak RSS (via `/usr/bin/time -l` or `runtime.ReadMemStats`)
- Wall time end-to-end

### 2.2 Comparison Implementations to Benchmark
Run the same tree through each strategy and record the same metrics:

| ID  | Strategy                         | Notes                                                  |
|-----|----------------------------------|--------------------------------------------------------|
| C0  | Current worker pool (baseline)   | `MssGoFsWalkerAdapter` as-is                          |
| C1  | `filepath.WalkDir` sequential    | Stdlib reference, no parallelism                       |
| C2  | `filepath.Walk` sequential       | Older stdlib, calls `Stat` per entry                  |
| C3  | Worker pool, workers = NumCPU    | Lower worker count than current `NumCPU*2`             |
| C4  | Worker pool, workers = NumCPU*4  | Higher worker count                                    |
| C5  | Streaming (channel emit, no buf) | Replace `events []MssWalkEvent` slice with channel     |
| C6  | Pre-allocated events slice       | `make([]MssWalkEvent, 0, estimatedFileCount)`          |

Do NOT introduce `cgo`, external syscall packages, or `getdents` shims — keep Go stdlib only unless the research conclusively shows ≥30% wall-time gain from a specific syscall path.

### 2.3 Profiling Checkpoints
Run `go test -bench . -cpuprofile cpu.prof -memprofile mem.prof` and inspect:
1. Top 20 CPU symbols — identify if hotspot is in `os.ReadDir`, `os.Lstat`, `filepath.Join`, mutex contention, or channel ops.
2. Heap profile — identify if events slice is the dominant allocator.
3. Lock contention — `go test -bench . -mutexprofile mutex.prof` — check if `mu.Lock()` in `emit` is a bottleneck.

### 2.4 Optimization Candidate Matrix
After profiling, classify each candidate by impact vs. risk:

| Candidate                                  | Estimated Gain | Complexity | Contract Risk | Verdict    |
|--------------------------------------------|----------------|------------|---------------|------------|
| Pre-alloc events slice (heuristic cap)     | 5–15% allocs   | Low        | None          | Try        |
| Stream events via channel (no mutex)       | 0–10% latency  | Medium     | Low           | Bench first|
| Tune worker count (NumCPU vs NumCPU*2)     | 0–20% I/O      | Trivial    | None          | Bench      |
| Replace `filepath.Join` with concat        | 1–3%           | Low        | Low           | Skip unless proven |
| Batch `de.Info()` calls                    | 0% (cached)    | None       | None          | Verify     |
| Increase queue buffer (65536 → 131072)     | 0–5% on large  | Trivial    | None          | Bench      |
| `sync.Pool` for event structs              | 5–15% allocs   | Medium     | Medium        | Bench first|

## 3. Acceptance Criteria (BDD Scenarios)

**Feature**: Benchmark suite exists and passes

**Scenario**: Go benchmark suite runs without error [`contract`]
- **Given** `go test -bench=. -benchtime=3s ./internal/mss/adapters/fs/` is run
- **Then** it exits 0 and emits ns/op + allocs/op metrics for at least the baseline and `filepath.WalkDir` comparison

**Scenario**: Baseline is faster than `filepath.Walk` sequential [`performance`]
- **Given** a tree of ≥50k entries
- **When** C0 (worker pool) and C2 (`filepath.Walk`) are benchmarked
- **Then** C0 wall-time is strictly less than C2

**Scenario**: Benchmark results are committed as evidence [`contract`]
- **Given** the benchmark suite runs on the dev machine
- **Then** `spec/evidence/benchmark-baseline.txt` contains the captured output
- **And** `spec/mac-storage-scout.macos-performance.reference.md` is updated with measured numbers

**Scenario**: Profiling summary is captured [`contract`]
- **Given** `go test -bench . -cpuprofile` is run
- **Then** top-20 CPU symbols are recorded in `spec/evidence/pprof-cpu-profile.txt`
- **And** the dominant hotspot is identified and documented

**Scenario**: Optimization decisions are recorded [`contract`]
- **Given** profiling and benchmarking are complete
- **Then** `spec/mac-storage-scout.macos-performance.reference.md` contains a "Measured Results" section with:
  - Which implementation won on each tree size
  - Which optimization candidates were ruled out and why
  - Which optimizations (if any) warrant a follow-up task

## 4. Benchmark Implementation Guide

### 4.1 Bench File Structure (`mss_walker_bench_test.go`)
```go
package fs

import (
    "context"
    "os"
    "path/filepath"
    "testing"
    // ...
)

// BenchmarkWalkerCurrent — baseline: current MssGoFsWalkerAdapter
func BenchmarkWalkerCurrent(b *testing.B) { ... }

// BenchmarkWalkerFilepathWalkDir — comparison: stdlib filepath.WalkDir
func BenchmarkWalkerFilepathWalkDir(b *testing.B) { ... }

// BenchmarkWalkerFilepathWalk — comparison: older stdlib filepath.Walk
func BenchmarkWalkerFilepathWalk(b *testing.B) { ... }

// BenchmarkWalkerWorkersCPU — tuning: workers = NumCPU
func BenchmarkWalkerWorkersCPU(b *testing.B) { ... }

// BenchmarkWalkerWorkersCPUx4 — tuning: workers = NumCPU*4
func BenchmarkWalkerWorkersCPUx4(b *testing.B) { ... }
```

Benchmark trees: use `os.Getenv("MSS_BENCH_ROOT")` with a fallback to `os.TempDir()` so benchmarks work in CI without a large tree but can be pointed at a real tree on dev machines.

### 4.2 Evidence Capture Commands
```bash
# Baseline benchmark (3s per bench, large tree)
MSS_BENCH_ROOT=/Users/<user>/Library \
  go test -bench=. -benchtime=3s -benchmem \
  ./internal/mss/adapters/fs/ \
  | tee spec/evidence/benchmark-baseline.txt

# CPU profile
MSS_BENCH_ROOT=/Users/<user>/Library \
  go test -bench=BenchmarkWalkerCurrent -benchtime=10s \
  -cpuprofile=spec/evidence/cpu.prof \
  ./internal/mss/adapters/fs/

# Inspect profile
go tool pprof -top -cum spec/evidence/cpu.prof \
  | head -30 > spec/evidence/pprof-cpu-profile.txt

# Mutex contention
MSS_BENCH_ROOT=/Users/<user>/Library \
  go test -bench=BenchmarkWalkerCurrent -benchtime=10s \
  -mutexprofile=spec/evidence/mutex.prof \
  ./internal/mss/adapters/fs/
go tool pprof -top spec/evidence/mutex.prof | head -20
```

### 4.3 Performance Reference Update Template
After capturing results, update `spec/mac-storage-scout.macos-performance.reference.md` with:
```
## Measured Results (TSK-09, <date>)
### Machine: <model>, <cores>c/<threads>t, macOS <version>, APFS
### Tree: <path>, <N> files, <N> dirs

| Strategy         | Wall Time | ns/entry | allocs/op | B/op  |
|------------------|-----------|----------|-----------|-------|
| C0 worker pool   | Xs        | Yns      | Z         | W     |
| C1 WalkDir       | Xs        | Yns      | Z         | W     |
| C2 Walk          | Xs        | Yns      | Z         | W     |
| ...              |           |          |           |       |

### Hotspot (pprof top-5)
1. ...

### Optimization Decisions
- <candidate>: <verdict> — <reason>
- Follow-up: TSK-10 (if any wins were found)
```

## 5. Verification Strategy
- **Test Levels**: `benchmark`, `contract` (evidence files present), `performance` (C0 beats C2).
- **Verification Commands**:
  ```bash
  go test -bench=. -benchtime=1s -benchmem ./internal/mss/adapters/fs/
  go test ./...
  go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout
  go run ./cmd/mss-contract-lint --root . --mode short --format text
  ls spec/evidence/benchmark-baseline.txt spec/evidence/pprof-cpu-profile.txt
  grep -q 'Measured Results' spec/mac-storage-scout.macos-performance.reference.md
  ```
- **Completion Rule**:
  - Benchmark suite compiles and runs without error.
  - `spec/evidence/benchmark-baseline.txt` exists with ≥3 strategy rows.
  - `spec/mac-storage-scout.macos-performance.reference.md` has `Measured Results` section.
  - Optimization candidates are classified with a verdict.

## 6. Execution Log (AI-to-AI)
## EXEC_LOG_META
- task_id: TSK-09
- status: done
- canonical_log: self (this task file)

## EXEC_TIMELINE
- 2026-05-01T17:00:00Z: started from fresh `master` per MASTER_SYNC_PROTOCOL; created branch `ai/tsk-09-perf-benchmarks`.
- 2026-05-01T17:05:00Z: committed governance changes — AGENTS.md checkpoint 0 (branch context check), CLAUDE.md bootstrap step 1, task-09 spec.
- 2026-05-01T17:10:00Z: wrote `mss_walker_bench_test.go` — 6 benchmark functions covering baseline + worker tuning + stdlib comparisons; MSS_BENCH_ROOT env override for real trees.
- 2026-05-01T17:15:00Z: added DAG (section 9) and DAG update policy to spec/mac-storage-scout.spec.md; added DOCUMENTATION_SYNC_RULES to AGENTS.md requiring DAG + Decision Summary updates on new tasks; added D-009.
- 2026-05-01T17:25:00Z: ran benchmarks on os.TempDir (small tree, 5s/bench); all 6 strategies pass; WalkerCurrent 1.4x faster than WalkDir, 2x faster than Walk.
- 2026-05-01T17:40:00Z: ran benchmarks on ~/Developer (476k entries, 33GB, 1x pass); WalkerCurrent 16.0s vs WalkDir 29.2s (+83% slower) vs Walk 69.2s (+333% slower); NumCPU×2 optimal (NumCPU 33% slower on real tree).
- 2026-05-01T17:45:00Z: captured CPU profile (pprof); hotspot: 89.85% syscall.syscall, 74.65% os.Lstat — engine is fully syscall-bound; no Go-level hotspot exists.
- 2026-05-01T17:50:00Z: verified runtime correctness — `go test ./...` PASS, `go build` PASS, `mss-contract-lint` violations=0; real scan on macos-core profile confirmed output format intact.
- 2026-05-01T17:55:00Z: saved evidence: benchmark-baseline.txt, pprof-cpu-profile.txt, runtime-scan-proof.txt; updated spec/mac-storage-scout.macos-performance.reference.md with Measured Results section.
