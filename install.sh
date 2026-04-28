#!/usr/bin/env bash
set -euo pipefail

REPO="ctru0009/ccswap"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-latest}"

# Detect OS and arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux)  OS="linux" ;;
  darwin) OS="darwin" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
  *)
    echo "Unsupported OS: $OS" >&2
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

EXT=""
if [ "$OS" = "windows" ]; then
  EXT=".exe"
fi

# Determine release URL
if [ "$VERSION" = "latest" ]; then
  API_URL="https://api.github.com/repos/$REPO/releases/latest"
  echo "Fetching latest release info..." >&2
  DOWNLOAD_URL=$(curl -sSL "$API_URL" | grep "browser_download_url" | grep "$OS" | grep "$ARCH" | head -1 | cut -d'"' -f4)
  if [ -z "$DOWNLOAD_URL" ]; then
    echo "No release found for $OS/$ARCH" >&2
    exit 1
  fi
  # Extract version from URL
  VERSION_TAG=$(echo "$DOWNLOAD_URL" | sed -n 's|.*/download/\(v[^/]*\)/.*|\1|p')
else
  VERSION_TAG="$VERSION"
  DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/ccswap_${VERSION}_${OS}_${ARCH}.tar.gz"
fi

echo "Downloading ccswap $VERSION_TAG for $OS/$ARCH..." >&2

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

ARCHIVE="$TMP_DIR/ccswap.tar.gz"
curl -sSL -o "$ARCHIVE" "$DOWNLOAD_URL"

# Extract
if [ "$OS" = "windows" ]; then
  unzip -o "$ARCHIVE" -d "$TMP_DIR" >/dev/null 2>&1 || tar -xzf "$ARCHIVE" -C "$TMP_DIR"
else
  tar -xzf "$ARCHIVE" -C "$TMP_DIR"
fi

# Find the binary
FOUND=$(find "$TMP_DIR" -name "ccswap$EXT" -type f 2>/dev/null | head -1)
if [ -z "$FOUND" ]; then
  echo "Binary not found in archive" >&2
  exit 1
fi

# Install
mkdir -p "$BIN_DIR"
install -m 755 "$FOUND" "$BIN_DIR/ccswap$EXT"

echo "✓ Installed ccswap$EXT to $BIN_DIR/ccswap$EXT" >&2
echo "  Make sure $BIN_DIR is in your PATH." >&2
