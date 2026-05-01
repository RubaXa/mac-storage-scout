# AGENTS.md

## DOC_META
- doc_id: AGENTS
- doc_type: ai_to_ai_operating_contract
- version: 1.2.0
- status: active
- repo: mac-storage-scout
- updated_utc: 2026-05-01T15:00:00Z

## EXECUTION_INTENT
Enable deterministic autonomous work on this repository with minimal ambiguity.

## AUTHORITATIVE_READ_ORDER
1. .ai/rules/go-devgen.contracts.md
2. .ai/rules/go-qa.testing.md
3. .ai/research/go-practices-evidence.md
4. README.md
5. .agent-skill/SKILL.md (universal CLI skill — build/scan/delete contract)
6. spec/mac-storage-scout.spec.md
7. spec/mac-storage-scout.output-format.source-of-truth.md
8. spec/SESSION-HANDOFF.md
9. spec/tasks/mac-storage-scout.task-*.md (Execution Log sections)

## AI_RULESET_LOCATION
- root: `.ai/`
- mandatory_rules:
  - `.ai/rules/go-devgen.contracts.md`
  - `.ai/rules/go-qa.testing.md`
- reference_research:
  - `.ai/research/go-practices-evidence.md`

## PROJECT_SETUP
- run once per clone:
  - `./scripts/setup-githooks.sh`

## HARD_CONTRACTS
- output.large_items:
  - rule: print every item where size >= threshold as explicit node
- output.small_items:
  - rule: include every item where size < threshold only inside `other`
- output.other_shape:
  - required_blocks:
    - top-N
    - rest
    - types
  - invariance: same structure in every section
- deletion.safety:
  - must_run_dry_run_before_real_delete: true
  - real_delete_requires_yes_flag: true
  - protected_paths:
    - /
    - /System
    - /usr
    - /bin
    - /sbin
    - /private/var/vm

## OPERATIONAL_WORKFLOW
1. load `.ai` rule set and affected specs
2. run verification baseline
3. implement scoped change
4. run verification again
5. update docs/specs if behavior changed
6. append Execution Log in the affected task spec file(s)
7. create branch `ai/<name>`
8. commit + push branch
9. open pull request into `master`

## REQUIRED_COMMANDS
Run from the project root (path agnostic):
```bash
go test ./...
go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout
go run ./cmd/mss-contract-lint --root . --mode short --format text
./bin/mac-storage-scout scan --profile macos-core --threshold 500MB --top 5
./bin/mac-storage-scout delete --dry-run <path> [path...]
./bin/mac-storage-scout delete --yes <path> [path...]
```

## COMMIT_GATES
- pre-commit hook path: `.githooks/pre-commit` (required)
- local git config must set: `core.hooksPath=.githooks`
- commit must be blocked if `go run ./cmd/mss-contract-lint --root . --mode short --format text` returns non-zero

## DOCUMENTATION_SYNC_RULES
When behavior/UX/CLI changes:
- update README.md
- update matching spec file(s)
- append/update Execution Log in affected `spec/tasks/*.md` with UTC timestamp

When a new task spec is created (`spec/tasks/mac-storage-scout.task-NN.md`):
- add the task row to `spec/README.md` (Execution Order DAG table) — this is the **living status tracker**; update it also when task status changes (pending → in PR → done)
- add the task to the Execution Order DAG in `spec/mac-storage-scout.spec.md` (section 9) before opening a PR
- add a D-NNN entry to the Decision Summary in `spec/mac-storage-scout.spec.md` (section 8) summarizing the key decision this task encodes

## GIT_FLOW_POLICY
- direct commits to `master`: forbidden
- required branch naming: `ai/<name>`
- required merge path: PR from `ai/<name>` into `master`
- PR body source: `.github/PULL_REQUEST_TEMPLATE.md` (required)
- merge gate checklist: `.github/MERGE_CHECKLIST.md` (required)
- reviewers/owners policy: `.github/CODEOWNERS`

### MASTER_SYNC_PROTOCOL (mandatory)
The agent MUST sync with `origin/master` at three checkpoints. Skipping any of them is a process violation.

0. **Branch context check** — before any work, run `git branch --show-current` and inspect the result:
   - If the result is `master`: proceed to checkpoint 1 below.
   - If the result is any other branch: **stop and ask the user**:
     > "You are currently on branch `<name>`, not `master`. Recommended: switch to `master` and start fresh. Do you want to (a) switch to `master` first, or (b) continue on this branch?"
   - Only proceed without switching if the user explicitly confirms option (b).
   - Rationale: silently starting or continuing work on a stale feature branch risks divergence, id collisions, and wasted effort.

1. **Before starting any task** — fetch and fast-forward `master`, then branch off:
   ```bash
   git checkout master
   git fetch origin
   git pull --ff-only origin master
   git checkout -b ai/<name>
   ```
   Rationale: starting from a stale base guarantees a future rebase conflict and may duplicate work that already landed (e.g. another agent claiming the next free `TSK-NN` id).

2. **Before opening a PR** — re-fetch and rebase the working branch onto fresh `origin/master`:
   ```bash
   git fetch origin
   git rebase origin/master
   # resolve conflicts deterministically, re-run verification, then:
   git push --force-with-lease origin ai/<name>
   ```
   Always `--force-with-lease`, never plain `--force`. If conflicts arise, follow `MERGE_CONFLICT_PROTOCOL` below.

3. **After the PR is merged** — return to `master` and update locally before offering further work:
   ```bash
   git checkout master
   git pull --ff-only origin master
   ```
   The agent MUST proactively offer this switch to the user once the PR shows `merged: true`. Do not start the next task on a leftover feature branch.

### MERGE_CONFLICT_PROTOCOL
When `git rebase origin/master` reports conflicts:
- Resolve each file deterministically — prefer master for cross-cutting policy/contract changes (e.g. contract-lint tags), apply branch-local deltas surgically on top.
- If task ids collide (e.g. both branches created the same `TSK-NN`), the branch that landed first keeps the id; the rebasing branch renumbers to the next free id and updates every cross-reference (decision summary, EXEC_LOG, evidence refs, PR body).
- Re-run the full verification baseline after every resolution batch (`go test`, `go build`, `mss-contract-lint`).
- Never resolve conflicts by accepting one whole side blindly — review semantics file by file.

### PR_PRECONDITIONS
- working branch is rebased on the latest `origin/master` (no merge commits from `master` in branch history)
- `go test ./...` passes
- `go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout` passes
- `go run ./cmd/mss-contract-lint --root . --mode short --format text` reports `violations=0`
- PR description includes: scope summary, change map, verification commands and results, runtime-smoke evidence link when behavior changed

## DONE_CRITERIA
- build_pass: true
- tests_pass: true
- output_contract_preserved: true
- delete_safety_contract_preserved: true
- task_execution_log_updated: true
