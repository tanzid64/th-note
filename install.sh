#!/bin/sh
# th-note installer for Linux and macOS.
#
#   curl -fsSL https://raw.githubusercontent.com/tanzid64/th-note/main/install.sh | sh
#
# Downloads the latest release binary for your OS/arch and installs it onto
# your PATH. Windows users: download the .zip from the Releases page instead.
set -eu

REPO="tanzid64/th-note"
BIN="th-note"

# --- Detect OS ---------------------------------------------------------------
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux | darwin) ;;
  *)
    echo "Unsupported OS: $OS" >&2
    echo "On Windows, download the .zip from https://github.com/$REPO/releases/latest" >&2
    exit 1
    ;;
esac

# --- Detect architecture -----------------------------------------------------
ARCH=$(uname -m)
case "$ARCH" in
  x86_64 | amd64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

ASSET="${BIN}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

# --- Download & extract ------------------------------------------------------
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Downloading ${ASSET}…"
if ! curl -fsSL "$URL" -o "$TMP/$ASSET"; then
  echo "Download failed. No release asset at:" >&2
  echo "  $URL" >&2
  echo "Check https://github.com/$REPO/releases for available builds." >&2
  exit 1
fi

tar -xzf "$TMP/$ASSET" -C "$TMP"

# --- Choose an install directory --------------------------------------------
if [ -w /usr/local/bin ]; then
  DEST="/usr/local/bin"
  SUDO=""
elif command -v sudo >/dev/null 2>&1; then
  DEST="/usr/local/bin"
  SUDO="sudo"
else
  DEST="$HOME/.local/bin"
  SUDO=""
  mkdir -p "$DEST"
fi

echo "Installing to ${DEST}/${BIN}…"
$SUDO install -m 0755 "$TMP/$BIN" "$DEST/$BIN"

# --- Done --------------------------------------------------------------------
echo ""
echo "✓ Installed $("$DEST/$BIN" --version 2>/dev/null || echo "$BIN")"
case ":$PATH:" in
  *":$DEST:"*) echo "Run it with: $BIN" ;;
  *)
    echo "Note: $DEST is not on your PATH."
    echo "Add this to your shell profile:"
    echo "  export PATH=\"\$PATH:$DEST\""
    echo "Then run: $BIN"
    ;;
esac
