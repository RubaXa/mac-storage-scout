# Task: [TSK-06] - CLI Rename To `mac-storage-scout` + Universal Skill Docs

## 1. Meta & Traceability
- **Purpose**: Eliminate global binary name collisions by renaming runtime command from `mss` to `mac-storage-scout`; document universal skill usage for agent systems.
- **Dependencies**: TSK-05
- **Supporting Artifacts**:
  - [mac-storage-scout.spec.md](../mac-storage-scout.spec.md)
  - [README.md](../../README.md)
  - [Session Handoff](../SESSION-HANDOFF.md)
- **Runtime Fidelity**: `runtime-hook-required`
- **Deferred Runtime Scope**: Release-packaged installer remains future work; source-build flow is canonical for now.
- **Target Files**:
  - `cmd/mac-storage-scout/main.go` (Update)
  - `internal/mss/adapters/report/mss_tree_text_report_adapter.go` (Update)
  - `README.md` (Update)
  - `AGENTS.md` (Update)
  - `CLAUDE.md` (Update)
  - `REVIEW.md` (Update)
  - `spec/SESSION-HANDOFF.md` (Update)
  - `spec/ascii-ux-research.md` (Update)
  - `spec/mac-storage-scout.spec.md` (Update)
  - `.github/PULL_REQUEST_TEMPLATE.md` (Update)
  - `.github/MERGE_CHECKLIST.md` (Update)
  - `.agent-skill/SKILL.md` (Create)
- **Target Test Files**:
  - `internal/mss/adapters/report/mss_tree_text_report_adapter_test.go` (Update)

## 2. Acceptance Criteria (BDD Scenarios)
**Feature**: Unique runtime command naming

**Scenario**: CLI usage outputs canonical command name
- **Given** user runs command with invalid args
- **When** usage text is printed
- **Then** it references `mac-storage-scout scan` and `mac-storage-scout delete`

**Scenario**: Agent skill docs are universal
- **Given** Codex/Claude/OpenCode style shell runtime
- **When** agent reads `.agent-skill/SKILL.md`
- **Then** it can build and run `mac-storage-scout` without ecosystem-specific adapters

**Scenario**: Skill is path-agnostic and self-consistent
- **Given** an agent that cloned the repo to an arbitrary path
- **When** it follows SKILL.md linearly
- **Then** every step (build → verify → dry-run → real delete) succeeds without local-path edits

**Scenario**: README opens with a depersonalized hero example
- **Given** a new reader landing on `README.md`
- **When** they scroll past the title
- **Then** the first content block is a styled scan example with no personal identifiers

## 3. Verification Strategy
- **Test Levels**: `unit`, `integration`, `runtime-smoke`
- **Verification Commands**:
  - `go test ./...`
  - `go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout`
  - `./bin/mac-storage-scout scan --profile macos-core --threshold 500MB --top 5`
- **Completion Rule**:
  - Tests and build pass.
  - Scan runs with renamed binary.

## 4. Execution Log (AI-to-AI)
## EXEC_LOG_META
- task_id: TSK-06
- status: done
- canonical_log: self (this task file)

## EXEC_TIMELINE
- 2026-05-01T15:53:15Z: renamed CLI command references and binary paths to `mac-storage-scout`.
- 2026-05-01T15:53:15Z: moved entrypoint directory to `cmd/mac-storage-scout`.
- 2026-05-01T15:53:15Z: added universal skill doc at `.agent-skill/SKILL.md`.
- 2026-05-01T15:53:15Z: synced README, AGENTS, SESSION-HANDOFF, and root spec references.
- 2026-05-01T18:30:00Z: cleaned residual `mss` references in `.github/PULL_REQUEST_TEMPLATE.md`, `.github/MERGE_CHECKLIST.md`, `spec/ascii-ux-research.md`.
- 2026-05-01T18:35:00Z: rewrote `README.md` to lead with a single depersonalized hero example; removed personal `/Users/k.lebedev/...` paths from build/test section.
- 2026-05-01T18:40:00Z: hardened `.agent-skill/SKILL.md` — removed personal install path, added `Invocation Convention` section, unified all examples to `./bin/mac-storage-scout` form, added depersonalized worked example.
- 2026-05-01T18:42:00Z: depersonalized `Canonical Paths` and `Standard Commands` blocks in `spec/SESSION-HANDOFF.md`.
- 2026-05-01T18:45:00Z: registered SKILL.md in `AGENTS.md::AUTHORITATIVE_READ_ORDER` and `CLAUDE.md::REQUIRED_BOOTSTRAP`.
- 2026-05-01T18:50:00Z: runtime-smoke verified — followed SKILL.md end-to-end (build → verify scan → dry-run → protected-path refusal → real delete on synthetic dir); all six steps PASS.

## EXEC_POINTER
For cross-task summary use `spec/mac-storage-scout.spec.md` section `Decision Summary (Root-Level)`.
