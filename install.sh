#!/usr/bin/env bash
set -euo pipefail

REPO="ctru0009/ccswap"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
VERSION="${VERSION:-latest}"

# Detect OS
RAW_OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$RAW_OS" in
  linux)  OS="linux"; ARCH_OS="linux" ;;
  darwin) OS="darwin"; ARCH_OS="macOS" ;;
  mingw*|msys*|cygwin*) OS="windows"; ARCH_OS="windows" ;;
  *)
    echo "Unsupported OS: $RAW_OS" >&2
    exit 1
    ;;
esac

# Detect arch
RAW_ARCH="$(uname -m)"
case "$RAW_ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $RAW_ARCH" >&2
    exit 1
    ;;
esac

# File extension and archive format per platform
EXT=""
ARCHIVE_EXT=".tar.gz"
if [ "$OS" = "windows" ]; then
  EXT=".exe"
  ARCHIVE_EXT=".zip"
fi

# Resolve download URL
if [ "$VERSION" = "latest" ]; then
  API_URL="https://api.github.com/repos/$REPO/releases/latest"
  echo "Fetching latest release info..." >&2
  API_JSON="$(curl -fSL "$API_URL" 2>/dev/null)" || {
    echo "Failed to fetch release info from GitHub API" >&2
    echo "Check your network or GitHub API rate limit (60 req/hr unauthenticated)" >&2
    exit 1
  }
  DOWNLOAD_URL="$(echo "$API_JSON" \
    | tr -d '\n' \
    | grep -o '"browser_download_url":"[^"]*'"$ARCH_OS"'[^"]*'"$ARCH"'[^"]*"' \
    | head -1 \
    | cut -d'"' -f4)" || true
  if [ -z "$DOWNLOAD_URL" ]; then
    echo "No release found for $RAW_OS/$RAW_ARCH" >&2
    echo "Check available releases: https://github.com/$REPO/releases" >&2
    exit 1
  fi
  VERSION_TAG="$(echo "$DOWNLOAD_URL" | sed -n 's|.*/download/\(v[^/]*\)/.*|\1|p')"
else
  VERSION_TAG="$VERSION"
  DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/ccswap_${VERSION}_${ARCH_OS}_${ARCH}${ARCHIVE_EXT}"
fi

echo "Downloading ccswap $VERSION_TAG for $RAW_OS/$RAW_ARCH..." >&2

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

ARCHIVE="$TMP_DIR/ccswap${ARCHIVE_EXT}"
curl -fSL -o "$ARCHIVE" "$DOWNLOAD_URL" || {
  echo "Download failed: $DOWNLOAD_URL" >&2
  echo "The release or file may not exist." >&2
  exit 1
}

# Verify file is non-empty
if [ ! -s "$ARCHIVE" ]; then
  echo "Downloaded archive is empty" >&2
  exit 1
fi

# Extract
if [ "$ARCHIVE_EXT" = ".zip" ]; then
  unzip -o "$ARCHIVE" -d "$TMP_DIR" >/dev/null
else
  tar -xzf "$ARCHIVE" -C "$TMP_DIR"
fi

# Find the binary
FOUND="$(find "$TMP_DIR" -name "ccswap$EXT" -type f 2>/dev/null | head -1)"
if [ -z "$FOUND" ]; then
  echo "Binary not found in archive" >&2
  exit 1
fi

# Install
mkdir -p "$BIN_DIR"
install -m 755 "$FOUND" "$BIN_DIR/ccswap$EXT"

echo "✓ Installed ccswap$EXT to $BIN_DIR/ccswap$EXT" >&2
echo "  Make sure $BIN_DIR is in your PATH." >&2
