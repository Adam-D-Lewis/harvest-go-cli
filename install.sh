#!/bin/sh
set -e

REPO="aktech/harvest-go-cli"
BINARY="harvest-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  OS="Linux" ;;
  Darwin) OS="Darwin" ;;
  *)      echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="x86_64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  i386|i686)     ARCH="i386" ;;
  *)             echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

# Get latest release tag
TAG="$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4)"
if [ -z "$TAG" ]; then
  echo "Failed to fetch latest release" >&2
  exit 1
fi

ARCHIVE="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ARCHIVE}"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading ${BINARY} ${TAG} for ${OS}/${ARCH}..."
curl -sL "$URL" -o "${TMPDIR}/${ARCHIVE}"

echo "Extracting..."
tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

echo "Installing to ${INSTALL_DIR}/harvest..."
install -m 755 "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/harvest"

echo "harvest ${TAG} installed successfully!"
