# mws — 多任务工作区管理器

[![CI](https://github.com/paulo20223/mws/actions/workflows/ci.yml/badge.svg)](https://github.com/paulo20223/mws/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/paulo20223/mws)](https://github.com/paulo20223/mws/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/paulo20223/mws)](https://goreportcard.com/report/github.com/paulo20223/mws)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/paulo20223/mws)](https://go.dev/)
[![Downloads](https://img.shields.io/github/downloads/paulo20223/mws/total)](https://github.com/paulo20223/mws/releases)

[English](README.md) | **中文** | [Español](README_es.md) | [日本語](README_ja.md) | [한국어](README_ko.md) | [Português](README_pt.md) | [Deutsch](README_de.md)

---

同时在多个仓库和分支上处理多个任务。每个任务都会获得整个工作区的完整隔离副本——所有仓库、配置文件、compose 文件、共享数据——完全独立。

## 为什么需要它

`git worktree` 只能在单个仓库内工作。当你的项目跨越多个仓库以及共享文件（docker-compose、配置、数据）时，在任务之间切换意味着需要在各个仓库中切换分支，并祈祷共享文件不会冲突。`mws` 通过为每个任务提供完整的工作区副本来解决这个问题。

## 安装

### 快速安装（推荐）

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

### 从源码构建（需要 Go 1.22+）

```bash
go install github.com/paulo20223/mws/cmd/mws@latest
```

### 手动下载

从 [Releases](https://github.com/paulo20223/mws/releases/latest) 下载适合你平台的二进制文件，然后：

```bash
chmod +x mws-*
mv mws-* /usr/local/bin/mws
```

## 快速上手

```bash
cd /path/to/your/workspace    # 包含多个仓库的目录
mws init                      # 初始化（只需一次）

mws create feature-x -b feature/x -r "backend*,frontend"
# 创建 /path/to/your/workspace--feature-x/ 完整副本，
# 并将匹配的仓库切换到 feature/x 分支

mws create bugfix-y -b fix/y
# 创建另一个独立副本

mws list                      # 查看所有任务
mws status feature-x          # 仓库分支和修改状态
cd $(mws cd feature-x)        # 跳转到任务工作区
mws sync feature-x            # 从基础工作区同步非仓库更新
mws delete bugfix-y --force   # 完成，删除副本
```

## 命令

| 命令 | 描述 |
|---|---|
| `mws init` | 初始化工作区，发现所有 git 仓库 |
| `mws create <name> [-b branch] [-r globs]` | 完整工作区复制，切换分支 |
| `mws list` | 列出所有任务及大小和分支 |
| `mws status [name]` | 显示仓库分支、已修改/已暂存/未跟踪 |
| `mws delete <name> [--force]` | 删除任务副本和元数据 |
| `mws cd <name>` | 输出任务路径（使用：`cd $(mws cd name)`） |
| `mws open <name>` | 在编辑器中打开任务工作区 |
| `mws base` | 输出基础工作区路径 |
| `mws sync <name> [--dry-run]` | 从基础工作区同步非仓库文件 |

## Shell 集成

添加到 `~/.zshrc`：

```bash
# 快速导航
mcd() { cd "$(mws cd "$1")" }

# 在提示符中显示当前任务
mws_prompt() {
  local f="$PWD"
  while [[ "$f" != "/" ]]; do
    [[ -f "$f/.mws-task" ]] && { echo "[$(cat "$f/.mws-task")] "; return; }
    f="$(dirname "$f")"
  done
}
PROMPT='$(mws_prompt)'$PROMPT
```

## 工作原理

```
/path/to/workspace/                   # 基础（原始）
├── repo-a/
├── repo-b/
├── docker-compose.yml
├── shared-config/
└── .mws/                            # 元数据
    ├── config.toml                   # 排除规则、编辑器
    └── tasks/
        └── feature-x/
            └── manifest.toml         # 任务定义

/path/to/workspace--feature-x/       # 完全隔离的副本
├── repo-a/                          # 在 feature/x 分支
├── repo-b/                          # 在 feature/x 分支
├── docker-compose.yml               # 自己的副本，随意编辑
├── shared-config/                   # 自己的副本
└── .mws-task                        # 标记文件
```

复制时排除的内容（可在 `.mws/config.toml` 中配置）：`node_modules/`、`.venv/`、`__pycache__/`、`.next/`、`dist/`、`build/`、`vendor/`、`target/`。

## 配置

`mws init` 会创建 `.mws/config.toml`，包含合理的默认值：

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

## 许可证

MIT
