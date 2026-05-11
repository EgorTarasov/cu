#!/bin/sh
# Install the `cu` CLI from the latest GitHub Release.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/EgorTarasov/cu/main/install.sh | sh
#
# Honored environment variables:
#   CU_VERSION       — install a specific tag (e.g. v0.1.5). Default: latest.
#   CU_INSTALL_DIR   — install path. Default: ~/.local/bin.
#                      Pass /usr/local/bin to install system-wide (will sudo).
#
# Supported platforms: macOS (Intel/Apple Silicon), Linux x86_64/arm64
# (tested on Ubuntu, Fedora — any glibc Linux works).

set -eu

REPO="EgorTarasov/cu"
BIN_NAME="cu"
DEFAULT_DIR="$HOME/.local/bin"
INSTALL_DIR="${CU_INSTALL_DIR:-$DEFAULT_DIR}"

err() { printf '\033[31merror:\033[0m %s\n' "$*" >&2; exit 1; }
info() { printf '\033[36m::\033[0m %s\n' "$*"; }

require() {
    command -v "$1" >/dev/null 2>&1 || err "missing required tool: $1"
}

require uname
require curl
require tar
require grep
require awk
require mktemp

# --- Detect OS / ARCH (must match GoReleaser name_template) ---
raw_os="$(uname -s)"
case "$raw_os" in
    Darwin) os="Darwin" ;;
    Linux)  os="Linux"  ;;
    *) err "unsupported OS: $raw_os (only macOS and Linux are supported)" ;;
esac

raw_arch="$(uname -m)"
case "$raw_arch" in
    x86_64|amd64)  arch="x86_64" ;;
    arm64|aarch64) arch="arm64"  ;;
    *) err "unsupported architecture: $raw_arch" ;;
esac

# --- Resolve version ---
VERSION="${CU_VERSION:-}"
if [ -z "$VERSION" ]; then
    info "Resolving latest release..."
    VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name"' \
        | head -n1 \
        | awk -F '"' '{print $4}')"
fi
[ -n "$VERSION" ] || err "could not determine version (set CU_VERSION manually)"
case "$VERSION" in v*) ;; *) VERSION="v${VERSION}" ;; esac

archive="${BIN_NAME}_${os}_${arch}.tar.gz"
base_url="https://github.com/${REPO}/releases/download/${VERSION}"

# --- Download + verify + extract in a tmpdir ---
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT INT TERM

info "Downloading ${archive} (${VERSION})..."
curl -fsSL --retry 3 -o "${tmpdir}/${archive}"      "${base_url}/${archive}" \
    || err "failed to download ${archive}. Check ${base_url}"
curl -fsSL --retry 3 -o "${tmpdir}/checksums.txt"   "${base_url}/checksums.txt" \
    || err "failed to download checksums.txt"

info "Verifying checksum..."
(
    cd "$tmpdir"
    expected="$(awk -v f="$archive" '$2 == f {print $1}' checksums.txt)"
    [ -n "$expected" ] || err "checksum entry for $archive not found"

    if command -v sha256sum >/dev/null 2>&1; then
        actual="$(sha256sum "$archive" | awk '{print $1}')"
    elif command -v shasum >/dev/null 2>&1; then
        actual="$(shasum -a 256 "$archive" | awk '{print $1}')"
    else
        err "neither sha256sum nor shasum is available"
    fi

    [ "$expected" = "$actual" ] || err "checksum mismatch (expected $expected, got $actual)"
)

info "Extracting..."
tar -xzf "${tmpdir}/${archive}" -C "$tmpdir"

[ -f "${tmpdir}/${BIN_NAME}" ] || err "binary ${BIN_NAME} not found inside the archive"

# --- Install ---
need_sudo=""
if ! mkdir -p "$INSTALL_DIR" 2>/dev/null; then
    need_sudo="sudo"
fi
if [ -d "$INSTALL_DIR" ] && [ ! -w "$INSTALL_DIR" ]; then
    need_sudo="sudo"
fi

if [ -n "$need_sudo" ]; then
    info "Installing to $INSTALL_DIR (requires sudo)..."
    require sudo
    sudo mkdir -p "$INSTALL_DIR"
    sudo install -m 0755 "${tmpdir}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
else
    info "Installing to $INSTALL_DIR..."
    install -m 0755 "${tmpdir}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
fi

# macOS: remove Gatekeeper quarantine attribute so binary runs without warning.
if [ "$os" = "Darwin" ] && command -v xattr >/dev/null 2>&1; then
    if [ -n "$need_sudo" ]; then
        sudo xattr -dr com.apple.quarantine "${INSTALL_DIR}/${BIN_NAME}" 2>/dev/null || true
    else
        xattr -dr com.apple.quarantine "${INSTALL_DIR}/${BIN_NAME}" 2>/dev/null || true
    fi
fi

# --- Done ---
installed_version="$("${INSTALL_DIR}/${BIN_NAME}" --version 2>/dev/null | head -n1 || true)"
info "Installed: ${INSTALL_DIR}/${BIN_NAME}"
[ -n "$installed_version" ] && info "Version:   ${installed_version}"

# PATH hint
case ":${PATH}:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
        printf '\n'
        printf '\033[33mNote:\033[0m %s is not in your PATH.\n' "$INSTALL_DIR"
        printf 'Add this to ~/.bashrc, ~/.zshrc, etc.:\n\n'
        printf '    export PATH="$PATH:%s"\n\n' "$INSTALL_DIR"
        ;;
esac

printf 'Next steps:\n'
printf '  cu login           # authenticate via browser\n'
printf '  cu --help          # see all commands\n'
