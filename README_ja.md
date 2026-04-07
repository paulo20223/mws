# mws — マルチタスクワークスペースマネージャー

[![CI](https://github.com/paulo20223/mws/actions/workflows/ci.yml/badge.svg)](https://github.com/paulo20223/mws/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/paulo20223/mws)](https://github.com/paulo20223/mws/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/paulo20223/mws)](https://goreportcard.com/report/github.com/paulo20223/mws)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/paulo20223/mws)](https://go.dev/)
[![Downloads](https://img.shields.io/github/downloads/paulo20223/mws/total)](https://github.com/paulo20223/mws/releases)

[English](README.md) | [中文](README_zh.md) | [Español](README_es.md) | **日本語** | [한국어](README_ko.md) | [Português](README_pt.md) | [Deutsch](README_de.md)

---

異なるリポジトリやブランチにまたがる複数のタスクを同時に作業できます。各タスクはワークスペース全体の完全に分離されたコピーを取得します。すべてのリポジトリ、設定ファイル、composeファイル、共有データが完全に独立しています。

## なぜ必要か

`git worktree` は単一リポジトリ内でしか機能しません。プロジェクトが複数のリポジトリと共有ファイル（docker-compose、設定、データ）にまたがる場合、タスク間の切り替えはリポジトリ間でブランチを切り替え、共有ファイルが競合しないことを祈ることを意味します。`mws` は各タスクに完全なワークスペースのコピーを提供することでこの問題を解決します。

## インストール

### クイックインストール（推奨）

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

### ソースからビルド（Go 1.22+ が必要）

```bash
go install github.com/paulo20223/mws/cmd/mws@latest
```

### 手動ダウンロード

[Releases](https://github.com/paulo20223/mws/releases/latest) からプラットフォーム用のバイナリをダウンロードし、次のコマンドを実行します：

```bash
chmod +x mws-*
mv mws-* /usr/local/bin/mws
```

## クイックスタート

```bash
cd /path/to/your/workspace    # 複数のリポジトリがあるディレクトリ
mws init                      # 初期化（一度だけ）

mws create feature-x -b feature/x -r "backend*,frontend"
# /path/to/your/workspace--feature-x/ に完全コピーを作成し、
# 一致するリポジトリを feature/x ブランチに切り替え

mws create bugfix-y -b fix/y
# 別の独立したコピーを作成

mws list                      # すべてのタスクを表示
mws status feature-x          # リポジトリのブランチとダーティ状態
cd $(mws cd feature-x)        # タスクワークスペースへ移動
mws sync feature-x            # ベースからリポジトリ以外の更新を取得
mws delete bugfix-y --force   # 完了、コピーを削除
```

## コマンド

| コマンド | 説明 |
|---|---|
| `mws init` | ワークスペースを初期化し、すべてのgitリポジトリを検出 |
| `mws create <name> [-b branch] [-r globs]` | ワークスペースの完全コピー、ブランチ切り替え |
| `mws list` | サイズとブランチ付きで全タスクを一覧表示 |
| `mws status [name]` | リポジトリのブランチ、変更/ステージング/未追跡を表示 |
| `mws delete <name> [--force]` | タスクのコピーとメタデータを削除 |
| `mws cd <name>` | タスクのパスを出力（使い方：`cd $(mws cd name)`） |
| `mws open <name>` | エディタでタスクワークスペースを開く |
| `mws base` | ベースワークスペースのパスを出力 |
| `mws sync <name> [--dry-run]` | ベースからリポジトリ以外のファイルを同期 |

## シェル統合

`~/.zshrc` に追加：

```bash
# クイックナビゲーション
mcd() { cd "$(mws cd "$1")" }

# プロンプトに現在のタスクを表示
mws_prompt() {
  local f="$PWD"
  while [[ "$f" != "/" ]]; do
    [[ -f "$f/.mws-task" ]] && { echo "[$(cat "$f/.mws-task")] "; return; }
    f="$(dirname "$f")"
  done
}
PROMPT='$(mws_prompt)'$PROMPT
```

## 仕組み

```
/path/to/workspace/                   # ベース（オリジナル）
├── repo-a/
├── repo-b/
├── docker-compose.yml
├── shared-config/
└── .mws/                            # メタデータ
    ├── config.toml                   # 除外パターン、エディタ
    └── tasks/
        └── feature-x/
            └── manifest.toml         # タスク定義

/path/to/workspace--feature-x/       # 完全に分離されたコピー
├── repo-a/                          # feature/x ブランチ上
├── repo-b/                          # feature/x ブランチ上
├── docker-compose.yml               # 自分のコピー、自由に編集
├── shared-config/                   # 自分のコピー
└── .mws-task                        # マーカーファイル
```

コピーから除外されるもの（`.mws/config.toml` で設定可能）：`node_modules/`、`.venv/`、`__pycache__/`、`.next/`、`dist/`、`build/`、`vendor/`、`target/`。

## 設定

`mws init` により `.mws/config.toml` が適切なデフォルト値で作成されます：

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

## ライセンス

MIT
