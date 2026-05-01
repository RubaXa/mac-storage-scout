# 🛰️ Pull Request

```text
 __  __  ____   ____
|  \/  |/ ___| / ___|   mac-storage-scout
| |\/| |\___ \ \___ \   PR quality gate
| |  | | ___) | ___) |
|_|  |_||____/ |____/
```

## ✨ Summary
- What changed and why (2-5 bullets)

## 🎯 Scope
- Included:
  - 
- Excluded:
  - 

## 🧩 Change Map (Purpose-Oriented)
Describe *why* each change group exists and what behavior/process it affects.

| Change Group | Files/Paths | Purpose | Expected Effect |
|---|---|---|---|
| Example: Git workflow policy | `AGENTS.md`, `CLAUDE.md` | enforce `ai/<name>` + PR flow | consistent delivery and review path |
| Example: PR governance | `.github/PULL_REQUEST_TEMPLATE.md` | standardize PR quality signal | faster and clearer reviews |
| Example: Spec alignment | `spec/...` | sync docs with implemented policy | reduced ambiguity for new agents |

## 🧭 Spec / Task Traceability
- Root spec section(s):
  - 
- Task spec(s) updated:
  - `spec/tasks/...` (Execution Log updated)

## 🛡️ Risk Assessment
- Risk level: `low | medium | high`
- Potential regressions:
  - 
- Rollback plan:
  - 

## ✅ Verification
- [ ] `go test ./...`
- [ ] `go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout`
- [ ] runtime check (if behavior changed)

Verification output (short):
```text
paste key output lines here
```

## 🧪 Output Contract Check (if report logic touched)
- [ ] `>= threshold` items are explicit
- [ ] `< threshold` items are only in `other`
- [ ] `other` contains `top-N`, `rest`, `types`

## 🧹 Delete Safety Check (if delete logic touched)
- [ ] `--dry-run` validated first
- [ ] protected paths still blocked
- [ ] real delete requires `--yes`

## 📝 Reviewer Checklist
- [ ] scope is clear
- [ ] implementation matches spec
- [ ] verification is sufficient
- [ ] docs updated (README/spec/task Execution Log)

## 📚 Notes
- Additional context / screenshots / links
