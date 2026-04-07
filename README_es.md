# mws — Gestor de Espacios de Trabajo Multi-tarea

[![CI](https://github.com/paulo20223/mws/actions/workflows/ci.yml/badge.svg)](https://github.com/paulo20223/mws/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/paulo20223/mws)](https://github.com/paulo20223/mws/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/paulo20223/mws)](https://goreportcard.com/report/github.com/paulo20223/mws)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/paulo20223/mws)](https://go.dev/)
[![Downloads](https://img.shields.io/github/downloads/paulo20223/mws/total)](https://github.com/paulo20223/mws/releases)

[English](README.md) | [中文](README_zh.md) | **Español** | [日本語](README_ja.md) | [한국어](README_ko.md) | [Português](README_pt.md) | [Deutsch](README_de.md)

---

Trabaja en varias tareas simultáneamente a través de diferentes repos y ramas. Cada tarea obtiene una copia completa y aislada de todo el espacio de trabajo: todos los repos, configs, archivos compose, datos compartidos, completamente independiente.

## Por qué

`git worktree` solo funciona dentro de un único repositorio. Cuando tu proyecto abarca múltiples repos más archivos compartidos (docker-compose, configs, datos), cambiar entre tareas significa hacer malabarismos con ramas en varios repos y esperar que los archivos compartidos no entren en conflicto. `mws` resuelve esto dándole a cada tarea su propia copia completa del espacio de trabajo.

## Instalación

### Instalación rápida (recomendada)

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

### Desde el código fuente (requiere Go 1.22+)

```bash
go install github.com/paulo20223/mws/cmd/mws@latest
```

### Descarga manual

Descarga el binario para tu plataforma desde [Releases](https://github.com/paulo20223/mws/releases/latest), luego:

```bash
chmod +x mws-*
mv mws-* /usr/local/bin/mws
```

## Inicio rápido

```bash
cd /path/to/your/workspace    # directorio con múltiples repos
mws init                      # inicializar (una vez)

mws create feature-x -b feature/x -r "backend*,frontend"
# crea /path/to/your/workspace--feature-x/ con copia completa,
# cambia los repos coincidentes a la rama feature/x

mws create bugfix-y -b fix/y
# crea otra copia independiente

mws list                      # ver todas las tareas
mws status feature-x          # ramas de repos + estado de cambios
cd $(mws cd feature-x)        # ir al espacio de trabajo de la tarea
mws sync feature-x            # obtener actualizaciones no-repo del base
mws delete bugfix-y --force   # listo, eliminar copia
```

## Comandos

| Comando | Descripción |
|---|---|
| `mws init` | Inicializar espacio de trabajo, descubrir todos los repos git |
| `mws create <name> [-b branch] [-r globs]` | Copia completa del espacio de trabajo, cambiar ramas |
| `mws list` | Listar todas las tareas con tamaños y ramas |
| `mws status [name]` | Mostrar ramas de repos, modificados/staged/sin seguimiento |
| `mws delete <name> [--force]` | Eliminar copia de la tarea y metadatos |
| `mws cd <name>` | Imprimir ruta de la tarea (usar: `cd $(mws cd name)`) |
| `mws open <name>` | Abrir espacio de trabajo de la tarea en el editor |
| `mws base` | Imprimir ruta del espacio de trabajo base |
| `mws sync <name> [--dry-run]` | Sincronizar archivos no-repo desde el base |

## Integración con Shell

Agrega a `~/.zshrc`:

```bash
# navegación rápida
mcd() { cd "$(mws cd "$1")" }

# mostrar tarea actual en el prompt
mws_prompt() {
  local f="$PWD"
  while [[ "$f" != "/" ]]; do
    [[ -f "$f/.mws-task" ]] && { echo "[$(cat "$f/.mws-task")] "; return; }
    f="$(dirname "$f")"
  done
}
PROMPT='$(mws_prompt)'$PROMPT
```

## Cómo funciona

```
/path/to/workspace/                   # base (original)
├── repo-a/
├── repo-b/
├── docker-compose.yml
├── shared-config/
└── .mws/                            # metadatos
    ├── config.toml                   # patrones de exclusión, editor
    └── tasks/
        └── feature-x/
            └── manifest.toml         # definición de tarea

/path/to/workspace--feature-x/       # copia completamente aislada
├── repo-a/                          # en rama feature/x
├── repo-b/                          # en rama feature/x
├── docker-compose.yml               # copia propia, edita libremente
├── shared-config/                   # copia propia
└── .mws-task                        # archivo marcador
```

Excluidos de la copia (configurable en `.mws/config.toml`): `node_modules/`, `.venv/`, `__pycache__/`, `.next/`, `dist/`, `build/`, `vendor/`, `target/`.

## Configuración

`.mws/config.toml` se crea con `mws init` con valores predeterminados razonables:

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

## Licencia

MIT
