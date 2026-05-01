# REVIEW.md

## Mission
This is the AI-to-AI code review contract for `mac-storage-scout`.
The review objective is to keep the project fast, safe, and reliable on real macOS filesystems.

## Authoritative Inputs (Read Before Reviewing)
1. `README.md`
2. `AGENTS.md`
3. `spec/mac-storage-scout.spec.md`
4. `spec/mac-storage-scout.output-format.source-of-truth.md`
5. `spec/mac-storage-scout.macos-performance.reference.md`
6. `spec/SESSION-HANDOFF.md`
7. `spec/tasks/mac-storage-scout.task-*.md` (especially Acceptance Criteria + Execution Log)
8. Diff against `origin/master`

If a requirement in these sources conflicts with generic style guidance, these sources win.

## Review Outcome Contract
A review is considered complete only if it covers:
- correctness against output and delete contracts
- concurrency/goroutine safety
- filesystem safety and macOS/APFS semantics
- performance and memory risk
- test and verification sufficiency

## Spec-Derived Non-Negotiable Gates
Failing any gate is at least `P1` and often `P0`.

### Gate A: Output Contract Integrity
Must preserve all rules from `spec/mac-storage-scout.output-format.source-of-truth.md`:
- every item `>= threshold` is explicit
- every item `< threshold` appears only in `other`
- `other` shape is always:
  - `top-N`
  - `rest`
  - `types`
- ordering remains deterministic:
  - visible items by size desc, then name asc
  - `top-N` by size desc
  - `types` by size desc

### Gate B: Delete Safety Integrity
Must preserve all safety rules from `AGENTS.md` and spec decisions:
- real deletion requires `--yes`
- dry-run path remains functional
- protected paths remain blocked:
  - `/`
  - `/System`
  - `/usr`
  - `/bin`
  - `/sbin`
  - `/private/var/vm`
- no bypass via path normalization edge cases (`..`, repeated separators, relative aliases)

### Gate C: Runtime Resilience
Must preserve runtime behavior from TSK-02 and TSK-05:
- scan continues on expected non-fatal FS failures (`EPERM`, `EACCES`, `ENOENT`)
- non-fatal failures are surfaced in counters/summary
- traversal still terminates deterministically (no queue/workgroup deadlock)
- default behavior does not follow symlinks

### Gate D: Size Semantics Integrity
Must preserve:
- default mode = `logical` (`st_size`)
- optional mode `allocated` = `st_blocks * 512`
- no misleading presentation of allocated mode as exact physical ownership on APFS clones/snapshots

## Security Review Rules (Project-Specific)
Flag findings when the patch introduces:
- unsafe delete path expansion or weakened guard checks
- symlink-follow behavior in scan/delete paths without explicit contract update
- deletion of implicit broad paths due to bad normalization
- stderr/stdout ambiguity that can hide dangerous delete behavior from operators
- hidden behavior changes that bypass user confirmation workflow

## Concurrency and Goroutine Review Rules
For any touched concurrent code, verify:
- every goroutine has a deterministic shutdown path
- ticker/timer resources are always stopped
- channels are closed exactly once by correct owner
- no send-on-closed-channel or blocked send risk after cancellation
- `WaitGroup.Add`/`Done` accounting cannot drift under all branches
- context cancellation is honored in hot loops and enqueue paths
- progress rendering never blocks worker hot path
- shared state has race-safe access (`atomic`, mutex, ownership boundaries)

Typical bug patterns to catch:
- goroutine leak after command exit
- queue drain deadlock under cancellation/error paths
- unbounded retry/requeue loops
- lock contention around high-frequency event append/counter updates

## Performance and Memory Review Rules
The project value proposition is speed on large trees. Flag regressions when patch adds:
- recursive traversal replacing bounded worker-pool behavior
- unbounded memory growth (event buffering, queue growth, map growth) without guardrails
- extra `Stat/Info` calls in tight loops when one is sufficient
- blocking terminal I/O in filesystem worker goroutines
- expensive string/path transformations in hot loops without need

Explicit performance expectation:
- traversal remains bounded and completes on large live trees
- progress updates are periodic and lightweight
- non-fatal errors do not trigger global retry storms

## Filesystem Correctness Rules
When touching FS logic, verify:
- `Lstat` vs `Stat` usage matches symlink policy
- relative and `~` paths normalize correctly
- root path identity remains stable for aggregation/reporting
- racey disappear/recreate paths are tolerated without full scan abort
- permission failures are counted and review-visible

## Output/Formatting Correctness Rules
When touching aggregation/report code, verify:
- section total equals visible children + `other` total
- `other.Count`, `RestCount`, extension counts/sizes are coherent
- `top-N` behavior for `N<=0`, `N<available`, `N>available` is deterministic
- no `< threshold` leakage back into explicit siblings
- format consistency across all sections is preserved

## Test and Evidence Requirements for Review Approval
If behavior changed in runtime code, review should demand evidence:
- `go test ./...` passes
- `go build -o ./bin/mss ./cmd/mss` passes
- runtime scan sample validates output shape
- delete flow sample validates dry-run and guarded behavior (when delete logic touched)

Preferred evidence files:
- `spec/evidence/build-test-proof.txt`
- `spec/evidence/runtime-scan-proof.txt`
- `spec/evidence/profile-macos-core-proof.txt`
- `spec/evidence/size-mode-allocated-proof.txt`
- `spec/evidence/progress-tty-proof.txt`
- `spec/evidence/acceptance-matrix.md`

Missing or stale evidence is not always a code bug, but it is a release-readiness risk and should be called out.

## Diff-to-Risk Mapping (Where to Focus First)
- `cmd/mss/main.go`: CLI contract, safety gates, flag semantics
- `internal/mss/adapters/fs/*`: traversal safety, symlink policy, size semantics, error tolerance
- `internal/mss/app/mss_scan_orchestrator.go`: lifecycle correctness, queue/event/counter handling
- `internal/mss/adapters/aggregate/*`: threshold split and deterministic output semantics
- `internal/mss/adapters/report/*`: shape invariants and readability guarantees
- `internal/mss/adapters/progress/*`: goroutine/ticker/TTY behavior

## Priority Model
- `P0` (`priority=0`): release-blocking safety/correctness failure (data-loss or contract-break in normal use)
- `P1` (`priority=1`): urgent correctness/perf/concurrency issue likely in real workloads
- `P2` (`priority=2`): clear bug with narrower trigger conditions
- `P3` (`priority=3`): low-impact maintainability risk with concrete failure mode

Examples:
- `P0`: delete can execute without `--yes`; protected path can be deleted
- `P1`: deadlock/leak in walker/progress lifecycle
- `P1`: output violates threshold contract
- `P2`: unstable sort causing non-deterministic output order
- `P2`: incorrect non-fatal error accounting

## Review Procedure (Deterministic)
1. Compute merge base and inspect full diff against `origin/master`.
2. Map changed hunks to risk areas above.
3. Run non-negotiable gates (A-D) first.
4. Run concurrency + FS + performance checklist.
5. Verify tests/evidence match touched behavior.
6. Emit only discrete, actionable findings with precise file/line location.
7. If no findings, explicitly state why patch is correct against these gates.

## Reviewer Command Baseline
```bash
cd /Users/k.lebedev/Developer/mac-storage-scout
go test ./...
go build -o ./bin/mss ./cmd/mss
./bin/mss scan --profile macos-core --threshold 500MB --top 5 --no-progress
./bin/mss delete --dry-run <path> [path...]
```

If delete code changed, include explicit deny-case checks for protected paths.
