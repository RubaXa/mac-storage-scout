# AGENTS.md

## DOC_META
- doc_id: AGENTS
- doc_type: ai_to_ai_operating_contract
- version: 1.0.0
- status: active
- repo: mac-storage-scout
- updated_utc: 2026-05-01T13:12:00Z

## EXECUTION_INTENT
Enable deterministic autonomous work on this repository with minimal ambiguity.

## AUTHORITATIVE_READ_ORDER
1. README.md
2. spec/mac-storage-scout.spec.md
3. spec/mac-storage-scout.output-format.source-of-truth.md
4. spec/SESSION-HANDOFF.md
5. spec/tasks/mac-storage-scout.task-*.md (Execution Log sections)

## HARD_CONTRACTS
- output.large_items:
  - rule: print every item where size >= threshold as explicit node
- output.small_items:
  - rule: include every item where size < threshold only inside `other`
- output.other_shape:
  - required_blocks:
    - top-N
    - rest
    - types
  - invariance: same structure in every section
- deletion.safety:
  - must_run_dry_run_before_real_delete: true
  - real_delete_requires_yes_flag: true
  - protected_paths:
    - /
    - /System
    - /usr
    - /bin
    - /sbin
    - /private/var/vm

## OPERATIONAL_WORKFLOW
1. run verification baseline
2. implement scoped change
3. run verification again
4. update docs/specs if behavior changed
5. append Execution Log in the affected task spec file(s)
6. commit + push

## REQUIRED_COMMANDS
```bash
cd /Users/k.lebedev/Developer/mac-storage-scout
go test ./...
go build -o ./bin/mss ./cmd/mss
./bin/mss scan --profile macos-core --threshold 500MB --top 5
./bin/mss delete --dry-run <path> [path...]
./bin/mss delete --yes <path> [path...]
```

## DOCUMENTATION_SYNC_RULES
When behavior/UX/CLI changes:
- update README.md
- update matching spec file(s)
- append/update Execution Log in affected `spec/tasks/*.md` with UTC timestamp

## DONE_CRITERIA
- build_pass: true
- tests_pass: true
- output_contract_preserved: true
- delete_safety_contract_preserved: true
- task_execution_log_updated: true
