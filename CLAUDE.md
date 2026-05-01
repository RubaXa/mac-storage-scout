# CLAUDE.md

## DOC_META
- doc_id: CLAUDE_AGENT_ENTRYPOINT
- doc_type: ai_to_ai_entrypoint
- version: 1.1.0
- status: active
- updated_utc: 2026-05-01T15:00:00Z

## SOURCE_OF_TRUTH
- primary: AGENTS.md
- precedence_rule: if conflict exists, AGENTS.md wins

## REQUIRED_BOOTSTRAP
1. **branch context check first** — run `git branch --show-current`. If not on `master`, stop and ask the user whether to switch to `master` or continue on the current branch. Only if on `master` (or user explicitly confirms continuing): `git fetch origin && git pull --ff-only origin master`, then `git checkout -b ai/<name>`. See `AGENTS.md::MASTER_SYNC_PROTOCOL` checkpoint 0.
2. read AGENTS.md
3. read `.agent-skill/SKILL.md` (universal CLI skill — build/scan/delete)
4. read relevant task specs `spec/tasks/*.md` (Execution Log sections)
5. run verification baseline

## VERIFICATION_BASELINE
Run from the project root (path agnostic):
```bash
go test ./...
go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout
```

## BEHAVIOR_RULES
- preserve output contracts from AGENTS.md
- preserve deletion safety contracts from AGENTS.md
- append/update Execution Log in affected task spec file(s)
- follow Git flow policy from AGENTS.md:
  - **sync** `master` (`git fetch origin && git pull --ff-only origin master`) **before** branching, **before** opening a PR, and **after** merge
  - create `ai/<name>` branch off fresh `master`
  - rebase onto `origin/master` before opening a PR; push with `--force-with-lease` (never plain `--force`)
  - open PR to `master` using `.github/PULL_REQUEST_TEMPLATE.md`
  - satisfy `.github/MERGE_CHECKLIST.md`
  - after merge, **proactively offer the user to switch back to `master` and pull**; do not start the next task on a stale feature branch
