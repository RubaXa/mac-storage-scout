# Go Practices Research Evidence

## Metadata
- generated_utc: 2026-05-01T14:19:39Z
- environment: go1.23.1 darwin/arm64
- intent: adapt TS-oriented DevGen/QA contracts to idiomatic Go contract programming and testing.

## Source Set
1. https://raw.githubusercontent.com/golang/wiki/master/CodeReviewComments.md
2. https://raw.githubusercontent.com/golang/wiki/master/TestComments.md
3. `go doc testing`
4. `go help test`
5. `go help testflag`

## Extracted Rules Used In This Repository

### A. Production Code
- Enforce `gofmt` and idiomatic naming/initialisms.
- Validate errors explicitly; no silent `_` discard.
- Prefer explicit `error` returns over panic for normal failures.
- Keep error strings lowercase/no trailing punctuation.
- Pass `context.Context` as first argument for cancellable operations.
- Keep goroutine lifetimes explicit and bounded.

### B. Testing
- Prefer contract assertions over implementation assertions.
- Use readable subtest names.
- Prefer direct, diagnostic assertions with strong failure messages.
- Include function/input context in failure output.
- Keep tests running (`t.Errorf`) unless setup is broken.
- Use `t.Helper()` for setup helpers.
- Prefer stable semantic comparisons over brittle output-byte comparisons.

## Adaptation Notes From TS XML Rules
- TS JSDoc contract tags were mapped to Go doc comments plus machine tags (`@purpose/@pre/@post/@invariant`).
- TS `cause` chaining mapped to Go `%w` wrapping and `errors.Is/As`.
- TS structural START/END anchors retained as optional but mandatory for non-trivial policy blocks.
- TS test phase model adapted into Go test phase comments and optional anchors for complex tests.

## Exclusions / Non-Goals
- No third-party assertion framework introduced.
- No custom logger framework introduced.
- No broad architecture rewrite beyond contract and test quality.
