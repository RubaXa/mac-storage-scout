# Acceptance Matrix (mac-storage-scout)

## Criteria vs Evidence

1. CLI prints readable tree where large items are explicit (`>= threshold`) — PASS
- Evidence: `spec/evidence/runtime-scan-proof.txt`
- Snapshot line examples: `/Users/k.lebedev/Developer ... 27.2GB`, `@mail-core ... 5.69GB`, nested explicit nodes.

2. Every `other` block follows unified structure (`other`, `top-5`, `rest`, `types`) — PASS
- Evidence: `spec/evidence/runtime-scan-proof.txt`
- Lines show repeated pattern:
  - `other (<500MB each, N items) ...`
  - `top-5:`
  - `rest (... items)`
  - `types:`

3. `--profile macos-core` works and scans required macOS sections — PASS
- Evidence: `spec/evidence/profile-macos-core-proof.txt`
- Output starts from `~/Library/Application Support` and includes deep scan output from profile roots.

4. Default size mode is logical; `--size-mode allocated` is supported — PASS
- Evidence: `spec/evidence/size-mode-allocated-proof.txt`
- Command executes successfully with allocated mode and produces output.

5. Live progress in TTY mode — PASS (captured)
- Evidence: `spec/evidence/progress-tty-proof.txt`
- Captured lines like `scanned: dirs=... files=... bytes=... elapsed=...` from real TTY run.

6. Single binary build — PASS
- Evidence: `spec/evidence/artifact-proof.txt`
- Binary: `./bin/mss`, size ~2.5MB, SHA-256 recorded.

7. Tests and build pass — PASS
- Evidence: `spec/evidence/build-test-proof.txt`
- `go test ./...` passed for implemented packages; `go build` exit=0.

## Known Note
- Earlier deadlock defect during runtime scan was reproduced and fixed in walker queue lifecycle.
- Final proof files above were regenerated after fix.
