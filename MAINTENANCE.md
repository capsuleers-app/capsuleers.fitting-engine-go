# Maintaining parity with EVE + Pyfa

Four streams, each with its own automation. The first two come from `eve-fit-engine` by design (aligned
versions); the last two are this repo's own checks against Pyfa.

## 1. EVE SDE changes (CCP balance patches): automatic

New ships/modules/skills and changed attributes are data. `eve-fit-engine` rebuilds its bundle from CCP
daily and publishes a patch when it changes; [`sde-sync.yml`](.github/workflows/sde-sync.yml) follows it
here (sync `data/` from the npm tarball, gate, tag). Data-driven `modifierInfo` effects work with no code
change. Details: [RELEASE.md](RELEASE.md).

**When the gate goes red, this module stops following.** The usual cause is a new effect with empty
`modifierInfo`, referenced by a fittable type and registered nowhere: `TestEffectCoverage`
(`go test ./internal/engine -run TestEffectCoverage -v`) reports it as a "truly silent" gap.
`eve-fit-engine` hits the same gap at the same moment; the handler (in `LegacyEffectIDs`) or the
documented exclusion (in `OutOfScopeEffectIDs`) goes into BOTH repos.

## 2. Engine fixes in eve-fit-engine: manual, both repos

Every parity fix is made in both repositories, normally as two PRs opened and merged together (procedure
in [CONTRIBUTING.md](CONTRIBUTING.md#fixes-in-both-repos)). A code-only `eve-fit-engine` release makes
`sde-sync` open an issue: check the fixes are here, then run `release.yml`.

## 3. Pyfa hardcoded-effect changes: detect automatically, port by hand

Effects with empty `modifierInfo` live in hand-written Python in pyfa's `eos/effects.py`
(`class Effect<N>(BaseEffect)`); this engine mirrors the ones that matter as legacy handlers.

```bash
go run ./tools/drift            # pyfa master effects.py vs tools/drift/effects-snapshot.json + LegacyEffectIDs
go run ./tools/drift -update    # advance the baseline once handled
```

[`pyfa-drift.yml`](.github/workflows/pyfa-drift.yml) runs it on Mondays (as a built binary: exit 2 =
drift, which `go run` would flatten to 1) and opens an issue. For each added/changed effect: data-driven (no action) or needs a handler (port it; this module is
GPL like Pyfa). Then the fixture suite, the diff harness, `-update`. The snapshot format is the same as
`eve-fit-engine`'s `scripts/effects-snapshot.json`.

## 4. The pyfa pin: follows eve-fit-engine

The diff oracle and the stacking registry are calibrated against ONE pyfa commit, in [`PYFA_REF`](PYFA_REF),
always the same commit `eve-fit-engine` pins. [`pyfa-pin-bump.yml`](.github/workflows/pyfa-pin-bump.yml)
(Sundays 14:00 UTC, two hours after `eve-fit-engine`'s own bump) reads that pin and, if it moved: rebuilds
the oracle, regenerates the stacking registry, runs the gate, recalibrates `test/diff/known-diffs.json`,
and commits only when there is nothing to classify (otherwise a PR labelled `diff-parity` and a red run).
`diff-parity.yml` warns when the two pins differ. Manual procedure:
[docs/claude/parity.md](docs/claude/parity.md).

## Differential harness: the broad net

```bash
bash scripts/setup-pyfa-oracle.sh      # one-time: pyfa at PYFA_REF into .pyfa, venv, eve.db
go run ./tools/diff -source=all        # generated + implants + workbench; exit 0 iff no UNEXPECTED diffs
```

[`diff-parity.yml`](.github/workflows/diff-parity.yml) runs it on Mondays and opens an issue on unexpected
diffs or oracle staticdata drift. Compare with `eve-fit-engine`'s report of the same week: the same
finding in both points to pyfa/SDE or a shared bug (fix both), a finding only here points to this port.
