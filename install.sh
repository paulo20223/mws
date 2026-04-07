#!/bin/sh
# mws installer — https://github.com/paulo20223/mws
# Usage: curl -fsSL https://raw.githubusercontent.com/paulo20223/mws/main/install.sh | sh
set -e

REPO="paulo20223/mws"
BINARY="mws"

# ── detect platform ──────────────────────────────────────────────────────────

detect_os() {
  case "$(uname -s)" in
    Darwin*)  echo "darwin" ;;
    Linux*)   echo "linux" ;;
    *)        echo "unsupported"; return 1 ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)   echo "amd64" ;;
    arm64|aarch64)   echo "arm64" ;;
    *)               echo "unsupported"; return 1 ;;
  esac
}

OS="$(detect_os)"
ARCH="$(detect_arch)"

if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
  echo "Error: unsupported platform $(uname -s)/$(uname -m)"
  exit 1
fi

echo "Detected platform: ${OS}/${ARCH}"

# ── resolve latest version ───────────────────────────────────────────────────

if command -v curl >/dev/null 2>&1; then
  FETCH="curl -fsSL"
elif command -v wget >/dev/null 2>&1; then
  FETCH="wget -qO-"
else
  echo "Error: curl or wget required"
  exit 1
fi

VERSION="$($FETCH "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//')"
if [ -z "$VERSION" ]; then
  echo "Warning: could not detect latest version, using v0.2.0"
  VERSION="v0.2.0"
fi

echo "Latest version: ${VERSION}"

# ── download binary ──────────────────────────────────────────────────────────

ASSET="${BINARY}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

TMPFILE="$(mktemp)"
echo "Downloading ${URL} ..."

if command -v curl >/dev/null 2>&1; then
  curl -fSL -o "$TMPFILE" "$URL"
else
  wget -q -O "$TMPFILE" "$URL"
fi

chmod +x "$TMPFILE"

# ── shell integration helper ─────────────────────────────────────────────────

MARKER_BEGIN="# >>> mws initialize >>>"
MARKER_END="# <<< mws initialize <<<"

setup_shell() {
  MWS_BIN="$1"

  # detect rc file
  case "$SHELL" in
    */zsh)  RC_FILE="$HOME/.zshrc" ;;
    */bash) RC_FILE="$HOME/.bashrc" ;;
    *)
      if [ -f "$HOME/.zshrc" ]; then
        RC_FILE="$HOME/.zshrc"
      elif [ -f "$HOME/.bashrc" ]; then
        RC_FILE="$HOME/.bashrc"
      else
        RC_FILE=""
      fi
      ;;
  esac

  if [ -z "$RC_FILE" ]; then
    echo "Could not detect shell rc file. Add manually:"
    echo "  eval \"\$(mws shell-init)\""
    return
  fi

  # idempotent: skip if already present
  if [ -f "$RC_FILE" ] && grep -qF "$MARKER_BEGIN" "$RC_FILE"; then
    echo "Shell integration already configured in $RC_FILE"
    return
  fi

  printf '\n%s\neval "$(%s shell-init)"\n%s\n' "$MARKER_BEGIN" "$MWS_BIN" "$MARKER_END" >> "$RC_FILE"
  echo "Shell integration added to $RC_FILE"
}

# ── install ──────────────────────────────────────────────────────────────────

INSTALL_DIR=""

# prefer ~/.local/bin (no sudo needed)
if [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
  INSTALL_DIR="$HOME/.local/bin"
# fallback: /usr/local/bin
elif [ -w "/usr/local/bin" ]; then
  INSTALL_DIR="/usr/local/bin"
else
  # try with sudo
  echo "Need sudo to install to /usr/local/bin"
  sudo mkdir -p /usr/local/bin
  sudo mv "$TMPFILE" "/usr/local/bin/${BINARY}"
  sudo chmod +x "/usr/local/bin/${BINARY}"
  echo ""
  echo "Installed: /usr/local/bin/${BINARY}"
  /usr/local/bin/${BINARY} --version 2>/dev/null || /usr/local/bin/${BINARY} --help | head -1
  setup_shell "/usr/local/bin/${BINARY}"
  echo ""
  echo "Done! Run 'mws --help' to get started."
  exit 0
fi

mv "$TMPFILE" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"

echo ""
echo "Installed: ${INSTALL_DIR}/${BINARY}"

# ── PATH check ───────────────────────────────────────────────────────────────

case ":$PATH:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "WARNING: ${INSTALL_DIR} is not in your PATH."
    echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo ""
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    echo ""
    ;;
esac

${INSTALL_DIR}/${BINARY} --version 2>/dev/null || ${INSTALL_DIR}/${BINARY} --help | head -1

# ── shell integration ────────────────────────────────────────────────────────

setup_shell "${INSTALL_DIR}/${BINARY}"

echo ""
echo "Done! Run 'mws --help' to get started."
