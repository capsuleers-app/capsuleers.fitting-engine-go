# Rilascio e dati

Leggi se tocchi `VERSION`, `data/`, `tools/syncdata` o `.github/workflows/`. Flussi completi:
`RELEASE.md`, `MAINTENANCE.md`.

## Regole
- Tag `vX.Y.Z` = npm `eve-fit-engine@X.Y.Z`; `VERSION` = X.Y.Z corrente. Sempre `v0.x` (un `v2+` richiede `/v2`).
  Un rilascio può passare AVANTI al gemello (versione non ancora su npm): `data/` resta quello dell'ultima npm
  e `tools/syncdata` (default da `VERSION`) segue la versione pubblicata più alta ≤ `VERSION`; quando il TS
  raggiunge quella versione le due si riallineano.
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
- `release.yml`: **AUTOMATICO a ogni push su `main`** (= ogni PR unita) + dispatch manuale (`version` vuota =
  regola automatica, `X.Y.Z` = esattamente quella, `data/` da quel tarball). Regola (`scripts/next-version.sh`):
  ultima eve-fit-engine su npm > mio tag più alto → la raggiungo e sincronizzo `data/` da quel tarball;
  altrimenti tag più alto + patch con `data/` dell'ultima npm. Gate rosso = niente tag. Prima del tag
  ricontrolla versione, tip di main e tag (mai sovrascritto); `main` mosso → run rossa, la run in coda
  rilascia il nuovo tip. PR: solo job `preview` che stampa la versione.
  Anti-loop (solo push): tip con `[auto]` nel soggetto (release, sde-sync, pyfa-pin-bump) o già con tag `v*`
  → niente rilascio. `paths-ignore`: `**.md`, `docs/**`, `.github/**`, `test/**`, `tools/**`, `scripts/**`,
  `oracle/**`, `PYFA_REF`, `LICENSE`, `NOTICE`, `.gitignore`. Concurrency `release` condivisa con `sde-sync`.
- `pyfa-pin-bump.yml`, `diff-parity.yml`, `pyfa-drift.yml`: parity.md / MAINTENANCE.md.
- Push e tag della CI con la GitHub App (`CAPSULEERS_BOT_APP_ID` variabile, `CAPSULEERS_BOT_PRIVATE_KEY`
  segreto, App installata sul repo; bypass del ruleset "solo PR" se presente). Etichette usate:
  `sde-sync`, `diff-parity`, `pyfa-drift`. Nessun token npm.
- Un agente non crea tag né fa push su main: preparare i file e lasciare il tag alla CI o all'utente.

## Dopo il tag
`GOPROXY=https://proxy.golang.org go list -m github.com/capsuleers-app/capsuleers.fitting-engine-go@vX.Y.Z`
(la CI lo fa già) per far comparire subito la versione su pkg.go.dev.
