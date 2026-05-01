# Go DevGen Contract Rules (AI-to-AI)

## 1. Primary Goal
Produce Go code that acts as an executable specification: behavior must be recoverable from code, contracts, and high-signal comments without chat context.

## 2. Scope
Applies to all `.go` production files in this repository (`cmd/`, `internal/`).

## 3. Non-Negotiable Principles
- Contract-first: every non-trivial behavior has explicit preconditions, postconditions, and invariants in code and tests.
- YAGNI: no extension points, abstractions, or flags unless demanded by current contract.
- Determinism: sorting, formatting, and output shape must be stable.
- Safety before convenience: delete/scan guardrails are never weakened.

## 4. Structural Rules
- Keep package names short, lowercase, and meaningful.
- Use `gofmt` formatting.
- Keep exported API minimal.
- Avoid global mutable state.
- Keep interfaces at consumer boundaries, not speculative layers.

## 5. Contract Comment Protocol
Use Go doc comments plus machine tags for contract recovery.

### 5.1 Exported types/functions
Use this shape when logic is non-trivial:
```go
// MssXxx does Yyy.
//
// @purpose One-sentence business intent.
// @pre Required caller/system assumptions.
// @post Guaranteed outcome on success.
// @invariant Rule that must stay true after refactors.
```

### 5.2 Interface implementations
- On implementation type: `@implements {PortName} <path>`.
- On methods that mirror interface behavior: `@see {PortName#Method} <path>`.
- If behavior diverges from interface contract, write explicit `@pre/@post/@invariant` instead of `@see` only.

## 6. Control-Flow Anchors
Anchors are mandatory for non-trivial blocks (policy branches, failure boundaries, concurrency boundaries):
- `// START_<CONTEXT>_<INTENT>`
- `// END_<CONTEXT>_<INTENT>`

Use intent names (why), not syntax names (what).

Do not anchor trivial one-line guards with no hidden policy.

## 7. Error Semantics
- Never use `panic` for expected runtime failures.
- Return `error` with context and wrapping: `fmt.Errorf("[Trace] ...: %w", err)`.
- Error strings start lowercase and avoid trailing punctuation.
- Validate arguments early (fail-fast precondition checks).
- Preserve causal chains with `%w` and inspect via `errors.Is/As`.

Trace prefix format:
- Function: `[MssParseBytes]`
- Method: `[MssScanOrchestrator.Run]`

## 8. Context and Concurrency
- `context.Context` is first arg when cancellation/timeout applies.
- Do not store context in structs.
- Every goroutine must have deterministic shutdown.
- Channel ownership and close responsibilities must be explicit.
- Avoid blocking hot worker paths on UI/progress output.

## 9. I/O and Side Effects
- Filesystem/process/network side effects must be isolated at adapter/CLI boundaries.
- Domain and aggregation logic should stay pure where possible.
- Symlink and protected-path policies are explicit contracts; never implicit.

## 10. Naming Rules
- Use business-intent names over generic words (`Process`, `Manager`, `Handler` as standalone names are discouraged).
- Keep abbreviations standard (`ID`, `URL`, `EOF`), follow Go initialism style.

## 11. Logging Rules (Repository-Specific)
- For this CLI, prefer user-facing stderr/stdout messages at command boundaries.
- Do not add noisy logs inside hot loops unless contract requires observability.
- If adding logs, include trace prefix and state transition intent.

## 12. Refactor Safety Checklist
Before finalizing any code change, verify:
1. Output contract is intact (`>= threshold` explicit, `< threshold` only in `other`).
2. Delete safety is intact (`--yes`, dry-run, protected paths).
3. Error flow still preserves root cause.
4. Concurrency lifecycle still terminates cleanly.
5. All touched exported entities are documented.
