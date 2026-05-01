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

## Forbidden Output Patterns
- Mixing custom per-section formats.
- Omitting `other` detail where small items exist.
- Showing items `< threshold` as explicit siblings of `>= threshold` items.
