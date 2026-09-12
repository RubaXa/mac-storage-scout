# Task: [TSK-11] - Smart Metadata-First Anomaly Triage

## 1. Meta & Traceability
- **Purpose**: Detect generalized disk-growth anomalies quickly, then automatically measure the suspicious subtrees instead of requiring repeated ad-hoc shell investigations.
- **Dependencies**: TSK-10 (repeatable incident triage).
- **Runtime Fidelity**: live macOS metadata preflight plus targeted allocated-size traversal.
- **Target Files**: triage domain/ports/orchestrator/report, metadata anomaly adapter, CLI wiring, delete result accounting, tests, operator docs, runtime evidence.

## 2. Contract
- Default `triage` runs a bounded metadata-first preflight over both volatile roots and high-value broad roots such as Applications, Application Support, projects, caches, downloads, and temporary storage.
- The preflight reads filesystem metadata only, never file contents, never follows symlinks, and stops at explicit time, directory, depth, and per-directory entry budgets.
- Detection rules are product-independent:
  - extreme direct-entry or subdirectory fanout;
  - generated queue-like directories with substantial fanout;
  - repeated version-like sibling directories, including optional current/latest symlink evidence;
  - stale dense directories based on a bounded metadata sample.
- Every anomaly includes evidence, severity, and a concrete next-step hint. Findings are deterministically ranked.
- High-confidence anomaly paths are automatically measured in a separate targeted pass. Deep paths under a normal root are rescanned only when two-level base aggregation cannot expose their exact size; weak directory-heavy stale signals remain report-only to protect latency.
- CLI progress surfaces anomaly findings before targeted recursive measurement starts, so a slow subtree does not leave the operator without evidence.
- The final triage report includes preflight coverage/truncation and measured size/age when targeted measurement completes.
- Detection does not classify a path as delete-safe; existing process and safety evidence remains authoritative.
- Real deletion exits non-zero when any target is skipped or fails, and its reclaimed total excludes bytes that remain after a failed removal.

## 3. Acceptance Criteria
- Synthetic tests prove detection of extreme fanout without exhausting a directory, version accumulation without product names, stale generated queues, deterministic ranking, traversal budgets, and symlink non-following.
- Orchestrator tests prove anomaly paths are added to targeted collection without changing baseline scope.
- Report tests prove anomalies and truncated-preflight warnings are visible before ordinary hotspots.
- Delete tests prove failed targets are excluded from successful totals and produce a failing command result.
- Runtime evidence demonstrates that the preflight rediscovers live Crashpad-style churn or another live structural anomaly without app-specific rules.
- Full verification (`go test`, uncached tests, race tests, build, contract lint, and diff check) passes.

## 4. Execution Log (AI-to-AI)
## EXEC_LOG_META
- task_id: TSK-11
- status: in_pr
- canonical_log: self (this task file)

## EXEC_TIMELINE
- 2026-09-12T18:28:05Z: started from updated `master` on `ai/smart-anomaly-triage`; preserved unrelated untracked `fs.test`.
- 2026-09-12T18:28:05Z: baseline tests, build, and contract lint passed using an isolated Go build cache after the sandbox denied the user-level cache.
- 2026-09-12T18:28:05Z: converted live incident lessons into generalized metadata-first fanout, generated-queue, version-accumulation, giant-file, bounded-drilldown, early-progress, and truthful-delete contracts.
- 2026-09-12T18:28:05Z: live five-second preflight inspected 8,687 directories, streamed structural findings immediately, and automatically measured 21.9GB of mostly stale Downloads plus 3.70GB of application logs; proof recorded under `spec/evidence/task-11-smart-anomaly-triage-proof.txt`.
- 2026-09-12T18:37:06Z: a second live run converted a 105-second directory-heavy Containers drill-down into a report-only hint, completed recursive work in about 17 seconds, rediscovered 23 Yandex version directories generically, and surfaced multiple deep application cache fanout paths with exact targeted measurements.
- 2026-09-12T18:40:38Z: delete dry-run/yes smoke confirmed 1.00MB candidate and 1.00MB reclaimed with `deleted=1 failed=0 skipped=0`; injected unit failures proved failed and partially removed targets cannot inflate the successful total.
- 2026-09-12T18:40:38Z: `go test ./...`, `go test -count=1 ./...`, `go test -race ./...`, binary build, contract lint (`157` entities, `violations=0`), runtime profile scan, and `git diff --check` passed.
- 2026-09-12T18:44:00Z: rebased on current `origin/master`, pushed `ai/smart-anomaly-triage`, and opened PR #11 into `master` with runtime and delete-safety evidence.
