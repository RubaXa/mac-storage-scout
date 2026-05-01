# Go Test Checklist

1. Scenario set includes happy path, boundary, and failure path.
2. Tests verify contract-visible behavior only.
3. Failure messages include function + relevant input.
4. Prefer `t.Errorf` over `t.Fatalf` except setup blockers.
5. Subtests are human-readable.
6. Table-driven tests used when logic is shared.
7. Error checks use `errors.Is/As` when semantic type matters.
8. Test helpers call `t.Helper()`.
9. Tests are deterministic and isolated (`t.TempDir`, no machine-state coupling).
10. Verification:
    - `go test ./...`
    - `go test -count=1 ./...`
