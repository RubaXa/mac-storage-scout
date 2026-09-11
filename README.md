# 🛰️ mac-storage-scout

![mac-storage-scout cover](./assets/github-cover.png)

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

## 🎬 See It In Action

```bash
mac-storage-scout scan --threshold 200MB --top 5 "$HOME/Library/Caches"
```

```text
┌─ 🛰️  mac-storage-scout
│  🎚️  threshold: 200MB | 🔝 top: 5 | 📐 size-mode: logical
└─ 🧭 sections: 1

📂 [~/Library/Caches] 6.4GB
├─ 📁 com.example.Browser ............................. 2.1GB
├─ 📁 com.example.IDE ................................. 1.4GB
├─ 📁 Homebrew ........................................ 612MB
└─ 📦 other (<200MB each, 134 items) .................. 2.3GB
   ├─ 🔝 top-5:
   │  ├─ 📌 pip ....................................... 184MB
   │  ├─ 📌 go-build .................................. 171MB
   │  ├─ 📌 npm ....................................... 156MB
   │  ├─ 📌 yarn ...................................... 142MB
   │  └─ 📌 deno ...................................... 128MB
   ├─ 🧩 rest (129 items) ............................. 1.5GB
   └─ 🧪 types:
      └─ 🏷️  rest types................................ 2.3GB (47821 files)
```

Big stuff is explicit. Small stuff is grouped. One glance — you know where the gigabytes live.

## ✨ Why It Exists
- 📉 Shows where disk space is **actually** consumed — not 80,000 tiny files.
- 🧭 One consistent detail model: big items explicit, small items aggregated into `other` (with `top-N`, `rest`, `types`).
- 🛡️ Safe cleanup flow — `--dry-run` first, `--yes` to commit, protected roots refused.
- 📈 Repeatable incident triage — persistent baseline, path deltas, age buckets, and live-process evidence.
- 🧪 Built for real, noisy live systems — permission errors and races are tolerated, not fatal.
- 🧵 Wide directory trees are drained by a scheduler-owned queue without worker deadlocks.
- 🎨 Emoji mode for humans, `--plain` mode for pipes and CI.

## 🚀 Quick Start
```bash
# build (single static binary, no runtime deps)
go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout

# scan key macOS roots
./bin/mac-storage-scout scan --profile macos-core --threshold 500MB --top 5

# reconcile the complete writable Data volume with readable allocated files
./bin/mac-storage-scout audit --threshold 5GB --top 10 --no-progress

# diagnose fast-changing agent/cache/temp areas and update a persistent baseline
./bin/mac-storage-scout triage --threshold 500MB --top 20

# include slower app data, containers, projects, and downloads
./bin/mac-storage-scout triage --broad --threshold 1GB --top 30

# scan a custom path
./bin/mac-storage-scout scan --threshold 500MB --top 5 "$HOME/Library/Application Support"

# allocated blocks instead of logical size
./bin/mac-storage-scout scan --size-mode allocated --threshold 500MB --profile macos-core

# plain ASCII output (no emoji)
./bin/mac-storage-scout scan --profile macos-core --threshold 500MB --top 5 --plain

# safe delete flow
./bin/mac-storage-scout delete --dry-run <path> [path...]
./bin/mac-storage-scout delete --yes     <path> [path...]
```

## 🛡️ Safety Model for Delete
- Requires `--yes` for actual delete; `--dry-run` previews without touching the filesystem.
- Refuses dangerous roots: `/`, `/System`, `/usr`, `/bin`, `/sbin`, `/private/var/vm`.
- Intended for user caches, logs, and app data cleanup — not system surgery.

## 🎯 Output Contract
For each folder section:
- Items `>= threshold` are printed explicitly.
- Items `< threshold` are grouped into `other`.
- `--top` must be `>= 1`.
- Each `other` always uses the same shape: `top-N`, `rest`, `types`.

## 🧰 Standard Operator Workflow
1. Run `triage` first when free space is falling. The first run records a compact baseline; later runs show growth and shrinkage automatically.
2. Use `triage --broad` when the fast high-churn roots do not explain the loss.
3. Run `audit` when path totals still do not reconcile with APFS volume occupancy.
4. Review `safe`, `review`, `active`, and `inspect` evidence. `safe` is only emitted when process attribution succeeded and found no open files.
5. Run `delete --dry-run` and then `delete --yes` after confirmation.

`audit` uses allocated bytes, stays on one filesystem, and prints any gap between APFS volume occupancy and readable files as `unaccounted`; it never silently treats that gap as explained.

`triage` scans common volatile macOS locations rather than the complete disk. It stores only path totals and capture time in `~/.local/state/mac-storage-scout/triage-v1.json`; it does not store file contents. Explicit positional paths replace the defaults, and `--no-save` preserves the existing baseline.

## 🧱 Project Docs (Spec + Evidence)
- Main spec: `spec/mac-storage-scout.spec.md`
- Output format source of truth: `spec/mac-storage-scout.output-format.source-of-truth.md`
- macOS performance/reference notes: `spec/mac-storage-scout.macos-performance.reference.md`
- ASCII UX research + architecture diagram: `spec/ascii-ux-research.md`
- Universal agent skill: `.agent-skill/SKILL.md`
- Task DAG: `spec/tasks/*.md`
- Runtime/build proofs: `spec/evidence/*`
- Session handoff runbook: `spec/SESSION-HANDOFF.md`
- AI coding/testing rule set: `.ai/README.md`, `.ai/rules/*`

## 🔬 Build / Test
Run from the repository root (path agnostic):
```bash
go test ./...
go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout
```

## 🧱 Contract Lint Gate
```bash
# one-time repo setup
./scripts/setup-githooks.sh

# manual run (same check as pre-commit hook)
go run ./cmd/mss-contract-lint --root . --mode short --format text
```
