# Contract Lint Checklist

1. Run linter on production code:
   - `go run ./cmd/mss-contract-lint --root . --format text`
2. If needed, include tests too:
   - `go run ./cmd/mss-contract-lint --root . --include-tests --format text`
3. For agent handoff evidence, export markdown report:
   - `go run ./cmd/mss-contract-lint --root . --format markdown > spec/evidence/contract-lint-entity-index.md`
4. For API-only checks, use exported mode:
   - `go run ./cmd/mss-contract-lint --root . --exported-only --format text`
5. Interpretation:
   - `error`: required tags missing (`@purpose`, `@consumer`, conditional `@param`, `@returns`).
   - `warn`: required tags missing, therefore full contract recheck required (`@pre`, `@post`, `@invariant`, `@implements`, `@see`).
6. Fix workflow:
   - Address all `error` findings first.
   - Re-run linter until `errors=0`.
   - Then optionally refine optional tags where warnings suggested full recheck.
