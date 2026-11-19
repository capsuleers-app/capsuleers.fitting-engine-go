# Rilascio e dati

Leggi se tocchi `VERSION`, `data/`, `tools/syncdata` o `.github/workflows/`. Flussi completi:
`RELEASE.md`, `MAINTENANCE.md`.

## Regole
- Tag `vX.Y.Z` = npm `eve-fit-engine@X.Y.Z`; `VERSION` = X.Y.Z corrente. Niente versioni solo-Go.
  Sempre `v0.x` (un `v2+` richiede `/v2` nel path del modulo).
- `data/` = `package/data/` del tarball npm di quella versione, sha512 verificato contro il registry
  (`go run ./tools/syncdata`); `data/embed.go` è nostro e resta. CI `data` fallisce su qualunque deriva.
  Scelta motivata: sono dati CCP, non codice; rigenerarli significherebbe un secondo generatore da tenere
  in parità e potrebbe puntare solo all'ultima SDE, mai a quella di una versione passata.
- Gate di rilascio: `bash scripts/gate.sh` (vet, lint, stacking da JSON, test, GOAMD64=v3, 662/0, drift).

## Workflow
- `ci.yml`: ogni push/PR (gofmt, vet, lint, stacking -from-json, test, v3, 662/0) + job `data`.
- `sde-sync.yml` (06:30 UTC, dopo `sde-refresh` del TS alle 05:00): npm più nuovo + `data/` cambiato →
  sync, gate, commit + tag + push + warm del proxy; gate rosso → issue `sde-sync`, run rossa, niente tag;
  `data/` invariato (rilascio solo codice del TS) → issue: verificare il porting e lanciare `release.yml`.
- `release.yml` (manuale, input `version`, vuoto = ultima npm): sync, `VERSION`, gate, commit, tag, push.
- `pyfa-pin-bump.yml`, `diff-parity.yml`, `pyfa-drift.yml`: parity.md / MAINTENANCE.md.
- Push e tag della CI con la GitHub App (`CAPSULEERS_BOT_APP_ID` variabile, `CAPSULEERS_BOT_PRIVATE_KEY`
  segreto, App installata sul repo; bypass del ruleset "solo PR" se presente). Etichette usate:
  `sde-sync`, `diff-parity`, `pyfa-drift`. Nessun token npm.
- Un agente non crea tag né fa push su main: preparare i file e lasciare il tag alla CI o all'utente.

## Dopo il tag
`GOPROXY=https://proxy.golang.org go list -m github.com/capsuleers-app/capsuleers.fitting-engine-go@vX.Y.Z`
(la CI lo fa già) per far comparire subito la versione su pkg.go.dev.
