# 🛰️ mac-storage-scout

> ⚡ Fast, readable, action-oriented disk explorer for macOS terminal.
>  
> Find what eats your disk, decide safely, clean confidently.

```text
 __  __  ____   ____
|  \/  |/ ___| / ___|   mac-storage-scout
| |\/| |\___ \ \___ \   map disk usage, keep output readable
| |  | | ___) | ___) |
|_|  |_||____/ |____/
```

## ✨ Why It Exists
- 📉 Shows where disk space is actually consumed.
- 🧭 Uses one consistent detail model: big items explicit, small items aggregated.
- 🛡️ Supports safe cleanup flow with `dry-run` and guarded delete.
- 🧪 Built for real, noisy live systems (permissions/race errors tolerated).

## 📦 Binary
- `bin/mss`

## 🚀 Quick Start
```bash
# scan selected roots
./bin/mss scan --profile macos-core --threshold 500MB --top 5

# scan custom path
./bin/mss scan --threshold 500MB --top 5 "$HOME/Library/Application Support"

# use allocated blocks instead of logical file size
./bin/mss scan --size-mode allocated --threshold 500MB --top 5 --profile macos-core

# force ASCII/plain output (no emoji icons)
./bin/mss scan --profile macos-core --threshold 500MB --top 5 --plain

# delete workflow
./bin/mss delete --dry-run <path> [path...]
./bin/mss delete --yes <path> [path...]
```

## 🎯 Output Contract (Short)
For each folder section:
- Items `>= threshold` are printed explicitly.
- Items `< threshold` are grouped into `other`.
- `--top` must be `>= 1`.
- Each `other` uses one identical structure:
  - `top-N`
  - `rest`
  - `types`

Example shape:
```text
[/some/path] 27.2GB
├─ big-item-A ......................................... 5.6GB
├─ big-item-B ......................................... 3.2GB
└─ other (<500MB each, 877 items) ..................... 688MB
   ├─ top-5:
   │  googleapis ...................................... 119MB
   │  @vkontakte ...................................... 118MB
   │  typescript ...................................... 63.8MB
   │  ...
   ├─ rest (872 items) ................................ 296MB
   └─ types:
      .json ........................................... 640KB (1 files)
```

## 🖼️ Styled Output (Emoji Mode)
```text
┌─ 🛰️  mss :: mac-storage-scout
│  🎚️  threshold: 500MB | 🔝 top: 5 | 📐 size-mode: logical
└─ 🧭 sections: 4

📂 [/Users/you/Library/Application Support] 45.9GB
├─ 📁 Claude........................................... 22.5GB
└─ 📦 other (<500MB each, 877 items) .................. 688MB
   ├─ 🔝 top-5:
   │  ├─ 📌 googleapis................................. 119MB
   │  └─ 📌 @vkontakte................................. 118MB
   ├─ 🧩 rest (872 items) ............................. 296MB
   └─ 🧪 types:
      ├─ 🏷️  .json..................................... 640KB (1 files)
      └─ 🏷️  rest types................................ 3.2GB (5912 files)
```

## 🛡️ Safety Model for Delete
`mss delete` has explicit safeguards:
- Requires `--yes` for actual delete.
- `--dry-run` available for verification.
- Refuses dangerous roots: `/`, `/System`, `/usr`, `/bin`, `/sbin`, `/private/var/vm`.
- Intended for user caches/logs/app data cleanup only.

## 🧱 Project Docs (Spec + Evidence)
- Main spec: `spec/mac-storage-scout.spec.md`
- Output format source of truth: `spec/mac-storage-scout.output-format.source-of-truth.md`
- macOS performance/reference notes: `spec/mac-storage-scout.macos-performance.reference.md`
- ASCII UX research + architecture ASCII diagram: `spec/ascii-ux-research.md`
- Task DAG: `spec/tasks/*.md`
- Runtime/build proofs: `spec/evidence/*`
- Session handoff runbook: `spec/SESSION-HANDOFF.md`
- AI coding/testing rule set: `.ai/README.md`, `.ai/rules/*`

## 🧰 Standard Operator Workflow
1. Scan with `--profile macos-core`.
2. Pick candidates (logs/caches/unused app data).
3. Run `delete --dry-run`.
4. Run `delete --yes` after confirmation.
5. Re-scan and record reclaimed space.

## 🔬 Build / Test
```bash
cd /Users/k.lebedev/Developer/mac-storage-scout
go test ./...
go build -o ./bin/mss ./cmd/mss
```

## 🧱 Contract Lint Gate
```bash
# one-time repo setup
./scripts/setup-githooks.sh

# manual run (same check as pre-commit hook)
go run ./cmd/mss-contract-lint --root . --mode short --format text
```
