# Security Policy

## Supported versions

`capsuleers.fitting-engine-go` follows a rolling-release model (tags `vX.Y.Z`, aligned with the npm
package `eve-fit-engine`). **Only the latest tagged version is supported.** Please reproduce any issue
against the most recent release before reporting it; older versions are not back-patched.

## Reporting a vulnerability

**Do not open a public GitHub issue for security problems.**

Report privately, through either channel:

- **GitHub Security Advisories**: the
  [*Report a vulnerability*](https://github.com/capsuleers-app/capsuleers.fitting-engine-go/security/advisories/new)
  button on this repository (preferred).
- **Email**: [info@capsuleers.app](mailto:info@capsuleers.app).

Please include:

- a description of the issue and its impact,
- the module version (`vX.Y.Z`) and Go version,
- a minimal reproduction (a fit input + the call you made),
- any panics, stack traces or logs.

We aim to acknowledge a report within **72 hours** and to provide a remediation plan or fix timeline
within **7 days**. Coordinated disclosure is appreciated: please give us a reasonable window to ship a fix
before any public write-up. A flaw that also affects `eve-fit-engine` is fixed in both.

## Scope

This is a pure computation library (Pyfa-parity fitting math) with no dependencies outside the Go
standard library. Things especially worth reporting:

- panics, unbounded resource consumption (CPU / memory), or denial of service triggered by crafted fit
  or EFT input passed to the public API,
- unsafe handling of attacker-controlled input (e.g. a crafted dataset passed to `bundle.LoadDataset`).

### Out of scope

- Incorrect stat numbers / Pyfa-parity mismatches: those are **correctness bugs**, not security issues;
  open a normal GitHub issue (see [CONTRIBUTING.md](CONTRIBUTING.md)).
- The embedded SDE being out of date: that is a maintenance stream, see [MAINTENANCE.md](MAINTENANCE.md).
- The repository tooling under `tools/` and `scripts/` (developer/CI only, not part of the library).
- Reports requiring a compromised build environment or physical access.
- Missing best-practice hardening with no demonstrable impact.
