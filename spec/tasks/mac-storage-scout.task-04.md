# Task: [TSK-04] - ASCII Report Renderer (`other/top-5/rest/types`)

## 1. Meta & Traceability
- **Purpose**: Render unified readable tree output for all sections.
- **Dependencies**: TSK-03
- **Supporting Artifacts**:
  - [mac-storage-scout.output-format.source-of-truth.md](../mac-storage-scout.output-format.source-of-truth.md)
- **Runtime Fidelity**: `contract-only`
- **Deferred Runtime Scope**: None.
- **Target Files**:
  - `internal/mss/adapters/report/mss_tree_text_report_adapter.go` (Create)
  - `internal/mss/domain/mss_format.go` (Create)
- **Target Test Files**:
  - `internal/mss/adapters/report/mss_tree_text_report_adapter_test.go` (Create)
- **Traceability**:
  - **Contract**: [Port: MssReportComposerPort](../mac-storage-scout.spec.md#port-mssreportcomposerport)
  - **Contract**: [Adapter: MssTreeTextReportAdapter](../mac-storage-scout.spec.md#adapter-msstreetextreportadapter)
  - **Source of Truth**: [Output Format Rules](../mac-storage-scout.output-format.source-of-truth.md#canonical-rules)
- **Implementation Rule**:
  - Add file headers:
    - `// @task spec/tasks/mac-storage-scout.task-04.md`
    - `// @purpose <short-english-ai-to-ai-purpose>`

## 2. Acceptance Criteria (BDD Scenarios)
**Feature**: Human-readable, stable output

**Scenario**: Renderer prints unified other block [`contract`]
- **Given** aggregated node with hidden items
- **When** report is rendered
- **Then** block contains `other`, `top-5`, `rest`, and `types`

**Scenario**: Renderer sorting remains stable [`contract`]
- **Given** equal-size nodes
- **When** report is rendered
- **Then** tie-break by name is deterministic

## 3. Verification Strategy
- **Test Levels**: `unit`, `contract`
- **Verification Commands**:
  - `go test ./...`
- **Completion Rule**:
  - Golden-style output checks pass with deterministic snapshots.

## 4. Test Scenario Coverage
- **Scenario**: Renderer prints unified other block
  - `internal/mss/adapters/report/mss_tree_text_report_adapter_test.go` :: `should_render_other_with_top_rest_and_types`
- **Scenario**: Renderer sorting remains stable
  - `internal/mss/adapters/report/mss_tree_text_report_adapter_test.go` :: `should_sort_output_stably_for_equal_sizes`

## 5. Execution Log (AI-to-AI)
## EXEC_LOG_META
- task_id: TSK-04
- status: done
- canonical_log: self (this task file)
- evidence_refs:
  - ../evidence/runtime-scan-proof.txt
  - ../evidence/acceptance-matrix.md

## EXEC_TIMELINE
- 2026-05-01T15:50:00Z: renderer unified for `other/top/rest/types`
- 2026-05-01T16:03:00Z: ASCII UX improvements + emoji-enhanced readability pass
- 2026-05-01T16:03:00Z: verification `go test ./...` -> PASS

## EXEC_POINTER
For cross-task summary use `spec/mac-storage-scout.spec.md` section `Decision Summary (Root-Level)`.
