# CLAUDE.md

## DOC_META
- doc_id: CLAUDE_AGENT_ENTRYPOINT
- doc_type: ai_to_ai_entrypoint
- version: 1.0.0
- status: active
- updated_utc: 2026-05-01T13:12:00Z

## SOURCE_OF_TRUTH
- primary: AGENTS.md
- precedence_rule: if conflict exists, AGENTS.md wins

## REQUIRED_BOOTSTRAP
1. read AGENTS.md
2. read relevant task specs `spec/tasks/*.md` (Execution Log sections)
3. run verification baseline

## VERIFICATION_BASELINE
```bash
cd /Users/k.lebedev/Developer/mac-storage-scout
go test ./...
go build -o ./bin/mss ./cmd/mss
```

## BEHAVIOR_RULES
- preserve output contracts from AGENTS.md
- preserve deletion safety contracts from AGENTS.md
- append/update Execution Log in affected task spec file(s)
