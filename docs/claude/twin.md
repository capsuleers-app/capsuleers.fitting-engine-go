# Gemello TS (eve-fit-engine)

Leggi se porti un fix da/verso il repo TS (`../capsuleers.fitting-engine`, npm `eve-fit-engine`), tocchi
tabelle scritte a mano, fixture o harness. Procedura completa per i contributori: CONTRIBUTING.md §1.

## Assetto
- PARI GRADO: ognuno si valida solo contro Pyfa (fixture 662/0 + harness vs pyfa headless allo stesso
  `PYFA_REF`). Non esiste più l'oracolo TS↔Go (rimosso all'estrazione): nessuna CI confronta i due.
- Versioni allineate: tag Go `vX.Y.Z` = npm `X.Y.Z`, `data/` identico al tarball (tools/syncdata).
- Mappa file TS ↔ Go: `docs/mapping.md` (un file Go per file TS, stessi nomi).

## Regola operativa: ogni fix in entrambi
1. Fix nel repo dove nasce, con la prova (fixture, voce di diff, riferimento al sorgente pyfa).
2. Porto letterale nell'altro (mapping.md, regole di CONTRIBUTING.md §6).
3. Gate in entrambi: qui `bash scripts/gate.sh` + `go run ./tools/diff -source=all`; nel TS
   `npm run test:pyfa`, `npm run audit:coverage`, `npm run diff -- --source=all`.
4. Due PR collegate tra loro, con i risultati di entrambi; si uniscono insieme. Il rilascio TS successivo
   porta il fix in entrambi gli artefatti (sde-sync / release.yml).
- Riguarda anche: tabelle doppie (constants, modifier tables, stacking groups, profiles), fixture
  (`test/parity/fixtures.json`, stessi valori attesi), harness (generatore/schema/tolleranze: corpus
  `generated` e `implants` identici a quelli TS, verificati spec per spec all'estrazione con `-dump-corpus`), oracolo python.
- Il registro stacking deve uscire identico nei due repo allo stesso pin (`generated/stacking-penalised.json`).

## Stato all'estrazione (2026-11-19, v0.2.26, pin b79cd90)
- Corpus `generated` identico (1720/1720 spec), `implants` = sottoinsieme ordinato filtrato dal preflight
  `known` come nel TS. Risultati confrontati con `npm run diff` sui 6 scafi che divergono da pyfa
  (identici); il corpus `workbench` non è stato confrontato spec per spec.
- Il generatore chiamava `canFitModuleOnShip` con gli argomenti invertiti (filtro mai attivo): corretto
  in entrambi a 2026-11 (fix/parity-twin-followups), con gli hardpoint presi dal fit calcolato (T3C);
  corpus rigenerato, ancora identico spec per spec tra i due repo.
- Dal ramo fix/parity-twin-followups il TS legge `test/parity/fixtures.json` (identico byte per byte
  al nostro) e `stacking:build` scrive anche `generated/stacking-penalised.json` (identico al nostro;
  `--check --from-json` nel suo `ci.yml`).
