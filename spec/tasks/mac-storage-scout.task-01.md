# Task: [TSK-01] - Bootstrap CLI + Domain Contract Wiring

## 1. Meta & Traceability
- **Purpose**: Create the executable CLI entrypoint and core domain/port skeleton.
- **Dependencies**: None
- **Supporting Artifacts**:
  - [mac-storage-scout.macos-performance.reference.md](../mac-storage-scout.macos-performance.reference.md)
  - [mac-storage-scout.output-format.source-of-truth.md](../mac-storage-scout.output-format.source-of-truth.md)
- **Runtime Fidelity**: `contract-only`
- **Deferred Runtime Scope**: Real filesystem traversal is deferred to TSK-02.
- **Target Files**:
  - `go.mod` (Create)
  - `cmd/mac-storage-scout/main.go` (Create)
  - `internal/mss/domain/mss_types.go` (Create)
  - `internal/mss/ports/mss_scan_orchestrator_port.go` (Create)
  - `internal/mss/ports/mss_filesystem_walker_port.go` (Create)
  - `internal/mss/ports/mss_node_aggregator_port.go` (Create)
  - `internal/mss/ports/mss_report_composer_port.go` (Create)
  - `internal/mss/ports/mss_progress_emitter_port.go` (Create)
- **Target Test Files**:
  - `internal/mss/domain/mss_types_test.go` (Create)
- **Traceability**:
  - **Contract**: [Port: MssScanOrchestratorPort](../mac-storage-scout.spec.md#port-mssscanorchestratorport)
  - **Contract**: [Port: MssFilesystemWalkerPort](../mac-storage-scout.spec.md#port-mssfilesystemwalkerport)
  - **Contract**: [Port: MssNodeAggregatorPort](../mac-storage-scout.spec.md#port-mssnodeaggregatorport)
  - **Contract**: [Port: MssReportComposerPort](../mac-storage-scout.spec.md#port-mssreportcomposerport)
  - **Contract**: [Port: MssProgressEmitterPort](../mac-storage-scout.spec.md#port-mssprogressemitterport)
  - **Constraints**: [Assumptions & Constraints](../mac-storage-scout.spec.md#4-assumptions--constraints-context-for-dbc)
  - **Consumer**: [File Mapping](../mac-storage-scout.spec.md#6-file-structure--component-mapping)
- **Implementation Rule**:
  - For every newly created implementation/test file in this ticket, add file-header:
    - `// @task spec/tasks/mac-storage-scout.task-01.md`
    - `// @purpose <short-english-ai-to-ai-purpose>`

## 2. Acceptance Criteria (BDD Scenarios)
**Feature**: CLI bootstrap and stable contracts

**Scenario**: CLI accepts valid scan configuration [`contract`]
- **Given** valid flags for threshold, top and size mode
- **When** CLI parses arguments
- **Then** normalized config is created
- **And** invalid options are rejected with clear message

**Scenario**: Domain constraints enforce valid thresholds [`contract`]
- **Given** threshold value
- **When** threshold parsing is executed
- **Then** positive byte value is produced
- **And** zero/negative input is rejected

## 3. Verification Strategy
- **Test Levels**: `unit`, `contract`
- **Verification Commands**:
  - `go test ./...`
  - `go build ./cmd/mac-storage-scout`
- **Completion Rule**:
  - No TODO in port contracts.
  - Domain tests pass.
  - CLI build succeeds.

## 4. Test Scenario Coverage
- **Scenario**: CLI accepts valid scan configuration
  - `internal/mss/domain/mss_types_test.go` :: `should_parse_valid_scan_config`
- **Scenario**: Domain constraints enforce valid thresholds
  - `internal/mss/domain/mss_types_test.go` :: `should_reject_invalid_threshold`

## 5. Execution Log (AI-to-AI)
## EXEC_LOG_META
- task_id: TSK-01
- status: done
- canonical_log: self (this task file)
- evidence_refs:
  - ../evidence/build-test-proof.txt
  - ../evidence/acceptance-matrix.md

## EXEC_TIMELINE
- 2026-05-01T12:00:00Z: task initialized and scaffold created
- 2026-05-01T12:00:00Z: files created/updated per target list
- 2026-05-01T12:00:00Z: verification `go test ./...` -> PASS
- 2026-05-01T12:00:00Z: verification `go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout` -> PASS

## EXEC_POINTER
For cross-task summary use `spec/mac-storage-scout.spec.md` section `Decision Summary (Root-Level)`.
