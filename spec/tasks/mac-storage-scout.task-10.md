# Task: [TSK-10] - Whole-Volume Audit and Repeatable Incident Triage

## 1. Meta & Traceability
- **Purpose**: Explain the complete writable macOS volume balance and make recurring high-churn disk incidents comparable without rebuilding an ad-hoc shell investigation.
- **Dependencies**: TSK-02 (walker), TSK-05 (orchestration), TSK-09 (performance baseline).
- **Runtime Fidelity**: full-volume macOS/APFS runtime audit.
- **Target Files**: CLI entrypoint, statfs adapter, audit and triage orchestrators/reports, triage collector/process/state adapters, one-filesystem walker guard, tests, operator docs, runtime evidence.

## 2. Contract
- `audit` defaults to `/System/Volumes/Data` when present and falls back to `/`.
- Audit scans allocated bytes and never descends into another mounted filesystem.
- Volume capacity, occupancy, and availability come from `statfs`, sampled before and after traversal.
- Output explicitly balances readable file allocation against ending occupancy.
- A positive remainder is `unaccounted`; a negative remainder is represented as `allocated-size-overcount`.
- Narrow custom roots are identified separately from their containing volume and produce a scope warning.
- The detailed tree preserves the canonical threshold and `other/top-N/rest/types` contract.
- `triage` defaults to generic high-churn package/agent/cache/temp/VM/update roots; `--broad` adds application data, containers, projects, and downloads.
- Triage collects allocated size at root plus two descendant levels with bounded root-level parallelism.
- The first saved run creates a versioned JSON baseline; later runs report volume growth and per-path growth/shrinkage.
- Each hotspot reports bytes modified within 24 hours, within seven days, and older than seven days; text output surfaces `old>7d`.
- Best-effort `lsof` attribution marks paths `active`; probe failure can never produce a `safe` result.
- Generic stale cache/log/temp paths may be `safe`; worktrees/snapshots/sessions/node_modules require `review`; everything else is `inspect`.
- Triage never deletes; all cleanup still requires the existing dry-run/yes contract.

## 3. Acceptance Criteria
- Unit tests cover statfs conversion, reconciliation, APFS-style overcount, rendering, and one-filesystem traversal.
- Invalid flags and audit failures return non-zero exit status.
- Full verification (`go test`, build, contract lint) passes.
- A real `/System/Volumes/Data` audit is captured under `spec/evidence/` and reports the current balance.
- Unit tests cover triage age aggregation, atomic state round-trip, baseline deltas, disappeared paths, activity safety, root overlap removal, and report output.
- Runtime evidence proves targeted triage and fast default triage on a live macOS filesystem.

## 4. Execution Log (AI-to-AI)
## EXEC_LOG_META
- task_id: TSK-10
- status: done
- canonical_log: self (this task file)

## EXEC_TIMELINE
- 2026-09-04T12:39:14Z: started from updated `master`; created branch `ai/volume-audit` after confirming the prior deadlock fix had already merged.
- 2026-09-04T12:39:14Z: implemented statfs volume snapshots, allocated-size reconciliation, explicit unaccounted/overcount/growth fields, and one-filesystem traversal.
- 2026-09-04T12:39:14Z: added outcome-first audit reporting and scope warnings for partial roots; targeted Go tests passed.
- 2026-09-04T12:47:00Z: full Data-volume audit reconciled 450GB occupied against 395GB readable allocation, exposed 55.4GB unaccounted, 588 access errors, and 4.03GB growth during the scan.
- 2026-09-04T12:47:00Z: targeted follow-up attributed active growth to `/private/tmp` flow-eval fixtures created by a live `gennady` Claude session; saved proof in `spec/evidence/task-10-volume-audit-proof.txt`.
- 2026-09-04T12:47:00Z: full Go tests, binary build, and contract lint passed (`92` entities, `0` violations).
- 2026-09-04T12:49:00Z: rebased on current `origin/master`, pushed `ai/volume-audit`, and opened PR #10.
- 2026-09-10T15:20:00Z: user approved continuing the existing branch and expanding TSK-10 after live incidents showed that audit still required manual age/process/delta investigation.
- 2026-09-10T15:33:00Z: implemented triage domain/ports, allocated-size age collector, atomic JSON baseline, lsof process evidence, conservative safety policy, and outcome-first report.
- 2026-09-10T15:40:00Z: optimized default triage from an aborted 93.58s sequential broad scan to 12.03s using bounded root parallelism and fast-vs-broad scope separation.
- 2026-09-10T15:46:00Z: live default triage completed in 19.01s with lsof evidence; identified inactive npm cache as safe while preserving active OpenCode, Codex, Library cache, and tmp paths.
- 2026-09-10T15:50:00Z: added non-overlapping candidate policy, bounded process-owner rendering, atomic-state and safety tests; `go test -count=1 ./...`, `go test -race ./...`, build, contract-lint (`violations=0`), and `git diff --check` passed.
- 2026-09-12T18:28:05Z: PR #10 was confirmed merged into `master`; task status synchronized to done before TSK-11 work began.
