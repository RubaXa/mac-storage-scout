# Task: [TSK-10] - Whole-Volume Occupancy Reconciliation Audit

## 1. Meta & Traceability
- **Purpose**: Explain the complete writable macOS volume balance instead of presenting selected path scans as if they account for all occupied space.
- **Dependencies**: TSK-02 (walker), TSK-05 (orchestration), TSK-09 (performance baseline).
- **Runtime Fidelity**: full-volume macOS/APFS runtime audit.
- **Target Files**: CLI entrypoint, statfs adapter, audit orchestrator/report, one-filesystem walker guard, tests, operator docs, runtime evidence.

## 2. Contract
- `audit` defaults to `/System/Volumes/Data` when present and falls back to `/`.
- Audit scans allocated bytes and never descends into another mounted filesystem.
- Volume capacity, occupancy, and availability come from `statfs`, sampled before and after traversal.
- Output explicitly balances readable file allocation against ending occupancy.
- A positive remainder is `unaccounted`; a negative remainder is represented as `allocated-size-overcount`.
- Narrow custom roots are identified separately from their containing volume and produce a scope warning.
- The detailed tree preserves the canonical threshold and `other/top-N/rest/types` contract.

## 3. Acceptance Criteria
- Unit tests cover statfs conversion, reconciliation, APFS-style overcount, rendering, and one-filesystem traversal.
- Invalid flags and audit failures return non-zero exit status.
- Full verification (`go test`, build, contract lint) passes.
- A real `/System/Volumes/Data` audit is captured under `spec/evidence/` and reports the current balance.

## 4. Execution Log (AI-to-AI)
## EXEC_LOG_META
- task_id: TSK-10
- status: in_pr
- canonical_log: self (this task file)

## EXEC_TIMELINE
- 2026-09-04T12:39:14Z: started from updated `master`; created branch `ai/volume-audit` after confirming the prior deadlock fix had already merged.
- 2026-09-04T12:39:14Z: implemented statfs volume snapshots, allocated-size reconciliation, explicit unaccounted/overcount/growth fields, and one-filesystem traversal.
- 2026-09-04T12:39:14Z: added outcome-first audit reporting and scope warnings for partial roots; targeted Go tests passed.
- 2026-09-04T12:47:00Z: full Data-volume audit reconciled 450GB occupied against 395GB readable allocation, exposed 55.4GB unaccounted, 588 access errors, and 4.03GB growth during the scan.
- 2026-09-04T12:47:00Z: targeted follow-up attributed active growth to `/private/tmp` flow-eval fixtures created by a live `gennady` Claude session; saved proof in `spec/evidence/task-10-volume-audit-proof.txt`.
- 2026-09-04T12:47:00Z: full Go tests, binary build, and contract lint passed (`92` entities, `0` violations).
- 2026-09-04T12:49:00Z: rebased on current `origin/master`, pushed `ai/volume-audit`, and opened PR #10.
