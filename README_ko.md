# mws — 멀티태스크 워크스페이스 매니저

[![CI](https://github.com/paulo20223/mws/actions/workflows/ci.yml/badge.svg)](https://github.com/paulo20223/mws/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/paulo20223/mws)](https://github.com/paulo20223/mws/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/paulo20223/mws)](https://goreportcard.com/report/github.com/paulo20223/mws)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/paulo20223/mws)](https://go.dev/)
[![Downloads](https://img.shields.io/github/downloads/paulo20223/mws/total)](https://github.com/paulo20223/mws/releases)

[English](README.md) | [中文](README_zh.md) | [Español](README_es.md) | [日本語](README_ja.md) | **한국어** | [Português](README_pt.md) | [Deutsch](README_de.md)

---

여러 저장소와 브랜치에 걸쳐 여러 작업을 동시에 수행할 수 있습니다. 각 작업은 전체 워크스페이스의 완전히 격리된 복사본을 받습니다. 모든 저장소, 설정 파일, compose 파일, 공유 데이터가 완전히 독립적입니다.

## 왜 필요한가

`git worktree`는 단일 저장소 내에서만 동작합니다. 프로젝트가 여러 저장소와 공유 파일(docker-compose, 설정, 데이터)에 걸쳐 있을 때, 작업 간 전환은 여러 저장소에서 브랜치를 전환하고 공유 파일이 충돌하지 않기를 바라는 것을 의미합니다. `mws`는 각 작업에 완전한 워크스페이스 복사본을 제공하여 이 문제를 해결합니다.

## 설치

### 빠른 설치 (권장)

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

### 소스에서 빌드 (Go 1.22+ 필요)

```bash
go install github.com/paulo20223/mws/cmd/mws@latest
```

### 수동 다운로드

[Releases](https://github.com/paulo20223/mws/releases/latest)에서 플랫폼에 맞는 바이너리를 다운로드한 후:

```bash
chmod +x mws-*
mv mws-* /usr/local/bin/mws
```

## 빠른 시작

```bash
cd /path/to/your/workspace    # 여러 저장소가 있는 디렉토리
mws init                      # 초기화 (한 번만)

mws create feature-x -b feature/x -r "backend*,frontend"
# /path/to/your/workspace--feature-x/ 에 전체 복사본을 생성하고,
# 일치하는 저장소를 feature/x 브랜치로 전환

mws create bugfix-y -b fix/y
# 또 다른 독립적인 복사본 생성

mws list                      # 모든 작업 보기
mws status feature-x          # 저장소 브랜치 + 변경 상태
cd $(mws cd feature-x)        # 작업 워크스페이스로 이동
mws sync feature-x            # 기본 워크스페이스에서 비-저장소 업데이트 가져오기
mws delete bugfix-y --force   # 완료, 복사본 삭제
```

## 명령어

| 명령어 | 설명 |
|---|---|
| `mws init` | 워크스페이스 초기화, 모든 git 저장소 탐색 |
| `mws create <name> [-b branch] [-r globs]` | 전체 워크스페이스 복사, 브랜치 전환 |
| `mws list` | 크기와 브랜치와 함께 모든 작업 나열 |
| `mws status [name]` | 저장소 브랜치, 수정/스테이징/미추적 표시 |
| `mws delete <name> [--force]` | 작업 복사본과 메타데이터 삭제 |
| `mws cd <name>` | 작업 경로 출력 (사용법: `cd $(mws cd name)`) |
| `mws open <name>` | 편집기에서 작업 워크스페이스 열기 |
| `mws base` | 기본 워크스페이스 경로 출력 |
| `mws sync <name> [--dry-run]` | 기본에서 비-저장소 파일 동기화 |

## 셸 통합

`~/.zshrc`에 추가:

```bash
# 빠른 이동
mcd() { cd "$(mws cd "$1")" }

# 프롬프트에 현재 작업 표시
mws_prompt() {
  local f="$PWD"
  while [[ "$f" != "/" ]]; do
    [[ -f "$f/.mws-task" ]] && { echo "[$(cat "$f/.mws-task")] "; return; }
    f="$(dirname "$f")"
  done
}
PROMPT='$(mws_prompt)'$PROMPT
```

## 작동 원리

```
/path/to/workspace/                   # 기본 (원본)
├── repo-a/
├── repo-b/
├── docker-compose.yml
├── shared-config/
└── .mws/                            # 메타데이터
    ├── config.toml                   # 제외 패턴, 편집기
    └── tasks/
        └── feature-x/
            └── manifest.toml         # 작업 정의

/path/to/workspace--feature-x/       # 완전히 격리된 복사본
├── repo-a/                          # feature/x 브랜치
├── repo-b/                          # feature/x 브랜치
├── docker-compose.yml               # 자체 복사본, 자유롭게 편집
├── shared-config/                   # 자체 복사본
└── .mws-task                        # 마커 파일
```

복사에서 제외되는 항목 (`.mws/config.toml`에서 설정 가능): `node_modules/`, `.venv/`, `__pycache__/`, `.next/`, `dist/`, `build/`, `vendor/`, `target/`.

## 설정

`mws init`으로 `.mws/config.toml`이 합리적인 기본값으로 생성됩니다:

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

## 라이선스

MIT
