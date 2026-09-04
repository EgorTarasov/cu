#!/usr/bin/env bash
# End-to-end test for install.sh.
#
# Builds a release archive from the current source, serves it over localhost,
# and drives the real installer against it. No network, no published release.
#
# Usage: test/e2e/install_test.sh
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
BIN_NAME="cuni"
VERSION="v0.0.0-e2e"

pass=0
fail=0
ok()   { printf '  \033[32m✓\033[0m %s\n' "$1"; pass=$((pass + 1)); }
bad()  { printf '  \033[31m✗\033[0m %s\n' "$1"; fail=$((fail + 1)); }
head_() { printf '\n\033[36m%s\033[0m\n' "$1"; }

# path_lookup <PATH> <name> — resolve <name> under a specific PATH.
# `command` is a shell builtin; macOS also ships /usr/bin/command but Linux
# does not, so `env PATH=... command -v` dies with 127 there. Going through
# sh -c keeps the builtin available on both.
path_lookup() {
    env PATH="$1" sh -c 'command -v "$1"' sh "$2" 2>/dev/null || true
}

case "$(uname -s)" in
    Darwin) os="Darwin" ;;
    Linux)  os="Linux"  ;;
    *) echo "unsupported OS for this test: $(uname -s)" >&2; exit 1 ;;
esac
case "$(uname -m)" in
    x86_64|amd64)  arch="x86_64" ;;
    arm64|aarch64) arch="arm64"  ;;
    *) echo "unsupported arch: $(uname -m)" >&2; exit 1 ;;
esac

workdir="$(mktemp -d)"
serve_dir="$workdir/release"
fake_home="$workdir/home"
mkdir -p "$serve_dir" "$fake_home"

server_pid=""
cleanup() {
    [ -n "$server_pid" ] && kill "$server_pid" 2>/dev/null || true
    rm -rf "$workdir"
}
trap cleanup EXIT INT TERM

head_ "Building release archive"
go build -o "$workdir/$BIN_NAME" "$REPO_ROOT/cmd/$BIN_NAME"
archive="${BIN_NAME}_${os}_${arch}.tar.gz"
tar -czf "$serve_dir/$archive" -C "$workdir" "$BIN_NAME"
(
    cd "$serve_dir"
    if command -v shasum >/dev/null 2>&1; then
        shasum -a 256 "$archive" > checksums.txt
    else
        sha256sum "$archive" > checksums.txt
    fi
)
echo "  $archive ($(wc -c < "$serve_dir/$archive" | tr -d ' ') bytes)"

head_ "Serving release on localhost"
port="$(python3 -c 'import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()')"
(cd "$serve_dir" && exec python3 -m http.server "$port" --bind 127.0.0.1) > "$workdir/server.log" 2>&1 &
server_pid=$!
base_url="http://127.0.0.1:${port}"

ready=""
for _ in $(seq 1 60); do
    if curl -fsS "${base_url}/checksums.txt" -o /dev/null 2>/dev/null; then ready=1; break; fi
    perl -e 'select(undef,undef,undef,0.1)'
done
[ -n "$ready" ] || { echo "server did not start"; cat "$workdir/server.log"; exit 1; }
echo "  $base_url"

