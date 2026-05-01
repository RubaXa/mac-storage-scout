# Contract Lint Checklist

## Modes
- `--mode short`: only problematic entities.
- `--mode detailed`: all entities with explicit confirmation:
  - required tags for this entity
  - present tags detected
  - missing required tags
  - findings (if any)

## Commands
1. Production, short report:
   - `go run ./cmd/mss-contract-lint --root . --mode short --format text`
2. Production, full confirmation report:
   - `go run ./cmd/mss-contract-lint --root . --mode detailed --format text`
3. Markdown entity index (full):
   - `go run ./cmd/mss-contract-lint --root . --mode detailed --format markdown > spec/evidence/contract-lint-entity-index.md`
4. Include tests if needed:
   - `go run ./cmd/mss-contract-lint --root . --include-tests --mode detailed --format text`
5. Exported API focus:
   - `go run ./cmd/mss-contract-lint --root . --exported-only --mode short --format text`

## Severity semantics
- `error`: required tags missing (`@purpose`, `@consumer`, conditional `@param`, `@returns`).
- `warn`: required tags missing; full contract recheck suggested (`@pre`, `@post`, `@invariant`, `@implements`, `@see`).
