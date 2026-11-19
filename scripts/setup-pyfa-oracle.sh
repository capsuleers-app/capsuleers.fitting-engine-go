#!/usr/bin/env bash
# Set up the pyfa-org/Pyfa headless oracle used by the differential parity
# harness (tools/diff) and by the stacking generator (tools/stacking). Fetches
# pyfa at the pinned commit, creates a venv with the minimal eos deps, stubs wx
# (config.py imports only wx.Colour; eos compute is wx-free), and builds the
# gamedata eve.db. Idempotent — safe to re-run.
#
#   bash scripts/setup-pyfa-oracle.sh            # set up / verify at the pin in ./PYFA_REF
#   bash scripts/setup-pyfa-oracle.sh --rebuild  # force-rebuild eve.db
#   PYFA_REF=<commit|branch> bash scripts/setup-pyfa-oracle.sh   # override the pin
#
# Everything lands under .pyfa/ (gitignored). Not part of the module.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PYFA="$ROOT/.pyfa"
VENV="$PYFA/.venv"
STUBS="$PYFA/_oracle_stubs"
PY="${PYTHON:-python3}"

echo "[oracle] root: $PYFA"

# 1. fetch pyfa-org/Pyfa at the pin. The pin lives in ./PYFA_REF (the SAME commit
#    eve-fit-engine pins in its diff-parity.yml); PYFA_REF in the environment
#    overrides it. Only that commit is fetched (--depth 1): pyfa's full history is
#    over a gigabyte and nothing here needs it.
PYFA_REF="${PYFA_REF:-$(tr -d '[:space:]' < "$ROOT/PYFA_REF" 2>/dev/null || true)}"
PYFA_URL="https://github.com/pyfa-org/Pyfa.git"
REPIN=0
fetch_ref() {
  ( cd "$PYFA" && git fetch --quiet --depth 1 origin "${1:-master}" && git checkout --quiet --detach FETCH_HEAD )
}
if [ ! -d "$PYFA/eos" ]; then
  echo "[oracle] fetching pyfa-org/Pyfa @ ${PYFA_REF:-master (unpinned)}..."
  mkdir -p "$PYFA"
  ( cd "$PYFA" && git init --quiet && git remote add origin "$PYFA_URL" )
  fetch_ref "$PYFA_REF"
else
  # An EXISTING checkout must be re-pinned when PYFA_REF has moved. Skipping this
  # silently ran the harness against the previous oracle after a pin bump — the
  # diff then reports staticdata drift that the new pin already fixes.
  CUR="$(cd "$PYFA" && git rev-parse HEAD 2>/dev/null || echo '')"
  if [ -n "$PYFA_REF" ] && [ "$CUR" != "$PYFA_REF" ] && [ "${CUR:0:${#PYFA_REF}}" != "$PYFA_REF" ]; then
    echo "[oracle] re-pinning $(echo "$CUR" | cut -c1-9) -> $PYFA_REF"
    fetch_ref "$PYFA_REF"
    REPIN=1   # staticdata changed under us: eve.db MUST be rebuilt
  else
    echo "[oracle] pyfa already checked out ($(cd "$PYFA" && git rev-parse --short HEAD 2>/dev/null || echo '?'))"
  fi
fi

# 2. venv + minimal eos/db_update deps (no wx, no numpy, no matplotlib)
if [ ! -x "$VENV/bin/python" ]; then
  echo "[oracle] creating venv..."
  "$PY" -m venv "$VENV"
fi
#    The SQLAlchemy version is READ FROM THE PIN, not hardcoded: pyfa moved to
#    SQLAlchemy 2.x (2026-09, `attribute_keyed_dict`), and the old hardcoded
#    1.4.50 made every pin bump past that point die in db_update.py — the
#    weekly pyfa-pin-bump went red on 2026-09-27 and the pin froze at 8b04f3b
#    while CCP shipped a drone rework, so the diff filed the rework as engine
#    bugs. Pins without a pyproject.toml (pre-uv pyfa) keep 1.4.50.
SQLA_SPEC="sqlalchemy==1.4.50"
if [ -f "$PYFA/pyproject.toml" ]; then
  PINNED="$(grep -oE '"sqlalchemy==[0-9][0-9.a-z]*"' "$PYFA/pyproject.toml" | head -1 | tr -d '"')"
  [ -n "$PINNED" ] && SQLA_SPEC="$PINNED"
fi
echo "[oracle] installing deps ($SQLA_SPEC)..."
"$VENV/bin/pip" install --quiet --upgrade pip >/dev/null 2>&1 || true
"$VENV/bin/pip" install --quiet "$SQLA_SPEC" logbook python-dateutil pyyaml cryptography

# 3. wx stub — config.py imports wx only for wx.Colour (UI slot colours).
mkdir -p "$STUBS"
cat > "$STUBS/wx.py" <<'PY'
# Headless stub: pyfa's config.py imports wx only for wx.Colour. eos compute
# never touches wx (grep -rl wx eos/ == 0). This satisfies the import.
class Colour:
    def __init__(self, *a, **k): pass
def __getattr__(name):
    return type(name, (), {'__init__': lambda self, *a, **k: None})
PY

# 4. build gamedata eve.db from the repo's staticdata
if [ "${1:-}" = "--rebuild" ] || [ "$REPIN" = "1" ]; then rm -f "$PYFA/eve.db"; fi
if [ ! -f "$PYFA/eve.db" ]; then
  echo "[oracle] building eve.db (this takes ~30-60s)..."
  ( cd "$PYFA" && PYTHONPATH="$STUBS" "$VENV/bin/python" db_update.py )
else
  echo "[oracle] eve.db already built ($(du -h "$PYFA/eve.db" | cut -f1))"
fi

# 5. smoke: compute a bare Rifter to confirm the oracle runs
echo "[oracle] smoke test..."
( cd "$PYFA" && PYTHONPATH="$STUBS" "$VENV/bin/python" - <<'PY'
import sys; sys._called_from_test = True
import eos.db
from eos.saveddata.fit import Fit
from eos.saveddata.ship import Ship
from eos.saveddata.character import Character
f = Fit(); f.ship = Ship(eos.db.getItem(587)); f.character = Character.getAll5()
f.calculateModifiedAttributes()
cpu = f.ship.getModifiedItemAttr("cpuOutput")
assert abs(cpu - 162.5) < 1, f"unexpected Rifter cpuOutput {cpu}"
print(f"[oracle] OK — Rifter cpuOutput={cpu}")
PY
)
echo "[oracle] ready."