# run_install <prefix> [SHELL] [extra env assignments...]
run_install() {
    prefix="$1"; shift
    login_shell="${1:-/bin/zsh}"; [ $# -gt 0 ] && shift
    env HOME="$fake_home" \
        SHELL="$login_shell" \
        CU_VERSION="$VERSION" \
        CU_INSTALL_DIR="$prefix" \
        CU_DOWNLOAD_BASE_URL="$base_url" \
        PATH="/usr/bin:/bin:/usr/sbin:/sbin" \
        "$@" \
        sh "$REPO_ROOT/install.sh"
}

# ---------------------------------------------------------------- happy path
head_ "Install into a clean prefix"
install_dir="$fake_home/.local/bin"
if run_install "$install_dir" > "$workdir/install.log" 2>&1; then
    ok "installer exited 0"
else
    bad "installer failed"; sed 's/^/      /' "$workdir/install.log"
fi

if [ -x "$install_dir/$BIN_NAME" ]; then ok "binary installed and executable"
else bad "binary missing at $install_dir/$BIN_NAME"; fi

if "$install_dir/$BIN_NAME" --version >/dev/null 2>&1; then ok "--version runs"
else bad "--version failed"; fi

# The whole point of the rename: nothing in the base system may own this name.
head_ "Name is free on a stock system PATH"
shadow="$(path_lookup "/usr/bin:/bin:/usr/sbin:/sbin" "$BIN_NAME")"
if [ -z "$shadow" ]; then ok "no system binary named '$BIN_NAME'"
else bad "system binary shadows the install: $shadow"; fi

# Regression guard: `cu` collided with /usr/bin/cu (UUCP) on every macOS.
if [ "$os" = "Darwin" ]; then
    legacy="$(path_lookup "/usr/bin:/bin:/usr/sbin:/sbin" cu)"
    if [ -n "$legacy" ]; then ok "confirmed why 'cu' was unusable: $legacy"
    else printf '  \033[33m!\033[0m /usr/bin/cu absent on this host (unexpected on macOS)\n'; fi
fi

head_ "PATH is configured automatically (one-line install)"
zshrc="$fake_home/.zshrc"
if [ -f "$zshrc" ]; then ok "zsh rc created: ~/.zshrc"
else bad "installer did not write ~/.zshrc"; fi

line="$(grep -o 'export PATH="[^"]*"' "$zshrc" 2>/dev/null | head -n1 || true)"
if [ -z "$line" ]; then
    bad "no PATH line written"
elif printf '%s' "$line" | grep -q 'PATH="\$PATH:'; then
    bad "PATH line appends; a system binary would win: $line"
else
    ok "PATH line prepends: $line"
fi

if grep -Fq "$install_dir" "$zshrc" 2>/dev/null; then ok "points at the install dir"
else bad "PATH line does not mention $install_dir"; fi

# Sourcing the rc must actually put the binary on PATH.
sourced="$(env HOME="$fake_home" PATH="/usr/bin:/bin" /bin/sh -c \
    ". $zshrc >/dev/null 2>&1; command -v $BIN_NAME" 2>/dev/null || true)"
if [ "$sourced" = "$install_dir/$BIN_NAME" ]; then ok "sourcing the rc exposes $BIN_NAME"
else bad "after sourcing, resolved to '${sourced:-nothing}'"; fi

head_ "Re-running the installer is idempotent"
run_install "$install_dir" /bin/zsh > "$workdir/install2.log" 2>&1 || true
count="$(grep -c "added by ${BIN_NAME} installer" "$zshrc" 2>/dev/null || echo 0)"
if [ "$count" -eq 1 ]; then ok "PATH block written exactly once (found $count)"
else bad "PATH block duplicated: found $count"; fi

head_ "Shell detection picks the right rc file"
bash_home_marker="$fake_home/.bash_profile"
rm -f "$bash_home_marker"
run_install "$install_dir" /bin/bash > "$workdir/install3.log" 2>&1 || true
if [ "$os" = "Darwin" ]; then
    if [ -f "$bash_home_marker" ]; then ok "bash on macOS -> ~/.bash_profile (login shell)"
    else bad "bash on macOS did not write ~/.bash_profile"; fi
else
    if [ -f "$fake_home/.bashrc" ]; then ok "bash on Linux -> ~/.bashrc"
    else bad "bash on Linux did not write ~/.bashrc"; fi
fi

head_ "CU_NO_MODIFY_PATH is respected"
optout_home="$workdir/optout"; mkdir -p "$optout_home"
env HOME="$optout_home" SHELL=/bin/zsh CU_VERSION="$VERSION" \
    CU_INSTALL_DIR="$optout_home/.local/bin" CU_DOWNLOAD_BASE_URL="$base_url" \
    CU_NO_MODIFY_PATH=1 PATH="/usr/bin:/bin:/usr/sbin:/sbin" \
    sh "$REPO_ROOT/install.sh" > "$workdir/optout.log" 2>&1 || true
if [ ! -e "$optout_home/.zshrc" ]; then ok "no rc file written when opted out"
else bad "rc file written despite CU_NO_MODIFY_PATH=1"; fi
if grep -q 'export PATH=' "$workdir/optout.log"; then ok "printed the PATH line instead"
else bad "opt-out did not print the manual PATH line"; fi

resolved="$(path_lookup "${install_dir}:/usr/bin:/bin" "$BIN_NAME")"
if [ "$resolved" = "$install_dir/$BIN_NAME" ]; then ok "resolves to the installed binary"
else bad "resolved to $resolved"; fi

# ------------------------------------------------------------ integrity path
head_ "Checksum mismatch is rejected"
printf 'corrupted' >> "$serve_dir/$archive"
if run_install "$workdir/bad-prefix" > "$workdir/bad.log" 2>&1; then
    bad "installer accepted a corrupted archive"
else
    if grep -q "checksum mismatch" "$workdir/bad.log"; then ok "rejected with checksum mismatch"
    else bad "failed, but not on checksum: $(tail -n1 "$workdir/bad.log")"; fi
fi
if [ ! -e "$workdir/bad-prefix/$BIN_NAME" ]; then ok "nothing installed from the bad archive"
else bad "corrupted archive still got installed"; fi

# ---------------------------------------------------------------------- done
printf '\n\033[36mResult\033[0m: %d passed, %d failed\n' "$pass" "$fail"
[ "$fail" -eq 0 ]
