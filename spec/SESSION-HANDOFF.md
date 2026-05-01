# Session Handoff Runbook (For Any New Agent)

## Goal
Continue work autonomously with deterministic behavior, consistent reporting, and proof artifacts.

## Canonical Paths
- Project root: `/Users/k.lebedev/Developer/mac-storage-scout`
- Binary: `/Users/k.lebedev/Developer/mac-storage-scout/bin/mss`
- Specs: `/Users/k.lebedev/Developer/mac-storage-scout/spec`
- Evidence: `/Users/k.lebedev/Developer/mac-storage-scout/spec/evidence`

## Required Reading Order
1. `.ai/rules/go-devgen.contracts.md`
2. `.ai/rules/go-qa.testing.md`
3. `.ai/checklists/go-change-checklist.md`
4. `.ai/checklists/go-test-checklist.md`
5. `README.md`
6. `spec/mac-storage-scout.spec.md`
7. `spec/mac-storage-scout.output-format.source-of-truth.md`
8. `spec/tasks/mac-storage-scout.task-*.md`
9. latest files in `spec/evidence/`

## Non-Negotiable Contracts
- Print all items `>= threshold` explicitly.
- Group all items `< threshold` into `other`.
- `other` must be identical in every section:
  - `top-N`
  - `rest`
  - `types`
- Tool must stay fast on live macOS filesystem and tolerate permission/race errors.

## Runtime Profiles
- `--profile macos-core` scans:
  - `~/Library/Containers`
  - `~/Library/Application Support`
  - `/private/var/vm`
  - `/private/var/folders`

## Safety Policy
- Never delete without `--dry-run` check first.
- Never delete protected system paths.
- Prefer removing logs/caches/known unused app data.
- Treat `Claude/vm_bundles` as opt-in only (user decision required).

## Evidence Protocol
After each meaningful change, append proofs under `spec/evidence/`:
- command run
- key output lines
- before/after size
- any failure and fallback

Minimum proof set:
- build: `go build` success
- tests: `go test ./...` success
- runtime scan sample output
- delete dry-run and delete output samples

## Standard Commands
```bash
cd /Users/k.lebedev/Developer/mac-storage-scout

# verify
go test ./...
go build -o ./bin/mss ./cmd/mss

# scan
./bin/mss scan --profile macos-core --threshold 500MB --top 5

# targeted scan
./bin/mss scan --threshold 500MB --top 5 "$HOME/Library/Application Support"

# delete workflow
./bin/mss delete --dry-run <path> [path...]
./bin/mss delete --yes <path> [path...]
```

## Acceptance Checklist
- Binary runs from `bin/mss`.
- Report formatting is readable and aligned in monospaced terminal.
- `other/top/rest/types` is present and consistent.
- `delete` command enforces confirmation and path guards.
- Re-scan shows measurable reclaimed space when deletion executed.

## Known Cleanup Strategy (Low Risk First)
1. App logs.
2. Browser/app caches.
3. Crash dumps.
4. Unused app payloads explicitly confirmed by user.
5. High-risk/system-like data only with explicit user approval.
