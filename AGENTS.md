# AGENTS.md

## DOC_META
- doc_id: AGENTS
- doc_type: ai_to_ai_operating_contract
- version: 1.1.0
- status: active
- repo: mac-storage-scout
- updated_utc: 2026-05-01T14:19:39Z

## EXECUTION_INTENT
Enable deterministic autonomous work on this repository with minimal ambiguity.

## AUTHORITATIVE_READ_ORDER
1. .ai/rules/go-devgen.contracts.md
2. .ai/rules/go-qa.testing.md
3. .ai/checklists/go-change-checklist.md
4. .ai/checklists/go-test-checklist.md
5. .ai/research/go-practices-evidence.md
6. README.md
7. spec/mac-storage-scout.spec.md
8. spec/mac-storage-scout.output-format.source-of-truth.md
9. spec/SESSION-HANDOFF.md
10. spec/tasks/mac-storage-scout.task-*.md (Execution Log sections)

## AI_RULESET_LOCATION
- root: `.ai/`
- mandatory_rules:
  - `.ai/rules/go-devgen.contracts.md`
  - `.ai/rules/go-qa.testing.md`
- mandatory_checklists:
  - `.ai/checklists/go-change-checklist.md`
  - `.ai/checklists/go-test-checklist.md`
- reference_research:
  - `.ai/research/go-practices-evidence.md`

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
1. load `.ai` rule set and affected specs
2. run verification baseline
3. implement scoped change
4. run verification again
5. update docs/specs if behavior changed
6. append Execution Log in the affected task spec file(s)
7. create branch `ai/<name>`
8. commit + push branch
9. open pull request into `master`

## REQUIRED_COMMANDS
```bash
cd /Users/k.lebedev/Developer/mac-storage-scout
go test ./...
go build -o ./bin/mss ./cmd/mss
go run ./cmd/mss-contract-lint --root . --mode short --format text
./bin/mss scan --profile macos-core --threshold 500MB --top 5
./bin/mss delete --dry-run <path> [path...]
./bin/mss delete --yes <path> [path...]
```

## COMMIT_GATES
- pre-commit hook path: `.githooks/pre-commit` (required)
- local git config must set: `core.hooksPath=.githooks`
- commit must be blocked if `go run ./cmd/mss-contract-lint --root . --mode short --format text` returns non-zero

## DOCUMENTATION_SYNC_RULES
When behavior/UX/CLI changes:
- update README.md
- update matching spec file(s)
- append/update Execution Log in affected `spec/tasks/*.md` with UTC timestamp

## GIT_FLOW_POLICY
- direct commits to `master`: forbidden
- required branch naming: `ai/<name>`
- required merge path: PR from `ai/<name>` into `master`
- PR body source: `.github/PULL_REQUEST_TEMPLATE.md` (required)
- merge gate checklist: `.github/MERGE_CHECKLIST.md` (required)
- reviewers/owners policy: `.github/CODEOWNERS`
- before PR:
  - `go test ./...` must pass
  - `go build -o ./bin/mss ./cmd/mss` must pass
- PR description must include:
  - scope summary
  - changed files
  - verification commands and results

## DONE_CRITERIA
- build_pass: true
- tests_pass: true
- output_contract_preserved: true
- delete_safety_contract_preserved: true
- task_execution_log_updated: true
