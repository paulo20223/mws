# mws — Gerenciador de Workspace Multi-tarefa

[![CI](https://github.com/paulo20223/mws/actions/workflows/ci.yml/badge.svg)](https://github.com/paulo20223/mws/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/paulo20223/mws)](https://github.com/paulo20223/mws/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/paulo20223/mws)](https://goreportcard.com/report/github.com/paulo20223/mws)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/paulo20223/mws)](https://go.dev/)
[![Downloads](https://img.shields.io/github/downloads/paulo20223/mws/total)](https://github.com/paulo20223/mws/releases)

[English](README.md) | [中文](README_zh.md) | [Español](README_es.md) | [日本語](README_ja.md) | [한국어](README_ko.md) | **Português** | [Deutsch](README_de.md)

---

Trabalhe em várias tarefas simultaneamente em diferentes repos e branches. Cada tarefa recebe uma cópia completa e isolada de todo o workspace — todos os repos, configs, arquivos compose, dados compartilhados — completamente independente.

## Por quê

`git worktree` funciona apenas dentro de um único repositório. Quando seu projeto abrange múltiplos repos e arquivos compartilhados (docker-compose, configs, dados), alternar entre tarefas significa fazer malabarismo com branches em vários repos e torcer para que os arquivos compartilhados não entrem em conflito. O `mws` resolve isso dando a cada tarefa sua própria cópia completa do workspace.

## Instalação

### Instalação rápida (recomendada)

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

### A partir do código-fonte (requer Go 1.22+)

```bash
go install github.com/paulo20223/mws/cmd/mws@latest
```

### Download manual

Baixe o binário para sua plataforma em [Releases](https://github.com/paulo20223/mws/releases/latest) e execute:

```bash
chmod +x mws-*
mv mws-* /usr/local/bin/mws
```

## Início rápido

```bash
cd /path/to/your/workspace    # diretório com múltiplos repos
mws init                      # inicializar (uma vez)

mws create feature-x -b feature/x -r "backend*,frontend"
# cria /path/to/your/workspace--feature-x/ com cópia completa,
# muda os repos correspondentes para o branch feature/x

mws create bugfix-y -b fix/y
# cria outra cópia independente

mws list                      # ver todas as tarefas
mws status feature-x          # branches dos repos + estado de alterações
cd $(mws cd feature-x)        # ir para o workspace da tarefa
mws sync feature-x            # buscar atualizações não-repo do base
mws delete bugfix-y --force   # pronto, remover cópia
```

## Comandos

| Comando | Descrição |
|---|---|
| `mws init` | Inicializar workspace, descobrir todos os repos git |
| `mws create <name> [-b branch] [-r globs]` | Cópia completa do workspace, trocar branches |
| `mws list` | Listar todas as tarefas com tamanhos e branches |
| `mws status [name]` | Mostrar branches dos repos, modificados/staged/não rastreados |
| `mws delete <name> [--force]` | Remover cópia da tarefa e metadados |
| `mws cd <name>` | Imprimir caminho da tarefa (usar: `cd $(mws cd name)`) |
| `mws open <name>` | Abrir workspace da tarefa no editor |
| `mws base` | Imprimir caminho do workspace base |
| `mws sync <name> [--dry-run]` | Sincronizar arquivos não-repo do base |

## Integração com Shell

Adicione ao `~/.zshrc`:

```bash
# navegação rápida
mcd() { cd "$(mws cd "$1")" }

# mostrar tarefa atual no prompt
mws_prompt() {
  local f="$PWD"
  while [[ "$f" != "/" ]]; do
    [[ -f "$f/.mws-task" ]] && { echo "[$(cat "$f/.mws-task")] "; return; }
    f="$(dirname "$f")"
  done
}
PROMPT='$(mws_prompt)'$PROMPT
```

## Como funciona

```
/path/to/workspace/                   # base (original)
├── repo-a/
├── repo-b/
├── docker-compose.yml
├── shared-config/
└── .mws/                            # metadados
    ├── config.toml                   # padrões de exclusão, editor
    └── tasks/
        └── feature-x/
            └── manifest.toml         # definição da tarefa

/path/to/workspace--feature-x/       # cópia completamente isolada
├── repo-a/                          # no branch feature/x
├── repo-b/                          # no branch feature/x
├── docker-compose.yml               # cópia própria, edite livremente
├── shared-config/                   # cópia própria
└── .mws-task                        # arquivo marcador
```

Excluídos da cópia (configurável em `.mws/config.toml`): `node_modules/`, `.venv/`, `__pycache__/`, `.next/`, `dist/`, `build/`, `vendor/`, `target/`.

## Configuração

`.mws/config.toml` é criado pelo `mws init` com valores padrão sensatos:

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

## Licença

MIT
