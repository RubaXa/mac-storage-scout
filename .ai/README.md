# .ai Rules For `mac-storage-scout`

## Purpose
This folder contains machine-readable engineering rules so autonomous agents can modify this repository without relying on chat history.

## Read Order
1. `.ai/rules/go-devgen.contracts.md`
2. `.ai/rules/go-qa.testing.md`
3. `.ai/research/go-practices-evidence.md`
4. `AGENTS.md`
5. `README.md` + `spec/*`

## Folder Layout
- `rules/`: normative coding and testing contracts.
- `research/`: source-backed rationale and external best practices.

## Mandatory Agent Flow
0. Run project setup for hooks:
   - `./scripts/setup-githooks.sh`
1. Read rules and affected specs.
2. Run baseline verification (`go test ./...`, `go build -o ./bin/mss ./cmd/mss`).
3. Implement minimal scoped change (YAGNI).
4. Add/update tests at the contract boundary.
5. Re-run verification and required runtime commands.
6. Update spec execution log and evidence when behavior/process changed.

## Non-Negotiable Contracts
- Preserve output rules from `spec/mac-storage-scout.output-format.source-of-truth.md`.
- Preserve delete safety contract from `AGENTS.md`.
- Keep scan resilient on permission/race errors.
- Keep deterministic ordering in report output.

## Required Commands
```bash
cd /Users/k.lebedev/Developer/mac-storage-scout
go test ./...
go build -o ./bin/mss ./cmd/mss
./bin/mss scan --profile macos-core --threshold 500MB --top 5
./bin/mss delete --dry-run <path> [path...]
./bin/mss delete --yes <path> [path...]
```
