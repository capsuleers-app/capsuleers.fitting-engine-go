# Parità con Pyfa: suite, known-diffs, pin

Leggi se tocchi `test/`, `tools/diff`, `tools/stacking`, `oracle/`, `PYFA_REF` o i workflow di parità.
Dettagli harness: `test/diff/README.md`.

## Due suite, un'invariante
1. **Fixture** (`go test ./test/parity`): 662 asserzioni su 21 fit curati vs screenshot Pyfa, All-V,
   profilo Uniform. Verità di riferimento e gate di rilascio. Verifica la riga
   `Summary: 662 passed, 0 failed.` (`-v`), non solo `ok`.
2. **Harness** (`go run ./tools/diff`): ogni stat vs pyfa headless (`oracle/pyfa_oracle.py` in `.pyfa/`,
   da `scripts/setup-pyfa-oracle.sh`) su tre corpus: `generated` 1720, `implants` ~1870, `workbench` ~900
   (`go run ./tools/workbench`). Exit 0 sse nessun diff INATTESO e nessuna deriva di staticdata.
   `test/diff/known-diffs.json` = diff accettati con causa radice (vuoto); `-strict` li fa fallire.

> **INVARIANTE: mai regredire 662/0 per far passare l'harness, mai allargare known-diffs.** Modifiche
> all'engine: girare entrambe le suite; un fix che risolve un diff accettato ne cancella la voce.

Misura all'estrazione (SDE 3720952, pin b79cd90, 2026-11-19): 4497/4503 fit uguali; 3 fit saltati per
deriva (98845 Imperial Navy Stasis Webifier assente dal pyfa pinnato) e 3 `fitting.cpuUsed` (Bestla,
Anhinga, Lachesis `mixed`, −9 CPU): risultato IDENTICO a `npm run diff` del TS sugli stessi dati, quindi
pin/SDE e non porting. Si risolvono col prossimo pin di eve-fit-engine.

## Pin pyfa
`PYFA_REF` = SEMPRE il commit pinnato da eve-fit-engine (`PYFA_REF:` nel suo `diff-parity.yml`).
Calibrati su quel commit: `PYFA_REF`, `test/diff/known-diffs.json`, `generated/stacking-penalised.json` +
`internal/gen/stacking_penalised_gen.go`. Si muovono insieme, mai il solo pin:

```bash
echo <commit-di-eve-fit-engine> > PYFA_REF
bash scripts/setup-pyfa-oracle.sh                              # ri-pin di .pyfa + rebuild eve.db
go run ./tools/stacking -fetch-sde && go run ./tools/stacking  # JSON identico a quello del TS
bash scripts/gate.sh                                           # 662/0
go run ./tools/diff -recalibrate     # PENDING REVIEW → classificare (bug → fix in entrambi i repo)
go run ./tools/diff -source=all      # exit 0
```
Automatico: `pyfa-pin-bump.yml` (domenica 14:00 UTC, dopo il bump TS delle 12:00) legge il pin TS,
rigenera, gate, ricalibra; commit solo senza voci da classificare, altrimenti PR `diff-parity` e run rossa.
`diff-parity.yml` avvisa se i due pin divergono.

## Generatore stacking (`tools/stacking`)
Porto riga per riga di `scripts/build-stacking-registry.mjs`: analizza `.pyfa/eos/effects.py` (chiamate
`boostItemAttr`/`filtered*` con `stackingPenalties`, `penaltyGroup`, contesti esenti), mappa i nomi su ID
con la SDE grezza (`.sde-src/dogmaAttributes.jsonl` + `dogmaEffects.jsonl`, da `-fetch-sde`), scrive JSON
+ Go. All'estrazione riproduce byte per byte il JSON del TS da pyfa b79cd90. `-check` (con pyfa) in
`diff-parity`, `-check -from-json` (senza pyfa) in `ci.yml`. Le 5 segnalazioni "handler drift" stampate
sono note (una ha l'override 6001).

## Note operative
- `.pyfa/` si costruisce con fetch `--depth 1` del solo commit (~450 MB con venv ed eve.db); SQLAlchemy
  letto dal `pyproject.toml` del pin. `$PYFA_DIR` punta harness e generatore a un'altra checkout.
- Un pyfa-pin-bump rosso non è innocuo: il pin si congela e i lunedì successivi arrivano falsi "engine
  diff" che sono solo oracolo vecchio. Leggerne il log prima delle issue.
