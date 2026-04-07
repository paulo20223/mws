# mws — Multi-task Workspace Manager

[![CI](https://github.com/paulo20223/mws/actions/workflows/ci.yml/badge.svg)](https://github.com/paulo20223/mws/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/paulo20223/mws)](https://github.com/paulo20223/mws/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/paulo20223/mws)](https://goreportcard.com/report/github.com/paulo20223/mws)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/paulo20223/mws)](https://go.dev/)
[![Downloads](https://img.shields.io/github/downloads/paulo20223/mws/total)](https://github.com/paulo20223/mws/releases)

**English** | [中文](README_zh.md) | [Español](README_es.md) | [日本語](README_ja.md) | [한국어](README_ko.md) | [Português](README_pt.md) | [Deutsch](README_de.md)

---

Work on multiple tasks simultaneously across different repos and branches. Each task gets a full, isolated copy of the entire workspace — all repos, configs, compose files, shared data — completely independent.

## Why

`git worktree` only works within a single repo. When your project spans multiple repos plus shared files (docker-compose, configs, data), switching between tasks means juggling branches across repos and hoping shared files don't conflict. `mws` solves this by giving each task its own complete workspace copy.

## Install

### Quick install (recommended)

```bash
curl -fsSL https://raw.githubusercontent.com/paulo20223/mws/main/install.sh | sh
```

### Homebrew (macOS / Linux)

```bash
brew install paulo20223/tap/mws
```

### Debian / Ubuntu

```bash
# amd64
curl -fLo mws.deb https://github.com/paulo20223/mws/releases/latest/download/mws_amd64.deb
sudo dpkg -i mws.deb

# arm64
curl -fLo mws.deb https://github.com/paulo20223/mws/releases/latest/download/mws_arm64.deb
sudo dpkg -i mws.deb
```

### From source (requires Go 1.22+)

```bash
go install github.com/paulo20223/mws/cmd/mws@latest
```

### Manual download

Download the binary for your platform from [Releases](https://github.com/paulo20223/mws/releases/latest), then:

```bash
chmod +x mws-*
mv mws-* /usr/local/bin/mws
```

## Quick start

```bash
cd /path/to/your/workspace    # directory with multiple repos
mws init                      # initialize (once)

mws create feature-x -b feature/x -r "backend*,frontend"
# creates /path/to/your/workspace--feature-x/ with full copy,
# switches matching repos to branch feature/x

mws create bugfix-y -b fix/y
# creates another independent copy

mws list                      # see all tasks
mws status feature-x          # repo branches + dirty state
cd $(mws cd feature-x)        # jump into task workspace
mws sync feature-x            # pull non-repo updates from base
mws delete bugfix-y --force   # done, remove copy
```

## Commands

| Command | Description |
|---|---|
| `mws init` | Initialize workspace, discover all git repos |
| `mws create <name> [-b branch] [-r globs]` | Full workspace copy, switch branches |
| `mws list` | List all tasks with sizes and branches |
| `mws status [name]` | Show repo branches, modified/staged/untracked |
| `mws delete <name> [--force]` | Remove task copy and metadata |
| `mws cd <name>` | Print task path (use: `cd $(mws cd name)`) |
| `mws open <name>` | Open task workspace in editor |
| `mws base` | Print base workspace path |
| `mws sync <name> [--dry-run]` | Sync non-repo files from base |

## Shell integration

Add to `~/.zshrc`:

```bash
# quick navigation
mcd() { cd "$(mws cd "$1")" }

# show current task in prompt
mws_prompt() {
  local f="$PWD"
  while [[ "$f" != "/" ]]; do
    [[ -f "$f/.mws-task" ]] && { echo "[$(cat "$f/.mws-task")] "; return; }
    f="$(dirname "$f")"
  done
}
PROMPT='$(mws_prompt)'$PROMPT
```

## How it works

```
/path/to/workspace/                   # base (original)
├── repo-a/
├── repo-b/
├── docker-compose.yml
├── shared-config/
└── .mws/                            # metadata
    ├── config.toml                   # exclude patterns, editor
    └── tasks/
        └── feature-x/
            └── manifest.toml         # task definition

/path/to/workspace--feature-x/       # full isolated copy
├── repo-a/                          # on branch feature/x
├── repo-b/                          # on branch feature/x
├── docker-compose.yml               # own copy, edit freely
├── shared-config/                   # own copy
└── .mws-task                        # marker file
```

Excluded from copy (configurable in `.mws/config.toml`): `node_modules/`, `.venv/`, `__pycache__/`, `.next/`, `dist/`, `build/`, `vendor/`, `target/`.

## Config

`.mws/config.toml` is created by `mws init` with sensible defaults:

```toml
[workspace]
base = "/path/to/workspace"
editor = "code"

[copy]
exclude = [
    "node_modules/",
    ".venv/",
    "__pycache__/",
    ".next/",
    "dist/",
    "build/",
    ".mws/",
    "*.pyc",
    ".DS_Store",
    "plans/",
    ".git/lfs/",
    "vendor/",
    "target/",
]
```

## License

MIT
