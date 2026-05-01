# Go QA Testing Rules (AI-to-AI)

## 1. Primary Goal
Generate contract-focused, deterministic tests that validate behavior at public boundaries and remain stable under refactor.

## 2. Scope
Applies to all `_test.go` files and test assets.

## 3. Test Design Axioms
- Test contract, not private mechanics.
- Prefer black-box behavior checks unless white-box access is required for package-local contracts.
- Each test should protect one scenario.
- Keep tests deterministic (time, randomness, and FS state controlled).

## 4. Standard Case Flow
For non-trivial cases, use explicit phases:
- `SETUP`
- `TRIGGER`
- `OBSERVE` (optional)
- `ASSERT`

Use paired anchors for meaningful non-trivial phases:
- `// START_<CASE>_<PHASE>_<INTENT>`
- `// END_<CASE>_<PHASE>_<INTENT>`

Skip anchor noise for trivial single-assertion tests.

## 5. Naming and Structure
- Prefer table-driven tests for same-logic multiple inputs.
- Use subtests with human-readable names (`t.Run("rejects zero top", ...)`).
- Split test functions when scenario families require different assertion logic.

## 6. Assertion Policy
- Use standard `testing` + direct comparisons by default.
- Failure message format: `Func(input) = got, want`.
- Print `got` before `want`.
- Keep going with `t.Errorf` when possible; reserve `t.Fatalf` for setup blockers.
- For error semantics, verify with `errors.Is` / `errors.As` when relevant.

## 7. Contract Coverage Minimum
For each contract surface, cover:
- happy path
- boundary conditions
- failure path
- invariants (sorting/order/aggregation consistency)

For this repository, explicitly cover:
- threshold boundary (`>= threshold` vs `< threshold`)
- mandatory `other` block shape (`top-N`, `rest`, `types`)
- delete safety guards and symlink alias cases
- non-fatal scan errors and continuation behavior

## 8. Test Helpers
- Mark helper functions with `t.Helper()`.
- Helpers may hide setup noise but must not hide test intent.
- Do not build custom assertion mini-frameworks.

## 9. Filesystem and Runtime Tests
- Use `t.TempDir()` for writable sandboxed fixtures.
- Avoid dependence on user machine state for unit tests.
- Keep integration tests explicit and optionally skippable in `-short` mode when expensive.

## 10. Fuzzing Guidance
Use fuzz tests when input normalization/parsing/formatting has broad input surface (e.g., size parser). Keep fuzz target deterministic and fast.

## 11. Required Verification Commands
```bash
go test ./...
go test -count=1 ./...
go test -race ./...   # when environment/time allows
```

## 12. Test Review Checklist
1. Does each test assert observable contract behavior?
2. Is failure output diagnostic and input-aware?
3. Are edge cases and error semantics covered?
4. Are tests deterministic and isolated?
5. Are subtests/table entries named for fast triage?
