// Porzione di src/fitChecks.ts necessaria a computeFit (il resto arriva con E5, in fitengine/checks).
package engine

import (
	"strings"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

// Gruppo di navi che portano sempre un modo tattico.
const tacticalDestroyerGroup = 1305

// L'unica altra nave con modi (pyfa la tratta per nome).
func modedHullName(name string) bool { return name == "Anhinga" }

// Gruppo che contiene gli item dei modi T3D ("<Ship> Defense Mode", ...).
const shipModifiersGroup = 1306

// DefaultModeTypeIDFor: il modo a cui un Tactical Destroyer ricade quando il fit non ne nomina nessuno
// (il type con id minore fra i moduli del gruppo 1306 il cui nome inizia col nome della nave).
func DefaultModeTypeIDFor(ship *types.SdeType, dataset *types.FittingDataset) (int, bool) {
	if ship.GroupID != tacticalDestroyerGroup && !modedHullName(ship.Name) {
		return 0, false
	}
	prefix := strings.ToLower(ship.Name)
	if prefix == "" {
		return 0, false
	}
	var best *types.SdeType
	mods := &dataset.TypesByBucket[types.BucketModules]
	for i := 0; i < mods.Len(); i++ {
		_, t := mods.At(i)
		if t.GroupID != shipModifiersGroup {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(t.Name), prefix) {
			continue
		}
		if best == nil || t.ID < best.ID {
			best = t
		}
	}
	if best == nil {
		return 0, false
	}
	return best.ID, true
}

// IsTurretWeapon: il modulo ha un effetto classificato come arma a torretta (WEAPON_EFFECT_KIND == TURRET).
func IsTurretWeapon(t *types.SdeType) bool {
	for _, e := range t.Effects {
		if k, ok := WeaponKindOf(e.ID); ok && k == WeaponTurret {
			return true
		}
	}
	return false
}
