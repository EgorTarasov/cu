#!/bin/sh
# Extract release notes for a given version from CHANGELOG.md.
#
# Usage:
#   scripts/extract-changelog.sh <version>
#
# Looks for "## [<version>]" first; falls back to "## [Unreleased]".
# Prints the section body (everything until the next "## [" heading).
# Empty output is a signal that the changelog wasn't updated — the caller
# should decide whether to substitute a default message.

set -eu

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <version>" >&2
    exit 2
fi

version="${1#v}"
file="${CHANGELOG_FILE:-CHANGELOG.md}"

if [ ! -f "$file" ]; then
    echo "$file not found" >&2
    exit 1
fi

extract() {
    awk -v heading="$1" '
        $0 ~ "^## \\[" heading "\\]" { flag = 1; next }
        flag && /^## \[/ { exit }
        flag { print }
    ' "$file"
}

# Try explicit version first, then [Unreleased].
out="$(extract "$version" | sed -e '/./,$!d' | awk 'NF {p=1} p')"
if [ -z "$(printf '%s' "$out" | tr -d '[:space:]')" ]; then
    out="$(extract "Unreleased" | sed -e '/./,$!d' | awk 'NF {p=1} p')"
fi

printf '%s\n' "$out"
