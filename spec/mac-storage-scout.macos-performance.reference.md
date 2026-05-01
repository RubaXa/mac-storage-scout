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

## Measured Results (TSK-09, 2026-05-01)

### Machine
Apple M3 Pro, arm64, macOS 14 (Sonoma), APFS, Go 1.22, 12 logical CPUs

### Tree: ~/Developer — 476k entries, 33GB

| Strategy                     | Wall Time | ns/entry  | B/op       | allocs/op  | vs baseline |
|------------------------------|-----------|-----------|------------|------------|-------------|
| **C0 WalkerCurrent (12w)**   | 16.0s     | 33.6 µs   | 2 089 MB   | 18 056 847 | **baseline** |
| C3 WalkerCurrent (6w)        | 21.4s     | 44.9 µs   | 2 089 MB   | 18 056 587 | +33% slower |
| C4 WalkerCurrent (1w)        | 139s      | 292.0 µs  | 2 089 MB   | 18 056 550 | **+769% slower** |
| C1 filepath.WalkDir (seq)    | 29.2s     | 61.4 µs   | 901 MB     | 9 178 705  | +83% slower, 2.3x less memory |
| C2 filepath.Walk (seq)       | 69.2s     | 145.3 µs  | 1 490 MB   | 11 466 411 | +333% slower |

### Small tree (os.TempDir, repeated iterations)

| Strategy                     | ns/op     | B/op      | allocs/op |
|------------------------------|-----------|-----------|-----------|
| **C0 WalkerCurrent (12w)**   | 5.5ms     | 3.88 MB   | 14 233    |
| C3 WalkerCurrent (6w)        | 5.5ms     | 3.88 MB   | 14 217    |
| C4 WalkerCurrent (24w)       | 5.8ms     | 3.89 MB   | 14 262    |
| C4 WalkerCurrent (1w)        | 12.6ms    | 3.88 MB   | 14 202    |
| C1 filepath.WalkDir          | 7.9ms     | 0.97 MB   | 8 089     |
| C2 filepath.Walk             | 11.4ms    | 1.38 MB   | 9 714     |

### Hotspot (pprof top-5, BenchmarkWalkerCurrent on ~/Developer)

1. `syscall.syscall` — **89.85% of CPU** (fully syscall-bound, no Go-level optimization possible)
2. `os.Lstat` — **74.65%** (one Lstat per directory entry, required for size on macOS/APFS)
3. `os.(*unixDirent).Info` — 63.74% (underlies the Lstat calls from de.Info())
4. `os.ReadDir` — 22.86% (directory listing)
5. `os.openDir` — 12.10% (directory open syscall)

### Optimization Decisions

| Candidate                                   | Verdict  | Reason                                                   |
|---------------------------------------------|----------|----------------------------------------------------------|
| Pre-alloc events slice (heuristic cap)      | TSK-10   | ~5-15% alloc reduction; 2.09GB for 476k entries is high |
| Skip redundant Lstat on enqueued dirs       | TSK-10   | ~238k fewer Lstats; est. 15-30% wall-time gain           |
| Stream events via channel (no mutex)        | Low ROI  | Lock is not in hot path per profile; mutex overhead <1%  |
| Tune worker count (current NumCPU×2)        | Done     | NumCPU×2 is optimal; NumCPU 33% slower on large tree    |
| Replace filepath.Join with string concat    | Skip     | Not in pprof top-20; complexity > gain                   |
| Increase queue buffer (65536→131072)        | Skip     | Queue never saturates on tested trees                    |
| sync.Pool for event structs                 | TSK-10   | High alloc count (18M for 476k); pool may help           |

### Follow-up
- **TSK-10**: implement dir-Lstat elimination (pass `isDir` hint in queue item) + pre-alloc events slice. Target: ≥15% wall-time improvement on large trees.
