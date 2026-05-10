#!/bin/sh
set -e

REPO="smsufyian/sprig"
BINARY="sprig"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

case "$OS" in
  linux|darwin) ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Fetch latest version tag
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version"
  exit 1
fi

ARCHIVE="sprig_${VERSION#v}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "Installing sprig ${VERSION} (${OS}/${ARCH})..."

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" | tar -xz -C "$TMP"

# Verify checksum
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
curl -fsSL "$CHECKSUM_URL" -o "$TMP/checksums.txt"
(cd "$TMP" && grep "${ARCHIVE}" checksums.txt | sha256sum -c -)

install -m 755 "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"

echo "sprig ${VERSION} installed to ${INSTALL_DIR}/${BINARY}"
echo "Run 'sprig --help' to get started."
