# Skill: mac-storage-scout

## Purpose
Install and use the `mac-storage-scout` CLI as a deterministic disk-usage scanner with a safe delete workflow.

## Supported Environments
- Codex
- Claude
- OpenCode
- Any agent runtime with shell access (`bash`/`zsh`) and a Go toolchain (`go >= 1.22`)

## Tool Name Contract
- Canonical command name: `mac-storage-scout`
- Canonical binary path inside this repository: `./bin/mac-storage-scout`
- Short alias `mss` is intentionally not created by default to avoid global name collisions.

## Install From Source (Deterministic)
Run from the repository root (whatever path the agent cloned it to):
```bash
go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout
```

Optional user-level install onto `PATH`:
```bash
mkdir -p "$HOME/.local/bin"
install -m 0755 ./bin/mac-storage-scout "$HOME/.local/bin/mac-storage-scout"
# ensure $HOME/.local/bin is on PATH, then `mac-storage-scout` resolves globally
```

## Invocation Convention
- Before user-level install → call as `./bin/mac-storage-scout ...` from the repo root.
- After user-level install (PATH) → call as `mac-storage-scout ...` from anywhere.

The examples below use `./bin/mac-storage-scout` so they work immediately after the build step.

## Verify
```bash
./bin/mac-storage-scout scan --threshold 500MB --top 5 --profile macos-core --no-progress
```

When the reported free space is disappearing or targeted scans do not add up, audit the complete writable macOS volume first:
```bash
./bin/mac-storage-scout audit --threshold 5GB --top 10 --no-progress
```
Treat `unaccounted` as an explicit investigation result. It can include unreadable paths, APFS snapshots/shared-volume accounting, clones, reserved or purgeable space, and deleted files still held open.

## Safety Contract
Always plan a delete first; only commit after the plan is reviewed.

```bash
# 1. plan only — no filesystem changes
./bin/mac-storage-scout delete --dry-run <path> [path...]

# 2. commit — requires explicit --yes
./bin/mac-storage-scout delete --yes <path> [path...]
```

Protected roots are refused even with `--yes`: `/`, `/System`, `/usr`, `/bin`, `/sbin`, `/private/var/vm`.

## Minimal Agent Workflow
1. Build the binary from source (`go build ...`).
2. Run `audit` for a whole-volume discrepancy; otherwise scan with `--profile macos-core` (or a specific path).
3. Propose candidate cleanup paths from the report.
4. Run `delete --dry-run` and surface the planned freed bytes.
5. Run `delete --yes` only after explicit user confirmation.

## Worked Example (Depersonalized)
```bash
# build
go build -o ./bin/mac-storage-scout ./cmd/mac-storage-scout

# scan user caches with a 200MB threshold
./bin/mac-storage-scout scan --threshold 200MB --top 5 --no-progress "$HOME/Library/Caches"

# pick two candidate buckets from the report, plan the delete
./bin/mac-storage-scout delete --dry-run \
  "$HOME/Library/Caches/com.example.Browser" \
  "$HOME/Library/Caches/Homebrew"

# after confirmation, commit
./bin/mac-storage-scout delete --yes \
  "$HOME/Library/Caches/com.example.Browser" \
  "$HOME/Library/Caches/Homebrew"
```
