# ✅ Merge Checklist (AI-to-AI)

## Preconditions
- PR branch name matches `ai/<name>`.
- PR targets `master`.
- At least one reviewer approval.
- No unresolved review comments.

## Required Validation
- `go test ./...` passed.
- `go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout` passed.
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
- Delete feature branch.
