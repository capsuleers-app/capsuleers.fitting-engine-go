# Releases

## The rule: aligned versions

Module tag `vX.Y.Z` = npm `eve-fit-engine@X.Y.Z`. Same SDE data (`data/` is byte-identical to the
`data/` of that npm tarball), same engine behaviour (every parity fix is in both repos, see
[CONTRIBUTING.md](CONTRIBUTING.md)). There is no Go-only version: a fix that only this repo needs waits
for, or asks for, the next `eve-fit-engine` release. Versions stay `v0.x`: a `v2+` would need the `/v2`
suffix in the module path.

`VERSION` holds the current `X.Y.Z`; `go run ./tools/syncdata -check` (CI job `data`) fails if `data/`
drifts from the tarball of that version.

## Why data/ is synced, not rebuilt

The bundle is CCP data, not engine code. `eve-fit-engine` builds it from CCP's SDE API and publishes it in
its npm tarball; `tools/syncdata` downloads that tarball, verifies its sha512 `integrity` against the npm
registry, and copies `package/data/` verbatim (`data/embed.go` is this module's own file and is kept).
Rebuilding here would mean a second bundle generator to keep at parity, and could only ever target CCP's
*latest* build, never the one a past version shipped.

## The release gate

`bash scripts/gate.sh`: `go vet`, `tools/lint`, `tools/stacking -check -from-json`, `go test ./...` (and
again with `GOAMD64=v3`), Pyfa fixture parity exactly `662 passed, 0 failed`, effect coverage "truly
silent 0", and `tools/syncdata -check`. Both release paths run it; red = no tag.

## Path 1: SDE releases (automatic, `sde-sync.yml`)

Daily at 06:30 UTC, after `eve-fit-engine`'s own `sde-refresh` (05:00):

1. `go run ./tools/syncdata -latest` vs `VERSION`. Same → nothing to do.
2. Newer, and `data/` changes → sync, gate. **Green**: commit `data/` + `VERSION`, tag `vX.Y.Z`, push,
   warm the module proxy (`go list -m github.com/capsuleers-app/capsuleers.fitting-engine-go@vX.Y.Z`).
   **Red**: one open issue labelled `sde-sync`, red run, no tag. Usually it means `eve-fit-engine` needed
   an engine change to accept the new SDE (a new effect, a new handler): port it here, then run path 2.
3. Newer, but `data/` unchanged → it was a code release of `eve-fit-engine`. Its fixes must already be
   mirrored here, but a robot can't prove it: the job opens an `sde-sync` issue and a human uses path 2.

## Path 2: manual releases (`release.yml`)

Actions → **release** → Run workflow, `version` = the `eve-fit-engine` version (empty = latest on npm).
It syncs `data/`, writes `VERSION`, runs the gate, commits, tags `vX.Y.Z`, pushes and warms the proxy. It
refuses an existing tag. Use it for `eve-fit-engine` code releases once the fixes are mirrored, and to
release after fixing a red `sde-sync`.

Locally, the same steps by hand (tag and push are for a maintainer):

```bash
go run ./tools/syncdata -version X.Y.Z && echo X.Y.Z > VERSION
bash scripts/gate.sh
git add data VERSION && git commit -m "chore(release): vX.Y.Z (= eve-fit-engine X.Y.Z)"
git tag -a vX.Y.Z -m "vX.Y.Z (= eve-fit-engine X.Y.Z)" && git push --follow-tags
GOPROXY=https://proxy.golang.org go list -m github.com/capsuleers-app/capsuleers.fitting-engine-go@vX.Y.Z
```

## Repository setup (one-time, by an admin)

- Variables/secrets: `CAPSULEERS_BOT_APP_ID` (variable) and `CAPSULEERS_BOT_PRIVATE_KEY` (secret) for the
  GitHub App that pushes release commits and tags (install the App on this repo; if the default branch
  has a "PR only" ruleset, give the App a bypass, as in `eve-fit-engine`); optional
  `EVEWORKBENCH_API_KEY` (secret) for the real-fit corpus of `diff-parity`.
- Labels used by the workflows: `sde-sync`, `diff-parity`, `pyfa-drift`.
- No npm token: nothing here publishes to npm.
