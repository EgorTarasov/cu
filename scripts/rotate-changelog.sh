#!/bin/sh
# Rotate CHANGELOG.md after a release:
#   - rename "## [Unreleased]" -> "## [<version>] - <YYYY-MM-DD>"
#   - insert a fresh empty "## [Unreleased]" section above it
#
# Usage:
#   scripts/rotate-changelog.sh <version>
#
# Idempotent: if "## [<version>]" already exists, exits without modifying.

set -eu

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <version>" >&2
    exit 2
fi

version="${1#v}"
file="${CHANGELOG_FILE:-CHANGELOG.md}"
date="$(date -u +%Y-%m-%d)"

if [ ! -f "$file" ]; then
    echo "$file not found" >&2
    exit 1
fi

if grep -q "^## \[$version\]" "$file"; then
    echo "Version $version already present in $file — nothing to rotate."
    exit 0
fi

if ! grep -q '^## \[Unreleased\]' "$file"; then
    echo "No [Unreleased] section in $file — nothing to rotate." >&2
    exit 0
fi

tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

awk -v version="$version" -v date="$date" '
    /^## \[Unreleased\]/ && !done {
        print "## [Unreleased]"
        print ""
        print "## [" version "] - " date
        done = 1
        next
    }
    { print }
' "$file" > "$tmp"

mv "$tmp" "$file"
echo "Rotated $file: [Unreleased] -> [$version] - $date"
