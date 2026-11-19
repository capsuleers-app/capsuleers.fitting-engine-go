# Contributing to capsuleers.fitting-engine-go

This module is the Go twin of [`eve-fit-engine`](https://github.com/capsuleers-app/capsuleers.fitting-engine)
(TypeScript, npm). The two are **peers**: same engine, same versions, same SDE data, each validated
against Pyfa on its own. Read §1 before your first change: it is the rule that keeps them twins.

## 0. TL;DR: the loop

```bash
# one-time
bash scripts/setup-pyfa-oracle.sh          # pyfa at PYFA_REF into .pyfa/ (venv + eve.db), for tools/diff
go run ./tools/stacking -fetch-sde         # raw SDE tables into .sde-src/, for tools/stacking

# edit the engine ... then ALWAYS, before committing:
go vet ./... && go run ./tools/lint && go test ./...    # fixture parity 662/0 is inside go test
go run ./tools/diff -source=all                        # exit 0 = no UNEXPECTED diff vs pyfa
bash scripts/gate.sh                                   # the full release gate (CI runs it too)
```

`go test` output can be condensed by wrapper tools: check for `Summary: 662 passed, 0 failed.` with
`go test ./test/parity -run TestPyfaParity -count=1 -v`, not just for `ok`.

## 1. Fixes in both repos

Every change that alters engine behaviour (a parity fix, a new legacy handler, a hand-written table:
constants, modifier tables, stacking groups, profiles) is made in **both** repositories, normally as two
PRs opened together, cross-linked, and merged together:

1. Fix it where you found it, with its evidence (failing fixture, diff entry, pyfa source reference).
2. Port it to the other repo. The file map is [`docs/mapping.md`](docs/mapping.md) (TS file ↔ Go file,
   same names); port literally (same order of operations, see §6).
3. Both gates green: here `bash scripts/gate.sh` + `go run ./tools/diff -source=all`; there
   `npm run test:pyfa`, `npm run audit:coverage`, `npm run diff -- --source=all`.
4. Each PR body links the other and states both results (e.g. `parity 662/0, diff 0 unexpected`).
5. Merge both. Each merge releases by itself (`release.yml`): the version is the npm `eve-fit-engine` one if it
   is higher than our tag, else our tag + patch, so the two merges end on the same version in either order
   (RELEASE.md). Docs/CI/test/tools-only merges release nothing.

There is no automatic TS↔Go comparison any more: the only thing that keeps the twins equal is this
procedure plus both being equal to Pyfa. So:

- A fix that changes a result in one repo and not the other is a bug in the port, not a "Go-only fix".
- `test/parity/fixtures.json` is byte-identical in both repos: a fixture added or corrected here goes there too
  (and vice versa), with the same expected values.
- The generated stacking registry must be byte-identical in both repos at the same pin:
  `generated/stacking-penalised.json` here is what `npm run stacking:build` writes there.
- Harness changes (corpus generators, schema, tolerances) also go into both, or the two diff reports stop
  being comparable. The Go harness builds the same generated and implants corpora as the TS one (verified spec by spec at
  the extraction with `-dump-corpus`).

## 2. Prerequisites & setup

- **Go** (version in `go.mod`; `toolchain` downloads it if needed). No module dependencies.
- **Python 3** + `venv` and **git**, only for the pyfa oracle (`scripts/setup-pyfa-oracle.sh`).

```bash
git clone git@github.com:capsuleers-app/capsuleers.fitting-engine-go.git
cd capsuleers.fitting-engine-go
go test ./...
```

Gitignored, never commit: `.pyfa/` (pyfa checkout + venv + eve.db), `.sde-src/` (raw CCP tables),
`.workbench/` (other capsuleers' published fits: read like the site does, never redistributed).
`data/` IS committed, and changes only through `go run ./tools/syncdata`.

## 3. Repository layout

```
fitengine/           public API: ComputeFit, EncodeJSON, types; eft/ export/ checks/ profiles/ bundle/
internal/engine/     orchestrator, modifier dispatcher + legacy handlers, stacking, item state, tables
internal/derived/    offense, ehp, tank, capacitor (port of eos/capSim.py), structure, application
internal/effects/    weapon + ewar helpers
internal/gen/        GENERATED stacking registry (tools/stacking)
internal/omap/       insertion-ordered Map/Set (Go maps are unordered: never range over them in the engine)
internal/jsmath/     JS numeric/string semantics where Go differs (round, parse, toFixed, compare)
internal/types/      dataset, fit and result types
data/                embedded SDE bundle (synced from eve-fit-engine; embed.go is ours)
generated/           stacking-penalised.json, the canonical model of the stacking registry
oracle/              pyfa_oracle.py: headless pyfa eos, batch of fit-specs in, normalized stats out
scripts/             setup-pyfa-oracle.sh, gate.sh
test/parity/         the fixture suite (fixtures.json, pyfa/ screenshots, parity_test.go)
test/diff/           known-diffs.json + README of the harness
tools/               lint, stacking, diff, syncdata, drift, workbench (all `go run ./tools/<name>`)
```

## 4. Command reference

| command | does |
|---|---|
| `go test ./...` | unit tests + fixture parity (662) + effect coverage + drone bandwidth |
| `go run ./tools/lint` | engine analyzer: no-map-range, fma-cast (blocking in CI) |
| `bash scripts/gate.sh` | the release gate (everything above + GOAMD64=v3 + data drift) |
| `bash scripts/setup-pyfa-oracle.sh` | pyfa at `PYFA_REF` into `.pyfa/` (`--rebuild` for eve.db) |
| `go run ./tools/diff [-source=generated\|workbench\|implants\|all] [-strict]` | differential harness vs pyfa |
| `go run ./tools/diff -recalibrate` | after a pin move: rewrite known-diffs.json + PYFA_REF |
| `go run ./tools/stacking [-check] [-from-json]` | stacking registry from pinned pyfa (+ `-fetch-sde`) |
| `go run ./tools/syncdata [-check] [-version X.Y.Z] [-latest]` | data/ from the eve-fit-engine tarball |
| `go run ./tools/drift [-update]` | pyfa effects.py drift vs snapshot + LegacyEffectIDs |
| `go run ./tools/workbench` | real-fit corpus from EVE Workbench (`EVEWORKBENCH_API_KEY`, concurrency 2) |
| `go test ./test/parity -run xxx -bench ComputeFit` | informative benchmark |

`tools/diff` flags: `-ships=587,29990`, `-limit=20`, `-group=Loki`, `-only=bonused`,
`-stats=offense,defense`, `-tol=0.02`, `-json`, `-per-hull=3`, `-dump-corpus=file.json`.
`$PYFA_DIR` points the harness and the generator at another pyfa checkout.

## 5. The two validation suites

**Fixture suite (`test/parity`, the gate).** 21 hand-curated fits at All-V with a Uniform damage
profile, ~662 assertions against values read from real Pyfa screenshots (`test/parity/pyfa/`). Never
weaken it to make another check pass. To add a fixture: build the fit in Pyfa (All-V, Uniform),
screenshot the stat panel into `test/parity/pyfa/`, add the fit + `Expected` block to `fixtures.json`,
`go test ./test/parity`, then add the same fixture to `eve-fit-engine` (§1).

**Differential harness (`tools/diff`, the wide net).** Every stat against a headless pyfa oracle over
three corpora: `generated` (4 synthetic fits per published ship, 1 720), `implants` (every implant set and
ship-affecting booster, drawbacks off and on, ~1 870), `workbench` (2-3 real published fits per hull,
~900). Exit 0 iff no UNEXPECTED diff and no oracle staticdata drift. Accepted differences live in
`test/diff/known-diffs.json`, each with a proven root cause (empty today). A shrinking corpus always says
so in the report. Mechanism: [`test/diff/README.md`](test/diff/README.md).

## 6. Porting rules (both directions)

- Literal port: same order of operations and accumulation, no "idiomatic" refactor; `float64` wherever
  the TS does arithmetic.
- `Math.round` → `jsmath.Round`, `Math.min/max` → `jsmath.Min/Max`, `Number/parseFloat/parseInt` →
  `jsmath.*`, `toFixed` → `jsmath.ToFixed`, string `<` → `jsmath.CompareUTF16`, `localeCompare` →
  `jsmath.LocaleCompare`; `Math.exp/log/pow` → `math`.
- **Never `range` over a Go map** in the engine packages (use `omap`), and **always `float64(a*b) + c`**
  where a product enters a sum (no FMA): `go run ./tools/lint` enforces both.
- Never edit `internal/gen/*` or `generated/stacking-penalised.json` by hand (`tools/stacking`).

## 7. Moving the pyfa pin

`PYFA_REF`, `test/diff/known-diffs.json` and the stacking registry are calibrated against ONE pyfa commit,
always the one `eve-fit-engine` pins. `pyfa-pin-bump.yml` follows it automatically; by hand:

```bash
echo <commit-of-eve-fit-engine> > PYFA_REF
bash scripts/setup-pyfa-oracle.sh         # re-pins .pyfa and rebuilds eve.db
go run ./tools/stacking -fetch-sde && go run ./tools/stacking   # must equal eve-fit-engine's JSON
bash scripts/gate.sh
go run ./tools/diff -recalibrate          # PENDING REVIEW entries → classify (bug → fix both repos)
go run ./tools/diff -source=all           # must exit 0
```

## 8. Porting / changing a Pyfa effect

1. `go run ./tools/drift` (or the `pyfa-drift` issue) → which effect ids changed.
2. Data-driven (`modifierInfo` present → nothing to do) or hardcoded (needs a legacy handler).
3. Mirror pyfa's `class Effect<N>(BaseEffect)` (allowed: GPL like pyfa), register the id in
   `LegacyEffectIDs` (and `LegacyHandledEffectIds` if the effect also has modifierInfo).
4. Add a fixture covering it; `bash scripts/gate.sh`; `go run ./tools/diff -source=all`.
5. Same change in `eve-fit-engine` (§1); then `go run ./tools/drift -update`.

## 9. Hard rules

- **Never regress fixture parity (662/0)**, for any reason, diff harness included.
- **Never grow `known-diffs.json` to turn a red run green.** A fix that resolves an accepted diff deletes
  its entry in the same commit.
- **Every engine change in both repos** (§1).
- **`data/` only via `tools/syncdata`**, generated files only via `tools/stacking`.
- **The pin only moves with the full procedure** (§7), never `PYFA_REF` alone.
- **Workbench fetch stays at concurrency 2**; `.workbench/` is never redistributed.
- **Match pyfa, cite pyfa** in comments when mirroring `eos` behaviour.
- Commands, validation flows or parity invariants changed → update this file and `CLAUDE.md` in the same
  PR.
- Commit style: Conventional Commits (`fix(capacitor): …`, `test(diff): …`, `ci: …`), parity result in
  the body.

## 10. CI workflows

| workflow | trigger | does |
|---|---|---|
| `ci.yml` | push to main/master + PR | gofmt, vet, lint, stacking from JSON, tests, GOAMD64=v3, 662/0, data drift |
| `diff-parity.yml` | Mondays 07:30 UTC + manual | pyfa oracle at the pin, stacking vs pyfa, harness on all corpora → issue |
| `pyfa-drift.yml` | Mondays 06:30 UTC + manual | pyfa effects.py drift → issue |
| `pyfa-pin-bump.yml` | Sundays 14:00 UTC + manual | follow eve-fit-engine's pin: regenerate, gate, recalibrate → commit or PR |
| `sde-sync.yml` | daily 06:30 UTC + manual | follow eve-fit-engine releases: sync data/, gate, tag (or issue) |
| `release.yml` | push to main (code paths) + manual | twin-aligned version, sync data/ if reaching npm, gate, tag; PR = version preview only |

## 11. Licence

GPL-3.0-or-later, a declared derivative of Pyfa: porting Pyfa handler code is permitted. Contributions are
accepted under the same licence.
