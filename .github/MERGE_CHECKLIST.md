# ✅ Merge Checklist (AI-to-AI)

## Preconditions
- PR branch name matches `ai/<name>`.
- PR targets `master`.
- **Branch is rebased on the latest `origin/master`** — no merge commits from `master` in branch history; PR shows `mergeable: MERGEABLE`.
- At least one reviewer approval.
- No unresolved review comments.

## Required Validation
- `go test ./...` passed.
- `go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout` passed.
- `go run ./cmd/mss-contract-lint --root . --mode short --format text` reported `violations=0`.
- Runtime validation included when behavior changed.

## Documentation Sync
- Root spec updated if decisions changed.
- Relevant task spec Execution Log updated.
- README updated if UX/CLI changed.

## Merge Policy
- Preferred strategy: `Squash and merge`.
- Merge commit message format:
  - `<type>: <short-summary>`

## Post-Merge
- Verify `master` is green.
- Delete feature branch (remote and local).
- **Switch local checkout back to `master` and fast-forward**:
  ```bash
  git checkout master
  git pull --ff-only origin master
  ```
  The agent must offer this to the user proactively — do not begin the next task on a stale feature branch.
