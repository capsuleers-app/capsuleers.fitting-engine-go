# File map: eve-fit-engine (TS) ↔ capsuleers.fitting-engine-go

Where a change in one repo goes in the other (CONTRIBUTING.md §1). One Go file per TS file, same names
(snake_case here), same order of statements. Paths on the left are in
[eve-fit-engine](https://github.com/capsuleers-app/capsuleers.fitting-engine).

## Engine

| eve-fit-engine | this repo |
|---|---|
| `src/types.ts` | `internal/types/types.go`, `internal/types/derived.go` (re-exported by `fitengine/types.go`, `fitengine/compute.go`) |
| `src/index.ts` (public surface) | `fitengine/*.go` + subpackages |
| `src/node.ts` (`loadBundledDataset`, `buildAllVSkillProfile`, `computeFromEft`) | `fitengine/bundle/bundle.go` (+ `data/embed.go` for the embedded bundle) |
| `src/engine.ts` (`computeFit`) | `internal/engine/engine.go` + `fitengine/compute.go` (derive stats, public options) |
| `src/modifierEngine.ts` | `internal/engine/modifier_engine.go`, `modifier_engine_tables.go`, `num_tables.go` |
| `src/constants.ts` | `internal/engine/constants.go` |
| `src/fitContext.ts` | `internal/engine/fit_context.go` |
| `src/itemState.ts` | `internal/engine/item_state.go` |
| `src/modifiedAttribute.ts` | `internal/engine/modified_attribute.go` |
| `src/stacking.ts` | `internal/engine/stacking.go` |
| `src/stackingGroups.ts` (hand-written) | `internal/engine/stacking_groups.go` (hand-written) |
| `src/stackingPenalised.ts` (GENERATED) | `internal/gen/stacking_penalised_gen.go` (GENERATED); shared model `generated/stacking-penalised.json` |
| `src/derived/{offense,ehp,tank,capacitor,structure,application}.ts` | `internal/derived/{offense,ehp,tank,capacitor,structure,application}.go` |
| `src/effects/{weapon,ewar}.ts` | `internal/effects/{weapon,ewar}.go` |
| `src/fitChecks.ts` | `fitengine/checks/fit_checks.go` (+ `internal/engine/fit_checks.go` for what the engine itself needs) |
| `src/skillCheck.ts` | `fitengine/checks/skill_check.go` |
| `src/eft/parser.ts`, `src/eft/format.ts` | `fitengine/eft/parser.go`, `fitengine/eft/format.go` |
| `src/export.ts` | `fitengine/export/export.go` |
| `src/profiles.ts` | `fitengine/profiles/profiles.go` |
| `src/t3cVariant.ts` | `fitengine/t3c_variant.go` |
| `src/marketGroupTree.ts` | `fitengine/market_group_tree.go` |
| JS semantics (`Math.round`, `Number`, `toFixed`, string compare, `localeCompare`) | `internal/jsmath/*` |
| JS `Map`/`Set` insertion order | `internal/omap/omap.go` |
| `JSON.stringify` of the result | `fitengine/encode.go` (`EncodeJSON`) |

## Data, tests and tooling

| eve-fit-engine | this repo |
|---|---|
| `data/` (built by `npm run build:data`, shipped on npm) | `data/` (synced from the npm tarball: `tools/syncdata`) |
| `test/parity/run-pyfa-parity.ts` (`npm run test:pyfa`) | `test/parity/parity_test.go` + `test/parity/fixtures.json` |
| `test/effect-coverage-audit.ts` (`npm run audit:coverage`) | `internal/engine/coverage_test.go` |
| `test/drone-bandwidth.ts` (`npm run test:drones`) | `fitengine/drone_bandwidth_test.go` |
| `scripts/build-stacking-registry.mjs` (`npm run stacking:build`) | `tools/stacking` |
| `test/diff/run-diff.mjs`, `stat-schema.mjs` (`npm run diff`) | `tools/diff/main.go`, `schema.go` |
| `test/diff/fit-generator.mjs`, `fit-sources.mjs` | `tools/diff/generator.go`, `sources.go` |
| `test/diff/known-diffs.mjs` | `test/diff/known-diffs.json` |
| `scripts/recalibrate-pyfa-pin.mjs` (`npm run diff:recalibrate`) | `tools/diff -recalibrate` (`tools/diff/knowndiffs.go`) |
| `scripts/check-pyfa-drift.mjs` + `scripts/effects-snapshot.json` (`npm run drift`) | `tools/drift` + `tools/drift/effects-snapshot.json` |
| `scripts/fetch-workbench-corpus.mjs` (`npm run corpus:fetch`) | `tools/workbench` |
| `scripts/setup-pyfa-oracle.sh` (`npm run diff:setup`) | `scripts/setup-pyfa-oracle.sh` (pin read from `PYFA_REF`, shallow fetch) |
| `oracle/pyfa_oracle.py` | `oracle/pyfa_oracle.py` (same code; a logic change goes in both) |
| `PYFA_REF` in `.github/workflows/diff-parity.yml` | `PYFA_REF` (same commit) |
| (none: TS has no FMA/map-order hazards) | `tools/lint` |
