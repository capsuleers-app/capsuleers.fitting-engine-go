#!/usr/bin/env bash
# Next module version, aligned with eve-fit-engine on npm (tag vX.Y.Z here = X.Y.Z there).
#   mine = the highest v* tag of this repo (none = 0.0.0)
#   twin = the latest eve-fit-engine on npm (go run ./tools/syncdata -latest)
#   twin > mine -> reach the twin's version (data/ then comes from that tarball)
#   otherwise   -> mine + patch (Go ahead of the twin; data/ stays the latest npm one)
# Prints X.Y.Z on stdout. Test hooks: MINE=x.y.z (empty = no tag) and TWIN=x.y.z.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
SEMVER='^[0-9]+\.[0-9]+\.[0-9]+$'

gt() { [ "$1" != "$2" ] && [ "$(printf '%s\n%s\n' "$1" "$2" | sort -V | tail -1)" = "$1" ]; }

if [ -z "${MINE+x}" ]; then
  # Remote tags, not local ones: a failure must stop the release instead of reading as "no tags".
  refs="$(git ls-remote --tags --refs origin 'v*')"
  MINE="$(printf '%s\n' "$refs" | sed -n 's#.*refs/tags/v##p' | { grep -E "$SEMVER" || true; } | sort -V | tail -1)"
fi
mine="${MINE:-0.0.0}"
twin="${TWIN:-$(go run ./tools/syncdata -latest)}"
grep -qE "$SEMVER" <<<"$twin" || { echo "bad npm version '$twin'" >&2; exit 1; }

if gt "$twin" "$mine"; then
  next="$twin"
else
  IFS=. read -r a b c <<<"$mine"
  next="$a.$b.$((c + 1))"
fi
echo "mine=$mine twin=$twin -> $next" >&2
echo "$next"
