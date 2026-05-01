# Task: [TSK-02] - Runtime Walker + Stat Semantics + Progress Counters

## 1. Meta & Traceability
- **Purpose**: Implement high-performance filesystem walker with logical/allocated size semantics and non-fatal error handling.
- **Dependencies**: TSK-01
- **Supporting Artifacts**:
  - [mac-storage-scout.macos-performance.reference.md](../mac-storage-scout.macos-performance.reference.md)
- **Runtime Fidelity**: `runtime-hook-required`
- **Deferred Runtime Scope**: APFS clone-aware physical ownership is deferred.
- **Target Files**:
  - `internal/mss/adapters/fs/mss_go_fs_walker_adapter.go` (Create)
  - `internal/mss/adapters/fs/mss_sys_stat_adapter.go` (Create)
  - `internal/mss/adapters/progress/mss_ansi_progress_adapter.go` (Create)
- **Target Test Files**:
  - `internal/mss/adapters/fs/mss_go_fs_walker_adapter_test.go` (Create)
  - `internal/mss/adapters/fs/mss_sys_stat_adapter_test.go` (Create)
- **Traceability**:
  - **Contract**: [Port: MssFilesystemWalkerPort](../mac-storage-scout.spec.md#port-mssfilesystemwalkerport)
  - **Contract**: [Adapter: MssGoFsWalkerAdapter](../mac-storage-scout.spec.md#adapter-mssgofswalkeradapter)
  - **Contract**: [Adapter: MssSysStatAdapter](../mac-storage-scout.spec.md#adapter-msssysstatadapter)
  - **Constraints**: [Assumptions & Constraints](../mac-storage-scout.spec.md#4-assumptions--constraints-context-for-dbc)
- **Implementation Rule**:
  - Add file headers:
    - `// @task spec/tasks/mac-storage-scout.task-02.md`
    - `// @purpose <short-english-ai-to-ai-purpose>`

## 2. Acceptance Criteria (BDD Scenarios)
**Feature**: Runtime metadata scan resilience and speed

**Scenario**: Walker continues on permission-denied paths [`runtime-hook`]
- **Given** mixed readable and restricted directories
- **When** scan runs
- **Then** readable subtree is fully processed
- **And** restricted paths are counted as non-fatal errors

**Scenario**: Size mode switch changes accounting semantics [`runtime-hook`]
- **Given** files with known metadata
- **When** mode `logical` is selected
- **Then** size is based on `st_size`
- **And** mode `allocated` uses `st_blocks*512`

## 3. Verification Strategy
- **Test Levels**: `unit`, `integration`
- **Verification Commands**:
  - `go test ./...`
- **Completion Rule**:
  - Walker tests cover non-fatal errors and mode switching.

## 4. Test Scenario Coverage
- **Scenario**: Walker continues on permission-denied paths
  - `internal/mss/adapters/fs/mss_go_fs_walker_adapter_test.go` :: `should_continue_on_permission_errors`
- **Scenario**: Size mode switch changes accounting semantics
  - `internal/mss/adapters/fs/mss_sys_stat_adapter_test.go` :: `should_compute_logical_and_allocated_sizes`

## 5. Execution Log
- [ ] `[timestamp]` Task initialized.
- [ ] `[timestamp]` Files created/updated.
- [ ] `[timestamp]` Verification executed: `go test ./...` -> `<pass/fail>`.
