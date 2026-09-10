# mac-storage-scout: Output Format Source of Truth

## Canonical Rules
1. Any item with `size >= threshold` is printed explicitly.
2. Any item with `size < threshold` is included only in `other` bucket.
3. `other` format is mandatory and identical everywhere:

```text
other (<threshold each, N items) ... <total>
  top-5:
  <name1> ... <size>
  ...
  rest (N-5 items) ... <size>
  types:
  .extA ... <size> (<count>)
  .extB ... <size> (<count>)
  rest types ... <size> (<count>)
```

4. If N < 5, show available items and `rest` with zero or omit `rest` line.
5. Sizes are printed in human units (binary GiB/MiB/KiB) with stable rounding.
6. Sorting:
   - explicit items: size desc, then name asc.
   - top-5 in `other`: size desc.
   - types: size desc.

## Section Baseline (macos-core profile)
- `~/Library/Containers`
- `~/Library/Application Support`
- `/private/var/vm`
- `/private/var/folders`

## Volume Audit Balance
`audit` prints this reconciliation before the canonical tree:

```text
mac-storage-scout audit
volume: /System/Volumes/Data
capacity: <size>
occupied: <size>
available: <size>
accounted-by-readable-files: <size> (<percent>)
unaccounted: <size>
growth-during-scan: <signed-size>
scan-errors: <count>
```

- `occupied` is `statfs capacity - available` at scan completion.
- `accounted-by-readable-files` is the allocated-size total reachable from the scan root on the same filesystem.
- `unaccounted` is always explicit; it is never silently omitted from the balance.
- If allocated file totals exceed filesystem occupancy (for example because of APFS clone sharing), print `allocated-size-overcount` instead of a negative remainder.
- If `--volume` names a path narrower than its containing mount, print `scan-root` and `scope-warning`.
- The tree following the balance keeps every canonical threshold/`other` rule above.

## Incident Triage Report

`triage` is an outcome report, not a canonical scan-tree section. It prints:

```text
mac-storage-scout triage
captured: <time>
occupied: <size>
available: <size>
since-baseline: <time> | volume-growth: <signed-size>
scan-errors: <count>
baseline: <path> (updated|read-only)

changes:
  <path> delta=<signed-size> now=<size> old>7d=<size>

hotspots:
  <path> size=<size> delta=<signed-size> old>7d=<size> status=<safe|review|active|inspect> owner=<process|inactive>

candidates:
  [safe|review] <path> old-bytes=<size> — <reason>
```

- `changes` is ordered by absolute delta so reclaimed paths remain visible.
- `hotspots` is ordered by current allocated size.
- `safe` requires successful process evidence, no open file below the path, a generic cache/log/temp classification, and bytes older than seven days.
- `review` is used for generated or historical structures such as worktrees, snapshots, sessions, and `node_modules`.
- `active` and `inspect` are never presented as delete-ready candidates.
- Candidate paths never overlap, so parent and child sizes cannot be counted twice.
- The baseline stores path totals and volume occupancy only; no file contents are persisted.

## Forbidden Output Patterns
- Mixing custom per-section formats.
- Omitting `other` detail where small items exist.
- Showing items `< threshold` as explicit siblings of `>= threshold` items.
