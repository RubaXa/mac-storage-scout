# mac-storage-scout Task Tracker

## Root Spec
- Primary spec: [mac-storage-scout.spec.md](./mac-storage-scout.spec.md)

## Supporting Artifacts
- [mac-storage-scout.macos-performance.reference.md](./mac-storage-scout.macos-performance.reference.md)
- [mac-storage-scout.output-format.source-of-truth.md](./mac-storage-scout.output-format.source-of-truth.md)
- [SESSION-HANDOFF.md](./SESSION-HANDOFF.md)

## Execution Order (DAG)

| Task ID | Name | Dependencies | Status |
|---|---|---|---|
| TSK-01 | Bootstrap CLI + domain contracts wiring | None | [x] DONE |
| TSK-02 | Runtime walker + stat semantics + progress counters | TSK-01 | [x] DONE |
| TSK-03 | Tree aggregation with threshold and `other` bucket | TSK-01, TSK-02 | [x] DONE |
| TSK-04 | ASCII report renderer (`other/top-5/rest/types`) | TSK-03 | [x] DONE |
| TSK-05 | Integration, tests, validation, acceptance check | TSK-02, TSK-03, TSK-04 | [x] DONE |

## Execution Agent Rules
- Start from lowest dependency depth.
- Each task file is source of execution details.
- Follow traceability links to root spec instead of duplicating contracts.
- Do not mark task `[x] DONE` without verification evidence in Execution Log.
- Always implement/verify `Target Test Files` ownership.
- Use Git flow: work in `ai/<name>` branch, then open PR to `master`.
- PR must follow `.github/PULL_REQUEST_TEMPLATE.md`.
- Merge readiness must satisfy `.github/MERGE_CHECKLIST.md`.
