# Engine

Leggi se tocchi `fitengine/` o `internal/`. Ogni modifica di comportamento va anche in eve-fit-engine
(twin.md). Percorsi relativi alla root del repo.

## Cosa è
Motore fitting **a parità con Pyfa**: `fitengine.ComputeFit(fit, dataset, opts)` → `ComputedFit`
(`Derived`: offense, defense, capacitor, tank, navigation, targeting, drones, projected, structure).
GPL-3.0-or-later, derivato dichiarato di pyfa `eos` (porting diretto ammesso).
- `fitengine` non usa fs né rete: il dataset si inietta. `fitengine/bundle` incorpora `data/`
  (`go:embed`) e offre `LoadBundledDataset`, `BuildAllVSkillProfile`, `ComputeFromEft`.
- Sotto-package pubblici: `eft` (parse/format), `export`, `checks`, `profiles`; `EncodeJSON` riproduce il
  JSON del TS campo per campo (tag `encode:"null"` = campo nil emesso come null).

## Architettura (`internal/`)
- `engine/engine.go` (`computeFit`): skill → nave → subsystem → impianti/booster (peer, poi outgoing) →
  pass `earliest` dei moduli → cariche (PRIMA del loro modulo: script/cristalli) → moduli → droni →
  fighter → mode → gestori legacy → derive. Le letture sono EAGER (quelle di pyfa lazy): una sorgente va
  applicata dopo tutto ciò che modifica gli attributi che inoltra (impianti 'Xyston' prima dei rig;
  `RuntimeEarliestEffectIds` = Rig Synchronizer per tutti i moduli prima).
- `engine/modifier_engine.go`: dispatcher `modifierInfo` + ~85 gestori `applyLegacy*` + tabelle
  (`modifier_engine_tables.go`, `num_tables.go`): scaling skill per bonus di scafo, bonus di ruolo FLAT,
  effetti scalati dalla sec status, gruppi di stacking. Target delle skill del proprietario: droni,
  fighter, cariche, MAI i moduli nave (fantasma +50% DPS sui moduli combo, come `Effect6556` di pyfa).
- `gen/stacking_penalised_gen.go` (GENERATO da `tools/stacking`): quali (effetto, attributo, ricevente)
  pyfa penalizza e in che `penaltyGroup`; la penalità è della CHIAMATA, non dell'attributo. Chiavi per
  ID attributo (anche non pubblicati, es. `warpSpeedMultiplier` 600). Override per handler pyfa trasposti
  (6001 ← Effect5998).
- `engine/stacking_groups.go`: effectID → penaltyGroup (a mano, doppia col TS); onorati tutti tranne `preMul`.
- `derived/`: offense, ehp, tank, capacitor (porto di `eos/capSim.py`: cicli interi, torrette a volley,
  cap booster `clipSize=0` = infinito, `moduleReactivationDelay` nel ciclo, costo di attivazione una volta
  per modulo anche con più effetti di scarica: Dual Afocal Light Laser I, 6633; niente stagger per chi ha
  l'effetto `turretFitted` 42, armi e anche laser minerari, come `mod.hardpoint == TURRET` di pyfa),
  structure, application.
- `engine/constants.go` (`LegacyEffectIDs`, `OutOfScopeEffectIDs`, attributi, categorie), `fit_context.go`,
  `item_state.go`, `modified_attribute.go`, `stacking.go`; `effects/` (weapon, ewar).
- `jsmath` (semantica JS dove Go differisce), `omap` (Map/Set con ordine di inserimento), `types`.

## Regole di codice
Porto letterale del TS (stesso ordine di operazioni), `float64(a*b) + c`, niente `range` su map nei
package dell'engine (lint bloccante), funzioni JS → `jsmath` (CONTRIBUTING.md §6). Gate dopo ogni
modifica: `bash scripts/gate.sh` + `go run ./tools/diff -source=all`.
