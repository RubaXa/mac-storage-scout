# Task: [TSK-05] - Integration, Validation, and Acceptance Gate

## 1. Meta & Traceability
- **Purpose**: Wire all adapters, validate acceptance criteria, and provide final executable workflow.
- **Dependencies**: TSK-02, TSK-03, TSK-04
- **Supporting Artifacts**:
  - [mac-storage-scout.spec.md](../mac-storage-scout.spec.md)
  - [mac-storage-scout.macos-performance.reference.md](../mac-storage-scout.macos-performance.reference.md)
  - [mac-storage-scout.output-format.source-of-truth.md](../mac-storage-scout.output-format.source-of-truth.md)
- **Runtime Fidelity**: `runtime-hook-required`
- **Deferred Runtime Scope**: Snapshot/purgeable/APFS-clone exact attribution remains deferred.
- **Target Files**:
  - `internal/mss/app/mss_scan_orchestrator.go` (Create)
  - `cmd/mac-storage-scout/main.go` (Update)
  - `README.md` (Create)
- **Target Test Files**:
  - `internal/mss/app/mss_scan_orchestrator_test.go` (Create)
  - `internal/mss/integration/mss_scan_integration_test.go` (Create)
- **Traceability**:
  - **Contract**: [Port: MssScanOrchestratorPort](../mac-storage-scout.spec.md#port-mssscanorchestratorport)
  - **Constraints**: [Acceptance Criteria](../mac-storage-scout.spec.md#8-acceptance-criteria)
- **Implementation Rule**:
  - Add file headers:
    - `// @task spec/tasks/mac-storage-scout.task-05.md`
    - `// @purpose <short-english-ai-to-ai-purpose>`

## 2. Acceptance Criteria (BDD Scenarios)
**Feature**: End-to-end CLI scan report

**Scenario**: CLI profile scan returns readable tree [`runtime-hook`]
- **Given** macos-core profile paths
- **When** scan command is executed
- **Then** report is rendered with explicit `>= threshold` items
- **And** all remaining entries are represented by `other`

**Scenario**: CLI does not fail on partial permission access [`runtime-hook`]
- **Given** at least one restricted path
- **When** scan command is executed
- **Then** command exits successfully
- **And** summary includes non-fatal error counters

## 3. Verification Strategy
- **Test Levels**: `unit`, `integration`
- **Verification Commands**:
  - `go test ./...`
  - `go build ./cmd/mac-storage-scout`
  - `./mac-storage-scout scan --threshold 500MB --profile macos-core --no-progress`
- **Completion Rule**:
  - All tests pass.
  - Binary builds.
  - Smoke run produces expected output shape.

## 4. Test Scenario Coverage
- **Scenario**: CLI profile scan returns readable tree
  - `internal/mss/integration/mss_scan_integration_test.go` :: `should_render_tree_with_threshold_and_other_bucket`
- **Scenario**: CLI does not fail on partial permission access
  - `internal/mss/app/mss_scan_orchestrator_test.go` :: `should_continue_when_some_paths_are_unreadable`

## 5. Execution Log (AI-to-AI)
## EXEC_LOG_META
- task_id: TSK-05
- status: done
- canonical_log: self (this task file)
- evidence_refs:
  - ../evidence/build-test-proof.txt
  - ../evidence/profile-macos-core-proof.txt
  - ../evidence/acceptance-matrix.md

## EXEC_TIMELINE
- 2026-05-01T15:50:00Z: integration wiring validated
- 2026-05-01T16:05:00Z: repository initialized and pushed to GitHub
- 2026-05-01T16:03:00Z: verification `go test ./...` -> PASS
- 2026-05-01T16:03:00Z: verification `go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout` -> PASS
- 2026-05-01T16:03:00Z: verification runtime scan -> PASS
- 2026-05-01T13:54:55Z: validated `--top >= 1` and hardened delete guard against protected descendants/aliases

## EXEC_POINTER
For cross-task summary use `spec/mac-storage-scout.spec.md` section `Decision Summary (Root-Level)`.
