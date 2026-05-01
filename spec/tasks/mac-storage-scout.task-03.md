# Task: [TSK-03] - Tree Aggregation with Threshold + Other Bucket

## 1. Meta & Traceability
- **Purpose**: Build deterministic tree aggregation where all `<threshold` entries are summarized in `other`.
- **Dependencies**: TSK-01, TSK-02
- **Supporting Artifacts**:
  - [mac-storage-scout.output-format.source-of-truth.md](../mac-storage-scout.output-format.source-of-truth.md)
- **Runtime Fidelity**: `contract-only`
- **Deferred Runtime Scope**: None.
- **Target Files**:
  - `internal/mss/adapters/aggregate/mss_tree_aggregator_adapter.go` (Create)
  - `internal/mss/domain/mss_threshold.go` (Create)
- **Target Test Files**:
  - `internal/mss/adapters/aggregate/mss_tree_aggregator_adapter_test.go` (Create)
- **Traceability**:
  - **Contract**: [Port: MssNodeAggregatorPort](../mac-storage-scout.spec.md#port-mssnodeaggregatorport)
  - **Contract**: [Adapter: MssTreeAggregatorAdapter](../mac-storage-scout.spec.md#adapter-msstreeaggregatoradapter)
  - **Constraints**: [Assumptions & Constraints](../mac-storage-scout.spec.md#4-assumptions--constraints-context-for-dbc)
- **Implementation Rule**:
  - Add file headers:
    - `// @task spec/tasks/mac-storage-scout.task-03.md`
    - `// @purpose <short-english-ai-to-ai-purpose>`

## 2. Acceptance Criteria (BDD Scenarios)
**Feature**: Deterministic threshold detail level

**Scenario**: Items above threshold are explicit [`contract`]
- **Given** mixed-size child items
- **When** aggregation runs
- **Then** all `>= threshold` items are explicit nodes

**Scenario**: Items below threshold are merged into other [`contract`]
- **Given** many `<threshold` entries
- **When** aggregation runs
- **Then** they appear only in `other`
- **And** `other` totals match the sum of hidden items

## 3. Verification Strategy
- **Test Levels**: `unit`, `contract`
- **Verification Commands**:
  - `go test ./...`
- **Completion Rule**:
  - Deterministic unit tests for threshold boundaries and total conservation pass.

## 4. Test Scenario Coverage
- **Scenario**: Items above threshold are explicit
  - `internal/mss/adapters/aggregate/mss_tree_aggregator_adapter_test.go` :: `should_emit_all_large_items_explicitly`
- **Scenario**: Items below threshold are merged into other
  - `internal/mss/adapters/aggregate/mss_tree_aggregator_adapter_test.go` :: `should_merge_small_items_into_other_bucket`

## 5. Execution Log
- [ ] `[timestamp]` Task initialized.
- [ ] `[timestamp]` Files created/updated.
- [ ] `[timestamp]` Verification executed: `go test ./...` -> `<pass/fail>`.
