#!/bin/sh
set -e

# CU CLI install script
# Usage: curl -fsSL https://raw.githubusercontent.com/EgorTarasov/cu/main/install.sh | sh

REPO="EgorTarasov/cu"
BINARY_NAME="cu"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS
OS="$(uname -s)"
case "$OS" in
    Linux*)  OS="linux" ;;
    Darwin*) OS="darwin" ;;
    *)
        echo "Error: unsupported OS: $OS"
        echo "For Windows, download the .exe from https://github.com/$REPO/releases/latest"
        exit 1
        ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)
        echo "Error: unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "Detected platform: ${OS}/${ARCH}"

# Get latest version
echo "Fetching latest release..."
if command -v curl >/dev/null 2>&1; then
    VERSION="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"
elif command -v wget >/dev/null 2>&1; then
    VERSION="$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')"
else
    echo "Error: curl or wget is required"
    exit 1
fi

if [ -z "$VERSION" ]; then
    echo "Error: could not determine latest version"
    exit 1
fi

echo "Latest version: $VERSION"

SUFFIX="${OS}-${ARCH}"
BINARY="cu-${VERSION}-${SUFFIX}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/${VERSION}/${BINARY}"
CHECKSUM_URL="${DOWNLOAD_URL}.sha256"

# Create temp directory
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading ${BINARY}..."
if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$TMP_DIR/$BINARY" "$DOWNLOAD_URL"
    curl -fsSL -o "$TMP_DIR/$BINARY.sha256" "$CHECKSUM_URL"
else
    wget -qO "$TMP_DIR/$BINARY" "$DOWNLOAD_URL"
    wget -qO "$TMP_DIR/$BINARY.sha256" "$CHECKSUM_URL"
fi

# Verify checksum
echo "Verifying checksum..."
cd "$TMP_DIR"
if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c "$BINARY.sha256"
elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 -c "$BINARY.sha256"
else
    echo "Warning: could not verify checksum (sha256sum/shasum not found)"
fi

# Install
echo "Installing to ${INSTALL_DIR}/${BINARY_NAME}..."
chmod +x "$BINARY"
if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY" "$INSTALL_DIR/$BINARY_NAME"
else
    sudo mv "$BINARY" "$INSTALL_DIR/$BINARY_NAME"
fi

echo ""
echo "Successfully installed cu ${VERSION}"
"$INSTALL_DIR/$BINARY_NAME" --version 2>/dev/null || true
