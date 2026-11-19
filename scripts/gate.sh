#!/usr/bin/env bash
# The release gate, in one place (sde-sync.yml and release.yml run it before any tag; run it locally
# before asking for a release). Every step must pass:
#   vet + engine analyzer, stacking registry rendered from its JSON, the whole test suite (Pyfa fixture
#   parity 662/0, effect coverage "truly silent 0", drone bandwidth, unit tests), the same suite with FMA
#   allowed (GOAMD64=v3), and data/ == data/ of eve-fit-engine@$(cat VERSION).
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."
go vet ./...
go run ./tools/lint
go run ./tools/stacking -check -from-json
go test ./...
GOAMD64=v3 go test ./...
parity="$(go test ./test/parity -run TestPyfaParity -count=1 -v)"
printf '%s\n' "$parity" | grep 'Summary:'
grep -q 'Summary: 662 passed, 0 failed.' <<<"$parity" || { echo "[gate] fixture parity is not 662/0"; exit 1; }
go run ./tools/syncdata -check
echo "[gate] green ($(tr -d '[:space:]' < VERSION))"
