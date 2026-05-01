# mac-storage-scout Spec Index

## Root Spec
- Primary spec: [mac-storage-scout.spec.md](./mac-storage-scout.spec.md)

## Supporting Artifacts
- [mac-storage-scout.macos-performance.reference.md](./mac-storage-scout.macos-performance.reference.md)
- [mac-storage-scout.output-format.source-of-truth.md](./mac-storage-scout.output-format.source-of-truth.md)
- [SESSION-HANDOFF.md](./SESSION-HANDOFF.md)
- [../.ai/README.md](../.ai/README.md)
- [../.ai/rules/go-devgen.contracts.md](../.ai/rules/go-devgen.contracts.md)
- [../.ai/rules/go-qa.testing.md](../.ai/rules/go-qa.testing.md)

## Execution Order (DAG)

> **This table is the living task status tracker.**
> Update it whenever a new task spec is created OR a task changes status.
> Governed by `AGENTS.md::DOCUMENTATION_SYNC_RULES`.

| Task ID | Name | Dependencies | Status |
|---------|------|--------------|--------|
| TSK-01 | Bootstrap CLI + Domain Contract Wiring | — | ✅ DONE |
| TSK-02 | Runtime Walker + Stat Semantics + Progress Counters | TSK-01 | ✅ DONE |
| TSK-03 | Tree Aggregation with Threshold + Other Bucket | TSK-01, TSK-02 | ✅ DONE |
| TSK-04 | ASCII Report Renderer (`other/top-5/rest/types`) | TSK-03 | ✅ DONE |
| TSK-05 | Integration, Validation, and Acceptance Gate | TSK-02, TSK-03, TSK-04 | ✅ DONE |
| TSK-06 | Go DevGen/QA Ruleset + Contract Hardening + Test Expansion | TSK-05 | ✅ DONE |
| TSK-07 | CLI Rename to `mac-storage-scout` + Universal Agent Skill | TSK-05, TSK-06 | ✅ DONE |
| TSK-08 | Master Sync Protocol + Post-Merge Return-To-Master Policy | TSK-07 | ✅ DONE |
| TSK-09 | Performance Benchmark Suite & Engine Optimization Research | TSK-02, TSK-05, TSK-08 | ✅ DONE |
| TSK-10 | Walker Optimization: Dir-Lstat Elimination + Pre-alloc Events | TSK-09 | ⏳ PENDING |

## Execution Agent Rules
- Start from lowest dependency depth.
- Each task file is source of execution details.
- Follow traceability links to root spec instead of duplicating contracts.
- Do not mark task `DONE` without verification evidence in Execution Log.
- **Always update this table** when creating a new task spec or changing task status.
- Use Git flow: work in `ai/<name>` branch, then open PR to `master`.
- PR must follow `.github/PULL_REQUEST_TEMPLATE.md`.
- Merge readiness must satisfy `.github/MERGE_CHECKLIST.md`.
