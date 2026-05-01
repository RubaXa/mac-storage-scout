# Task: [TSK-08] - Master Sync Protocol + Post-Merge Return-To-Master Policy

## 1. Meta & Traceability
- **Purpose**: Codify a deterministic git-flow sync protocol so that every autonomous agent (a) starts from a fresh `master`, (b) rebases before opening a PR, and (c) is required to offer the user a return-to-master switch after merge. Eliminate the class of failure where parallel branches silently diverge until PR-time conflicts surface.
- **Dependencies**: TSK-07 (governance docs already in place: AGENTS.md, CLAUDE.md, MERGE_CHECKLIST.md, PR_TEMPLATE.md).
- **Supporting Artifacts**:
  - [../../AGENTS.md](../../AGENTS.md)
  - [../../CLAUDE.md](../../CLAUDE.md)
  - [../../.github/MERGE_CHECKLIST.md](../../.github/MERGE_CHECKLIST.md)
  - [../../.github/PULL_REQUEST_TEMPLATE.md](../../.github/PULL_REQUEST_TEMPLATE.md)
- **Runtime Fidelity**: `governance-only` (no runtime behavior change).
- **Deferred Runtime Scope**: none.
- **Target Files**:
  - `AGENTS.md` (Update — add `MASTER_SYNC_PROTOCOL`, `MERGE_CONFLICT_PROTOCOL`, `PR_PRECONDITIONS`)
  - `CLAUDE.md` (Update — bootstrap step #1 = sync master; behavior rules reflect three sync checkpoints; post-merge return-to-master is a behavior rule)
  - `.github/MERGE_CHECKLIST.md` (Update — `Branch is rebased on origin/master` as precondition; contract-lint as required validation; post-merge return-to-master block)
  - `.github/PULL_REQUEST_TEMPLATE.md` (Update — add rebase + contract-lint checkboxes to Verification)

## 2. Acceptance Criteria (BDD Scenarios)
**Feature**: Three-checkpoint master sync

**Scenario**: Agent starts a new task [`contract`]
- **Given** any new autonomous task
- **When** the agent begins
- **Then** the first action is `git checkout master && git fetch origin && git pull --ff-only origin master`
- **And** the working branch is created off the freshly synced `master`

**Scenario**: Agent prepares to open a PR [`contract`]
- **Given** a working branch with completed scope
- **When** the agent is about to open a PR
- **Then** it re-fetches and rebases onto `origin/master`
- **And** any conflicts are resolved per `MERGE_CONFLICT_PROTOCOL`
- **And** the push uses `--force-with-lease`, never plain `--force`

**Scenario**: PR is merged [`contract`]
- **Given** a PR has been merged into `master`
- **When** the agent reports completion
- **Then** the agent proactively offers the user to switch back to `master` and fast-forward
- **And** the agent does not start the next task on the leftover feature branch

**Scenario**: Task id collision during rebase [`contract`]
- **Given** two branches independently created the same `TSK-NN` id
- **When** the second branch rebases
- **Then** the branch that landed first keeps the id
- **And** the rebasing branch renumbers to the next free id and updates every cross-reference

## 3. Verification Strategy
- **Test Levels**: `contract` (governance docs read order, presence of required sections).
- **Verification Commands**:
  - `grep -q 'MASTER_SYNC_PROTOCOL' AGENTS.md`
  - `grep -q 'sync \`master\` first' CLAUDE.md`
  - `grep -q 'Branch is rebased on the latest' .github/MERGE_CHECKLIST.md`
  - `grep -q 'branch rebased on latest \`origin/master\`' .github/PULL_REQUEST_TEMPLATE.md`
  - `go test ./...` (no behavior change — sanity check that governance edits did not break anything)
  - `go run ./cmd/mss-contract-lint --root . --mode short --format text` (violations=0)
- **Completion Rule**:
  - All four governance documents reference the three-checkpoint sync protocol.
  - PR template and merge checklist enforce the rebase precondition and contract-lint gate.

## 4. Execution Log (AI-to-AI)
## EXEC_LOG_META
- task_id: TSK-08
- status: done
- canonical_log: self (this task file)

## EXEC_TIMELINE
- 2026-05-01T20:00:00Z: started fresh from `master` after TSK-07 merged (followed the very protocol this task codifies).
- 2026-05-01T20:05:00Z: added `MASTER_SYNC_PROTOCOL`, `MERGE_CONFLICT_PROTOCOL`, `PR_PRECONDITIONS` sections to `AGENTS.md::GIT_FLOW_POLICY`.
- 2026-05-01T20:08:00Z: updated `CLAUDE.md::REQUIRED_BOOTSTRAP` (step #1 = sync master) and `BEHAVIOR_RULES` (three checkpoints + proactive post-merge offer).
- 2026-05-01T20:10:00Z: updated `.github/MERGE_CHECKLIST.md` — rebase precondition, contract-lint validation, post-merge return-to-master block.
- 2026-05-01T20:12:00Z: updated `.github/PULL_REQUEST_TEMPLATE.md` — added rebase and contract-lint checkboxes to Verification.
- 2026-05-01T20:15:00Z: verification — `go test ./...` PASS; `mss-contract-lint` scanned=74 violations=0; required-grep checks PASS.

## EXEC_POINTER
For cross-task summary use `spec/mac-storage-scout.spec.md` section `Decision Summary (Root-Level)`.
