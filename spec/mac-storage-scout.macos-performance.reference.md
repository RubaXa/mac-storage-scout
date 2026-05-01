# mac-storage-scout: macOS Performance & Size Semantics Reference

## Objective
Capture normative decisions for correctness and speed on macOS/APFS.

## Size Semantics
- `logical` mode: use `st_size` (default, user-friendly).
- `allocated` mode: use `st_blocks * 512` (inode-reported estimate).

### APFS Caveats
- Sparse files: logical can be much larger than allocated.
- Clones: allocated-by-path can overcount shared extents.
- Snapshots: path traversal cannot attribute snapshot-retained bytes.
- Hardlinks: dedupe by `(dev,inode)` when needed to avoid double count.

## Traversal Performance
- Use iterative bounded worker-pool traversal.
- Prefer `os.ReadDir` + conditional `Info()`.
- Avoid follow-symlink by default.
- Treat `EACCES/EPERM/ENOENT` as non-fatal per-entry errors.
- Keep queue bounded to avoid memory explosion.

## Progress Rendering
- Renderer in dedicated goroutine.
- Worker path only updates counters; renderer refreshes at fixed rate.
- Enable ANSI only when stdout is TTY.

## Verification Guidance
- Benchmark on large trees and compare wall-time vs naive recursive walk.
- Validate that non-fatal errors do not abort scan.
- Validate semantic switch between logical and allocated modes.
