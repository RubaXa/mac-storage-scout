# Go Change Checklist

1. Read `.ai/rules/go-devgen.contracts.md` and affected specs.
2. Run baseline:
   - `go test ./...`
   - `go build -o ./bin/mss ./cmd/mss`
3. Validate preconditions at API/command boundaries.
4. Preserve output contract invariants.
5. Preserve delete safety invariants.
6. Add/update contract comments for changed exported surfaces.
7. Add anchors for non-trivial new control-flow blocks.
8. Re-run tests and build.
9. Run runtime smoke commands as required by `AGENTS.md`.
10. Update affected `spec/tasks/*.md` execution logs.
