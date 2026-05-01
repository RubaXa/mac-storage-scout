# Task: [TSK-06] - Go DevGen/QA Ruleset + Contract Hardening + Test Expansion

## 1. Meta & Traceability
- **Purpose**: Adapt external TS-oriented agent contracts to Go, codify autonomous coding/testing rules under `.ai`, and align repository code/tests with the resulting contract style.
- **Dependencies**: TSK-01, TSK-02, TSK-03, TSK-04, TSK-05
- **Supporting Artifacts**:
  - [../mac-storage-scout.spec.md](../mac-storage-scout.spec.md)
  - [../mac-storage-scout.output-format.source-of-truth.md](../mac-storage-scout.output-format.source-of-truth.md)
  - [../../AGENTS.md](../../AGENTS.md)
  - [../../.ai/rules/go-devgen.contracts.md](../../.ai/rules/go-devgen.contracts.md)
  - [../../.ai/rules/go-qa.testing.md](../../.ai/rules/go-qa.testing.md)
- **Runtime Fidelity**: `contract-and-runtime`
- **Deferred Runtime Scope**: none introduced.
- **Target Files**:
  - `.ai/README.md` (Create)
  - `.ai/rules/go-devgen.contracts.md` (Create)
  - `.ai/rules/go-qa.testing.md` (Create)
  - `.ai/research/go-practices-evidence.md` (Create)
  - `.ai/checklists/go-change-checklist.md` (Create)
  - `.ai/checklists/go-test-checklist.md` (Create)
  - `AGENTS.md` (Update)
  - `AGETNS.md` (Update)
  - `spec/README.md` (Update)
  - `spec/SESSION-HANDOFF.md` (Update)
  - `spec/mac-storage-scout.spec.md` (Update)
  - production/test files aligned with new contract comments and test quality rules
- **Traceability**:
  - **Contracts**: [Root Output Rules](../mac-storage-scout.output-format.source-of-truth.md#canonical-rules)
  - **Contracts**: [Delete Safety](../../AGENTS.md#hard_contracts)

## 2. Acceptance Criteria (BDD Scenarios)
**Feature**: Autonomous Go coding/testing governance

**Scenario**: Agent bootstrap loads `.ai` rules before implementation [`contract`]
- **Given** a new autonomous coding session
- **When** the session reads repo contracts
- **Then** `.ai/rules` and `.ai/checklists` are part of mandatory read order

**Scenario**: Production code is hardened with explicit contracts [`contract`]
- **Given** core runtime modules
- **When** they are updated
- **Then** non-trivial logic is documented with machine-readable intent/contracts
- **And** precondition errors are explicit and traceable

**Scenario**: Test suite covers contract boundaries more fully [`contract`]
- **Given** core adapters/orchestrator
- **When** tests are executed
- **Then** missing test surfaces are covered
- **And** failure diagnostics follow clear got/want and scenario naming rules

## 3. Verification Strategy
- **Test Levels**: `unit`, `integration`, `contract`
- **Verification Commands**:
  - `go test ./...`
  - `go build -o ./bin/mss ./cmd/mss`
  - `./bin/mss scan --profile macos-core --threshold 500MB --top 5`
  - `./bin/mss delete --dry-run <path>` and `./bin/mss delete --yes <path>` on safe temp fixture

## 4. Execution Log (AI-to-AI)
## EXEC_LOG_META
- task_id: TSK-06
- status: done
- canonical_log: self (this task file)
- evidence_refs:
  - ../evidence/task-06-go-rules-proof.txt

## EXEC_TIMELINE
- 2026-05-01T14:19:39Z: task initialized, branch `ai/go-devgen-qa-rules` active, baseline verification completed
- 2026-05-01T14:19:39Z: `.ai` rules/checklists/research files created and wired into agent/spec docs
- 2026-05-01T14:19:39Z: production contract hardening completed (`cmd`, `domain`, `app`, `adapters`, `ports`)
- 2026-05-01T14:19:39Z: test expansion completed (`fs`, `progress`, `app`, `integration`) and legacy tests aligned to QA rules
- 2026-05-01T14:19:39Z: verification `go test ./...` -> PASS
- 2026-05-01T14:19:39Z: verification `go test -count=1 ./...` -> PASS
- 2026-05-01T14:19:39Z: verification `go build -o ./bin/mss ./cmd/mss` -> PASS
- 2026-05-01T14:19:39Z: verification `./bin/mss scan --profile macos-core --threshold 500MB --top 5 --no-progress` -> PASS
- 2026-05-01T14:19:39Z: verification delete flow (`--dry-run` and `--yes` on temp fixture) -> PASS
- 2026-05-01T15:00:05Z: added explicit `@consumer` tags to contract roots (ports/adapters/orchestrator/domain) and updated rules/spec references
- 2026-05-01T15:00:05Z: verification `go test ./...` after `@consumer` pass -> PASS

## EXEC_POINTER
For cross-task summary use `spec/mac-storage-scout.spec.md` section `Decision Summary (Root-Level)`.
