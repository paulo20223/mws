# mws — Multi-Task Workspace Manager

[![CI](https://github.com/paulo20223/mws/actions/workflows/ci.yml/badge.svg)](https://github.com/paulo20223/mws/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/paulo20223/mws)](https://github.com/paulo20223/mws/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/paulo20223/mws)](https://goreportcard.com/report/github.com/paulo20223/mws)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/paulo20223/mws)](https://go.dev/)
[![Downloads](https://img.shields.io/github/downloads/paulo20223/mws/total)](https://github.com/paulo20223/mws/releases)

[English](README.md) | [中文](README_zh.md) | [Español](README_es.md) | [日本語](README_ja.md) | [한국어](README_ko.md) | [Português](README_pt.md) | **Deutsch**

---

Arbeiten Sie gleichzeitig an mehreren Aufgaben in verschiedenen Repos und Branches. Jede Aufgabe erhält eine vollständige, isolierte Kopie des gesamten Workspace — alle Repos, Konfigurationen, Compose-Dateien, gemeinsame Daten — vollständig unabhängig.

## Warum

`git worktree` funktioniert nur innerhalb eines einzelnen Repos. Wenn Ihr Projekt mehrere Repos und gemeinsame Dateien (docker-compose, Konfigurationen, Daten) umfasst, bedeutet das Wechseln zwischen Aufgaben, Branches in verschiedenen Repos zu jonglieren und zu hoffen, dass gemeinsame Dateien nicht kollidieren. `mws` löst dieses Problem, indem es jeder Aufgabe eine eigene vollständige Workspace-Kopie gibt.

## Installation

### Schnellinstallation (empfohlen)

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

### Aus dem Quellcode (erfordert Go 1.22+)

```bash
go install github.com/paulo20223/mws/cmd/mws@latest
```

### Manueller Download

Laden Sie die Binärdatei für Ihre Plattform von [Releases](https://github.com/paulo20223/mws/releases/latest) herunter und führen Sie aus:

```bash
chmod +x mws-*
mv mws-* /usr/local/bin/mws
```

## Schnellstart

```bash
cd /path/to/your/workspace    # Verzeichnis mit mehreren Repos
mws init                      # Initialisieren (einmalig)

mws create feature-x -b feature/x -r "backend*,frontend"
# erstellt /path/to/your/workspace--feature-x/ mit vollständiger Kopie,
# wechselt passende Repos zum Branch feature/x

mws create bugfix-y -b fix/y
# erstellt eine weitere unabhängige Kopie

mws list                      # alle Aufgaben anzeigen
mws status feature-x          # Repo-Branches + Änderungsstatus
cd $(mws cd feature-x)        # zum Aufgaben-Workspace wechseln
mws sync feature-x            # Nicht-Repo-Updates vom Base holen
mws delete bugfix-y --force   # fertig, Kopie entfernen
```

## Befehle

| Befehl | Beschreibung |
|---|---|
| `mws init` | Workspace initialisieren, alle Git-Repos erkennen |
| `mws create <name> [-b branch] [-r globs]` | Vollständige Workspace-Kopie, Branches wechseln |
| `mws list` | Alle Aufgaben mit Größen und Branches auflisten |
| `mws status [name]` | Repo-Branches, geändert/staged/unverfolgt anzeigen |
| `mws delete <name> [--force]` | Aufgabenkopie und Metadaten entfernen |
| `mws cd <name>` | Aufgabenpfad ausgeben (Verwendung: `cd $(mws cd name)`) |
| `mws open <name>` | Aufgaben-Workspace im Editor öffnen |
| `mws base` | Basis-Workspace-Pfad ausgeben |
| `mws sync <name> [--dry-run]` | Nicht-Repo-Dateien vom Base synchronisieren |

## Shell-Integration

Zu `~/.zshrc` hinzufügen:

```bash
# schnelle Navigation
mcd() { cd "$(mws cd "$1")" }

# aktuelle Aufgabe im Prompt anzeigen
mws_prompt() {
  local f="$PWD"
  while [[ "$f" != "/" ]]; do
    [[ -f "$f/.mws-task" ]] && { echo "[$(cat "$f/.mws-task")] "; return; }
    f="$(dirname "$f")"
  done
}
PROMPT='$(mws_prompt)'$PROMPT
```

## Funktionsweise

```
/path/to/workspace/                   # Base (Original)
├── repo-a/
├── repo-b/
├── docker-compose.yml
├── shared-config/
└── .mws/                            # Metadaten
    ├── config.toml                   # Ausschlussmuster, Editor
    └── tasks/
        └── feature-x/
            └── manifest.toml         # Aufgabendefinition

/path/to/workspace--feature-x/       # vollständig isolierte Kopie
├── repo-a/                          # auf Branch feature/x
├── repo-b/                          # auf Branch feature/x
├── docker-compose.yml               # eigene Kopie, frei bearbeitbar
├── shared-config/                   # eigene Kopie
└── .mws-task                        # Markierungsdatei
```

Von der Kopie ausgeschlossen (konfigurierbar in `.mws/config.toml`): `node_modules/`, `.venv/`, `__pycache__/`, `.next/`, `dist/`, `build/`, `vendor/`, `target/`.

## Konfiguration

`.mws/config.toml` wird von `mws init` mit sinnvollen Standardwerten erstellt:

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

## Lizenz

MIT
