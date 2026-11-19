# Differential Pyfa-parity harness (`go run ./tools/diff`)

Finds where this engine disagrees with **pyfa-org/Pyfa**: it builds a corpus of fits, computes every
stat with both engines and prints every difference. No screenshots. It is the Go counterpart of
`npm run diff` in eve-fit-engine and builds the same `generated` and `implants` corpora (the generator is
a literal port, RNG included; checked spec by spec with `-dump-corpus` at the extraction), so the two
repos' reports are directly comparable.

## Setup (one-time)

```bash
bash scripts/setup-pyfa-oracle.sh          # pyfa at PYFA_REF into .pyfa/, venv, eve.db (~450 MB)
bash scripts/setup-pyfa-oracle.sh --rebuild   # rebuild eve.db
```

The oracle is `oracle/pyfa_oracle.py`, run with the `.pyfa` venv: pyfa's `eos` is wx-free (a stub covers
`config.py`'s `wx.Colour`), only SQLAlchemy (version read from the pin's `pyproject.toml`) and a few small
deps are installed. `$PYFA_DIR` points the harness at another checkout.

## Corpora (`-source`)

| `-source` | corpus | size (SDE 3720952) | built by |
|---|---|---|---|
| `generated` (default) | 4 synthetic fits per published ship (bonused / non-bonused / t2 / mixed) | 1 720 | `generator.go` |
| `implants` | every implant SET + ship-affecting booster, drawbacks off AND on; full matrix on the first hull + a rotating slice per hull | ~1 870 | `sources.go` |
| `workbench` | 2-3 REAL fits per hull from EVE Workbench (`.workbench/corpus.json`, `go run ./tools/workbench`) | ~900 | `sources.go` |
| `all` | every corpus in one run | | |

The generated corpus is assembled by our own slot/charge predicates, so alone it only exercises what the
engine already believes is fittable; the other two exist because that blind spot found real bugs.

Coverage is DECLARED, never silent: unparseable fits, an absent corpus, items the pinned pyfa doesn't
know (`{"op":"known"}` preflight) are printed as notes. A corpus that quietly shrinks while still
reporting "no differences" is the failure this harness exists to prevent.

## Run

```bash
go run ./tools/diff                         # generated corpus, all published ships
go run ./tools/diff -source=all             # every corpus
go run ./tools/diff -ships=587,29990        # specific ships (typeIDs)
go run ./tools/diff -group=Loki -only=bonused -stats=offense,defense
go run ./tools/diff -tol=0.02 -json         # custom tolerance / machine output
go run ./tools/diff -dump-corpus=c.json     # write the oracle specs and stop (compare corpora)
```

Comparison: relative tolerance (default 1 %) plus 0.01 absolute; booleans exact; a value on one side and
null on the other is a diff when the present side exceeds 0.01. Units: resists in %, targeting/drone
range in km, cap stable in %, everything else raw EVE units.

Exit **0** iff there are no UNEXPECTED diffs and no oracle staticdata drift, non-zero otherwise (as a
built binary: 1 = diffs or drift, 2 = harness error; `go run` reports any failure as 1).

## Known / accepted differences

`test/diff/known-diffs.json` holds the accepted set, keyed by `(ship, fitType, key)`, each with its ROOT
CAUSE (a pyfa float/modelling/per-ship quirk whose fix would regress the fixture suite). Accepted diffs
are reported but don't fail the run; `-strict` fails on every diff. The list is empty today.

**Never grow it to turn a red run green, never regress the fixture suite (662/0) to satisfy the
harness.** A fix that resolves an accepted diff deletes its entry in the same commit.

## Oracle drift vs engine diffs

pyfa's staticdata trails CCP's. Two cases make a fit incomparable, and both are reported under
`=== ORACLE SDE DRIFT` and skipped (`oracle-skipped`), never diffed:

- our `data/` has a type the pinned pyfa doesn't: the oracle would fit one item FEWER (`dropped`);
- pyfa knows the type but with different BASE attributes (CCP changed it after the pin): every attribute
  of our bundle's type plus mass/volume/capacity is compared with pyfa's eve.db (oracle op `attrs`),
  drift if pyfa lacks it or `|ours - pyfa| > 1e-6 * max(1, |pyfa|)`; 1367 maxSubSystems is excluded
  (pyfa forces it from 5 to 4 on purpose). The report lists type, attribute, our value and pyfa's.

The rule is in `tools/diff/drift.go` (TS: `test/diff/attr-drift.mjs`), identical in both repos. Drift
makes the run exit non-zero, like unexpected diffs, but under its own heading. The fix is a newer pin, which follows eve-fit-engine (`pyfa-pin-bump.yml`); by hand: CONTRIBUTING.md §7.

## Pin recalibration

`go run ./tools/diff -recalibrate` (generated corpus, like eve-fit-engine's `diff:recalibrate`): reads the
`.pyfa` HEAD as the pin, runs strict, keeps the reasons of entries that still diff, drops resolved ones,
marks new ones `PENDING REVIEW`, rewrites `known-diffs.json` and `PYFA_REF`. Exits 1 while anything is
pending: a human classifies each one (real bug → fix the engine in both repos; pyfa quirk → write the root
cause).
