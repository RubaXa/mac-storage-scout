# ASCII UX Research (CLI Readability)

## Sources Reviewed
- CLIG: https://clig.dev/ (output consistency, progressive detail, human-first output)
- NO_COLOR: https://no-color.org/ (avoid hard dependency on color for semantics)
- Box-drawing conventions: https://en.wikipedia.org/wiki/Box-drawing_character

## Decisions for `mss`
1. Keep output useful without color; structure must be readable in plain monochrome terminal.
2. Use a single tree grammar everywhere:
   - `├─` intermediate
   - `└─` final
   - `│` vertical continuation
3. Keep one invariant shape for aggregated buckets:
   - `other`
   - `top-N`
   - `rest`
   - `types`
4. Use dotted padding only between label and size, never for section headers.
5. Add compact report header with scan context (threshold, top-N, size-mode, section count).

## Canonical Pretty Layout
```text
┌─ mss :: mac-storage-scout
│  threshold: 500MB | top: 5 | size-mode: logical
└─ sections: 4

[/path] 27.2GB
├─ heavy-folder........................................... 5.6GB
└─ other (<500MB each, 877 items) ........................ 688MB
   ├─ top-5:
   │  ├─ googleapis........................................ 119MB
   │  ├─ @vkontakte........................................ 118MB
   │  └─ typescript........................................ 63.8MB
   ├─ rest (872 items)..................................... 296MB
   └─ types:
      ├─ .mp4.............................................. 9.8GB (2104 files)
      ├─ .zip.............................................. 6.1GB (84 files)
      └─ rest types........................................ 3.2GB (5912 files)
```

## ASCII Architecture Diagram (for docs)
```text
+------------------+      +---------------------------+
| CLI (cmd/mss)    +----->+ Scan Orchestrator (app)   |
+------------------+      +-------------+-------------+
                                         |
                  +----------------------+----------------------+
                  |                      |                      |
        +---------v----------+ +---------v----------+ +---------v----------+
        | FS Walker Adapter  | | Tree Aggregator    | | Progress Adapter   |
        | (ReadDir + stat)   | | (threshold/other)  | | (TTY status line)  |
        +---------+----------+ +---------+----------+ +--------------------+
                  |                      |
        +---------v----------+ +---------v----------+
        | SysStat Adapter    | | Report Adapter     |
        | logical/allocated  | | ASCII tree output  |
        +--------------------+ +--------------------+
```
