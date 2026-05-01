# Task: [TSK-07] - CLI Rename To `mac-storage-scout` + Universal Agent Skill

## 1. Meta & Traceability
- **Purpose**: Eliminate global binary name collisions by renaming runtime command from `mss` to `mac-storage-scout`; document a universal, path-agnostic agent skill so any shell-capable runtime (Codex / Claude / OpenCode) can build, scan, and safely delete without ecosystem-specific adapters.
- **Dependencies**: TSK-05, TSK-06
- **Supporting Artifacts**:
  - [../mac-storage-scout.spec.md](../mac-storage-scout.spec.md)
  - [../../README.md](../../README.md)
  - [../../AGENTS.md](../../AGENTS.md)
  - [../../CLAUDE.md](../../CLAUDE.md)
  - [../SESSION-HANDOFF.md](../SESSION-HANDOFF.md)
- **Runtime Fidelity**: `runtime-hook-required`
- **Deferred Runtime Scope**: Release-packaged installer remains future work; source-build flow is canonical for now.
- **Target Files**:
  - `cmd/mac-storage-scout/main.go` (Move from `cmd/mss/main.go` + adjust strings + adjust `@consumer` tags)
  - `cmd/mac-storage-scout/main_test.go` (Move from `cmd/mss/main_test.go`)
  - `internal/mss/adapters/report/mss_tree_text_report_adapter.go` (Update report header strings + `@consumer` tags)
  - `internal/mss/adapters/report/mss_tree_text_report_adapter_test.go` (Update header expectations)
  - `README.md` (Update — lead with depersonalized hero example)
  - `AGENTS.md` (Update — register skill in read-order, refresh bin paths)
  - `CLAUDE.md` (Update — register skill in bootstrap, depersonalize)
  - `REVIEW.md` (Update — refresh bin paths)
  - `spec/SESSION-HANDOFF.md` (Update — depersonalize Canonical Paths and Standard Commands)
  - `spec/ascii-ux-research.md` (Update — rename `mss` heading)
  - `spec/mac-storage-scout.spec.md` (Update — record D-005, refresh paths)
  - `.github/PULL_REQUEST_TEMPLATE.md` (Update — refresh bin path)
  - `.github/MERGE_CHECKLIST.md` (Update — refresh bin path)
  - `.agent-skill/SKILL.md` (Create)
- **Traceability**:
  - **Decisions**: D-005 (canonical command name; no default `mss` alias) — see Decision Summary in root spec.
  - **Contracts**: [Delete Safety](../../AGENTS.md#hard_contracts) — preserved unchanged.

## 2. Acceptance Criteria (BDD Scenarios)
**Feature**: Unique runtime command naming + universal agent skill

**Scenario**: CLI usage outputs canonical command name [`unit`]
- **Given** user runs the binary with invalid args
- **When** usage text is printed
- **Then** it references `mac-storage-scout scan` and `mac-storage-scout delete`

**Scenario**: Agent skill docs are universal [`contract`]
- **Given** Codex/Claude/OpenCode style shell runtime
- **When** an agent reads `.agent-skill/SKILL.md`
- **Then** it can build and run `mac-storage-scout` without ecosystem-specific adapters

**Scenario**: Skill is path-agnostic and self-consistent [`runtime-smoke`]
- **Given** an agent that cloned the repo to an arbitrary path
- **When** it follows SKILL.md linearly
- **Then** every step (build → verify → dry-run → protected-path refusal → real delete) succeeds without local-path edits

**Scenario**: README opens with a depersonalized hero example [`contract`]
- **Given** a new reader landing on `README.md`
- **When** they scroll past the title
- **Then** the first content block is a styled scan example with no personal identifiers

## 3. Verification Strategy
- **Test Levels**: `unit`, `integration`, `runtime-smoke`, `contract` (TSK-06 contract-lint gate must pass)
- **Verification Commands**:
  - `go test ./...`
  - `go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout`
  - `go run ./cmd/mss-contract-lint --root . --mode short --format text`
  - `./bin/mac-storage-scout scan --profile macos-core --threshold 500MB --top 5 --no-progress`
  - end-to-end skill smoke (build → scan → dry-run → `/System` refusal → real delete on synthetic dir)
- **Completion Rule**:
  - Tests, build, and contract-lint pass.
  - Skill workflow runs end-to-end with the renamed binary.

## 4. Execution Log (AI-to-AI)
## EXEC_LOG_META
- task_id: TSK-07
- status: done
- canonical_log: self (this task file)
- evidence_refs:
  - ../evidence/skill-runtime-smoke.txt

## EXEC_TIMELINE
- 2026-05-01T15:53:15Z: renamed CLI command references and binary paths to `mac-storage-scout`; moved entrypoint to `cmd/mac-storage-scout/`; added universal skill doc at `.agent-skill/SKILL.md`; synced README, AGENTS, SESSION-HANDOFF, root spec.
- 2026-05-01T18:30:00Z: cleaned residual `mss` references in `.github/PULL_REQUEST_TEMPLATE.md`, `.github/MERGE_CHECKLIST.md`, `spec/ascii-ux-research.md`.
- 2026-05-01T18:35:00Z: rewrote `README.md` to lead with a single depersonalized hero example; removed `/Users/k.lebedev/...` paths from build/test sections.
- 2026-05-01T18:40:00Z: hardened `.agent-skill/SKILL.md` — removed personal install path, added `Invocation Convention`, unified examples to `./bin/mac-storage-scout` form, added depersonalized worked example.
- 2026-05-01T18:42:00Z: depersonalized Canonical Paths and Standard Commands in `spec/SESSION-HANDOFF.md`.
- 2026-05-01T18:45:00Z: registered SKILL.md in `AGENTS.md::AUTHORITATIVE_READ_ORDER` and `CLAUDE.md::REQUIRED_BOOTSTRAP`.
- 2026-05-01T18:50:00Z: runtime-smoke verified — followed SKILL.md end-to-end (build → verify scan → dry-run → protected-path refusal → real delete on synthetic dir); all six steps PASS. Evidence: `spec/evidence/skill-runtime-smoke.txt`.
- 2026-05-01T19:30:00Z: rebased onto master after TSK-06 (Go DevGen/QA) merged; renumbered own task TSK-06 → TSK-07 to resolve ID collision; merged contract-lint tags from TSK-06 into renamed `cmd/mac-storage-scout/main.go` and report adapter; updated `@consumer` refs from `cmd/mss/main.go` to `cmd/mac-storage-scout/main.go`; verified contract-lint clean (`scanned=N, violations=0`).

## EXEC_POINTER
For cross-task summary use `spec/mac-storage-scout.spec.md` section `Decision Summary (Root-Level)`.
