# capsuleers.fitting-engine-go

Motore di calcolo fitting EVE Online (navi e strutture Upwell) a parità con Pyfa, in Go: modulo
`github.com/capsuleers-app/capsuleers.fitting-engine-go`, `FittingDataset` + `Fit` → blocco stat completo.
Gemello PARI GRADO del pacchetto npm `eve-fit-engine` (repo TS `../capsuleers.fitting-engine`): stesse
versioni (`VERSION`), stesso `data/`, stesso pin pyfa (`PYFA_REF`); ognuno si confronta solo con Pyfa.
Percorsi relativi a questa cartella; orchestrazione in `../../CLAUDE.md`.

## Comandi
```bash
go vet ./... && go run ./tools/lint && go test ./...   # unit + parità fixture 662/0 + coverage + droni
go test ./test/parity -run TestPyfaParity -count=1 -v  # verifica "Summary: 662 passed, 0 failed." (rtk nasconde i FAIL)
bash scripts/gate.sh                                   # gate di rilascio completo (anche GOAMD64=v3 + drift data/)
bash scripts/setup-pyfa-oracle.sh                      # oracolo pyfa in .pyfa/ al commit di PYFA_REF
go run ./tools/diff -source=all                        # harness differenziale vs pyfa (generated|implants|workbench)
go run ./tools/diff -recalibrate                       # dopo un cambio di pin: known-diffs.json + PYFA_REF
go run ./tools/stacking [-check] [-from-json]          # registro stacking da pyfa pinnato (-fetch-sde prima)
go run ./tools/syncdata [-check|-version X.Y.Z|-latest]  # data/ dal tarball npm eve-fit-engine@VERSION
go run ./tools/drift [-update]                         # effects.py di pyfa vs snapshot + LegacyEffectIDs
go run ./tools/workbench                               # corpus fit reali (EVEWORKBENCH_API_KEY)
```
Comandi pesanti: `systemd-run --user --scope -p MemoryMax=6G …` e `go test -p 2`.

## Vincoli
- Licenza GPL-3.0-or-later (derivato di Pyfa `eos`, porting diretto ammesso); `LICENSE` intatta.
- **Ogni fix di parità/engine in ENTRAMBI i repo** (TS e Go), due PR collegate e unite insieme
  (CONTRIBUTING.md §1, mappa file `docs/mapping.md`). Niente oracolo TS↔Go: l'unico confronto è con Pyfa.
- **Mai regredire la parità fixture 662/0, mai allargare `test/diff/known-diffs.json`** per far passare
  `tools/diff`; un fix che risolve un diff accettato ne cancella la voce nello stesso commit.
- `data/` solo via `tools/syncdata` (mai a mano; `data/embed.go` è nostro). `internal/gen/*` e
  `generated/stacking-penalised.json` solo via `tools/stacking`, identici al TS allo stesso pin.
- Pin pyfa = quello di eve-fit-engine; si sposta solo con la procedura completa (docs/claude/parity.md).
- Engine: mai `range` su map Go (usa `internal/omap`), sempre `float64(a*b) + c` (lint bloccante).
- Nessuna dipendenza fuori dalla standard library.
- Tag `vX.Y.Z` = eve-fit-engine X.Y.Z, ma il rilascio è AUTOMATICO a ogni push su `main` (release.yml): versione =
  npm più recente se > del mio tag, altrimenti tag+patch (`scripts/next-version.sh`); `[auto]`/docs/CI/test/tools
  non rilasciano. Tag e push li fa la CI: un agente non li fa.
- `tools/workbench` a concorrenza 2; `.workbench/` è cache gitignorata di fit altrui, non ridistribuirla.
- Comandi, flussi di validazione/rilascio o invarianti cambiati → aggiorna CONTRIBUTING.md e questo file.
- Emblema in `docs/brand/`: non ritoccarlo qui.

## Aree (leggi solo se il task le tocca)
| area | file | leggi se |
|---|---|---|
| engine | docs/claude/engine.md | tocchi `fitengine/`, `internal/`: orchestratore, modificatori, stacking, derived |
| gemello TS | docs/claude/twin.md | porti un fix da/verso eve-fit-engine, tocchi tabelle doppie o fixture |
| parità | docs/claude/parity.md | tocchi `test/`, `tools/diff`, `tools/stacking`, oracolo/pin pyfa |
| rilascio | docs/claude/release.md | tocchi `VERSION`, `data/`, `tools/syncdata` o i workflow di CI/rilascio |
| marchio | ../../docs/claude/brand.md | tocchi l'emblema |
