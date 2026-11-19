package checks

import (
	"fmt"
	"math"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

// TypeLookup e' `{ getType(id): SdeType | undefined }`.
type TypeLookup interface {
	GetType(id int) (*types.SdeType, bool)
}

func readAttr(t *types.SdeType, id int) (float64, bool) { return attrOf(t, id) }

// TypeFitsSlotType: il type dichiara un effetto di slot corrispondente.
func TypeFitsSlotType(t *types.SdeType, slot types.SlotType) bool {
	for _, e := range t.Effects {
		if mapped, ok := engine.SlotEffectToSlotType(e.ID); ok && mapped == slot {
			return true
		}
	}
	return false
}

func hasWeaponKind(t *types.SdeType, kind engine.WeaponEffectKind) bool {
	for _, e := range t.Effects {
		if k, ok := engine.WeaponKindOf(e.ID); ok && k == kind {
			return true
		}
	}
	return false
}

func IsTurretWeapon(t *types.SdeType) bool    { return hasWeaponKind(t, engine.WeaponTurret) }
func IsMissileLauncher(t *types.SdeType) bool { return hasWeaponKind(t, engine.WeaponMissile) }
func IsSmartBomb(t *types.SdeType) bool       { return hasWeaponKind(t, engine.WeaponSmartbomb) }

func positiveRounded(t *types.SdeType, ids []int) []int {
	var out []int
	for _, id := range ids {
		if v, ok := readAttr(t, id); ok && v > 0 {
			out = append(out, int(jsmath.Round(v)))
		}
	}
	return out
}

func ShipGroupRestrictions(t *types.SdeType) []int {
	return positiveRounded(t, engine.CanFitShipGroupAttrs)
}
func ShipTypeRestrictions(t *types.SdeType) []int {
	return positiveRounded(t, engine.CanFitShipTypeAttrs)
}

// MaxGroupFittedFor / MaxTypeFittedFor: tetti per nave (ok=false = nessun tetto).
func MaxGroupFittedFor(mod *types.SdeType) (int, bool) {
	if v, ok := readAttr(mod, engine.AttrMaxGroupFitted); ok && v > 0 {
		return int(jsmath.Round(v)), true
	}
	return 0, false
}

func MaxTypeFittedFor(mod *types.SdeType) (int, bool) {
	if v, ok := readAttr(mod, engine.AttrMaxTypeFitted); ok && v > 0 {
		return int(jsmath.Round(v)), true
	}
	return 0, false
}

// FreeFitGroupSlotsFor: quanti altri moduli di questo gruppo/type si possono ancora montare (+Inf senza tetti).
func FreeFitGroupSlotsFor(mod *types.SdeType, fittedModules []int, dataset TypeLookup) float64 {
	groupCap, hasGroup := MaxGroupFittedFor(mod)
	typeCap, hasType := MaxTypeFittedFor(mod)
	if !hasGroup && !hasType {
		return math.Inf(1)
	}
	groupCount, typeCount := 0, 0
	for _, typeID := range fittedModules {
		if hasType && typeID == mod.ID {
			typeCount++
		}
		if hasGroup {
			if t, ok := dataset.GetType(typeID); ok && t.GroupID == mod.GroupID {
				groupCount++
			}
		}
	}
	free := math.Inf(1)
	if hasGroup {
		free = jsmath.Min(free, float64(groupCap-groupCount))
	}
	if hasType {
		free = jsmath.Min(free, float64(typeCap-typeCount))
	}
	return jsmath.Max(0, free)
}

// FitContextInfo e' il parametro opzionale fitContext di canFitModuleOnShip.
type FitContextInfo struct {
	FittedModules []int
	Dataset       TypeLookup
}

// CanFit e' il risultato di CanFitModuleOnShip.
type CanFit struct {
	OK     bool
	Reason string // "" = assente
}

// CanFitModuleOnShip: il modulo e' montabile in questo slot su questa nave? ship nil = nessun controllo di nave.
func CanFitModuleOnShip(mod, ship *types.SdeType, slot types.SlotType, fitContext *FitContextInfo) CanFit {
	if !TypeFitsSlotType(mod, slot) {
		return CanFit{false, fmt.Sprintf("Module does not declare a %s slot effect", slot)}
	}
	if ship == nil {
		return CanFit{OK: true}
	}
	allowedGroups := ShipGroupRestrictions(mod)
	allowedTypes := ShipTypeRestrictions(mod)
	if len(allowedGroups) > 0 || len(allowedTypes) > 0 {
		groupOk, typeOk := false, false
		for _, g := range allowedGroups {
			if g == ship.GroupID {
				groupOk = true
			}
		}
		for _, t := range allowedTypes {
			if t == ship.ID {
				typeOk = true
			}
		}
		if !groupOk && !typeOk {
			return CanFit{false, "Ship hull/group not allowed by this module"}
		}
	}
	if slot == types.SlotRIG {
		modSize, hasM := readAttr(mod, engine.AttrRigSize)
		shipSize, hasS := readAttr(ship, engine.AttrRigSize)
		if hasM && hasS && jsmath.Round(modSize) != jsmath.Round(shipSize) {
			return CanFit{false, "Rig size mismatch"}
		}
	}
	if slot == types.SlotSUBSYSTEM {
		if fitsTo, ok := readAttr(mod, 1380); ok && fitsTo > 0 && jsmath.Round(fitsTo) != float64(ship.ID) {
			return CanFit{false, "Subsystem locked to a different T3C hull"}
		}
	}
	if slot == types.SlotHI {
		if IsTurretWeapon(mod) {
			turrets, _ := readAttr(ship, engine.AttrTurretHardpoints)
			if turrets <= 0 {
				return CanFit{false, "Ship has no turret hardpoints"}
			}
		} else if IsMissileLauncher(mod) {
			launchers, _ := readAttr(ship, engine.AttrLauncherHardpoints)
			if launchers <= 0 {
				return CanFit{false, "Ship has no launcher hardpoints"}
			}
		}
	}
	if fitContext != nil {
		free := FreeFitGroupSlotsFor(mod, fitContext.FittedModules, fitContext.Dataset)
		if free <= 0 {
			cap, ok := MaxTypeFittedFor(mod)
			if !ok {
				cap, _ = MaxGroupFittedFor(mod)
			}
			return CanFit{false, fmt.Sprintf("Only %d of this module type can be fitted to a ship", cap)}
		}
	}
	return CanFit{OK: true}
}

func ChargeGroupsForModule(mod *types.SdeType) []int {
	return positiveRounded(mod, engine.ChargeGroupAttrs)
}

func ModuleAcceptsAnyCharge(mod *types.SdeType) bool { return len(ChargeGroupsForModule(mod)) > 0 }

// ModuleAcceptsChargeType: gruppo ammesso e taglia carica non superiore a quella del modulo.
func ModuleAcceptsChargeType(mod, charge *types.SdeType) bool {
	allowed := ChargeGroupsForModule(mod)
	if len(allowed) == 0 {
		return false
	}
	found := false
	for _, g := range allowed {
		if g == charge.GroupID {
			found = true
		}
	}
	if !found {
		return false
	}
	modSize, hasM := readAttr(mod, engine.AttrChargeSize)
	chargeSize, hasC := readAttr(charge, engine.AttrChargeSize)
	if hasM && hasC && jsmath.Round(chargeSize) > jsmath.Round(modSize) {
		return false
	}
	return true
}

const (
	onlineEffectID         = 16
	overloadEffectCategory = 5
)

func activeEffectCategory(c int) bool { return c == 1 || c == 2 || c == 3 }

// IsActivatableModule: un effetto di categoria attiva (escluso `online`).
func IsActivatableModule(mod *types.SdeType, effects *types.Table[types.SdeEffect]) bool {
	for _, e := range mod.Effects {
		if e.ID == onlineEffectID {
			continue
		}
		if eff, ok := effects.Get(e.ID); ok && activeEffectCategory(eff.EffectCategoryID) {
			return true
		}
	}
	return false
}

func IsOverloadableModule(mod *types.SdeType, effects *types.Table[types.SdeEffect]) bool {
	for _, e := range mod.Effects {
		if eff, ok := effects.Get(e.ID); ok && eff.EffectCategoryID == overloadEffectCategory {
			return true
		}
	}
	return false
}

// EffectiveModuleState normalizza lo stato richiesto in uno stato raggiungibile dal modulo.
func EffectiveModuleState(mod *types.SdeType, effects *types.Table[types.SdeEffect], state types.ModuleState) types.ModuleState {
	if state == types.StateOffline || state == types.StateOnline {
		return state
	}
	if !IsActivatableModule(mod, effects) {
		return types.StateOnline
	}
	if state == types.StateOverload && !IsOverloadableModule(mod, effects) {
		return types.StateActive
	}
	return state
}

// NextModuleState: il prossimo stato nel ciclo di click.
func NextModuleState(mod *types.SdeType, effects *types.Table[types.SdeEffect], state types.ModuleState) types.ModuleState {
	cur := EffectiveModuleState(mod, effects, state)
	if !IsActivatableModule(mod, effects) {
		if cur == types.StateOffline {
			return types.StateOnline
		}
		return types.StateOffline
	}
	if cur == types.StateOnline {
		return types.StateActive
	}
	if cur == types.StateActive {
		if IsOverloadableModule(mod, effects) {
			return types.StateOverload
		}
		return types.StateOnline
	}
	return types.StateOnline
}

func defaultOfflineActivationGroup(g int) bool { return g == 330 || g == 4117 }

// DefaultStateForModule: ACTIVE per i moduli attivabili, ONLINE altrimenti (e per i gruppi cloak/jump drive).
func DefaultStateForModule(mod *types.SdeType, effects *types.Table[types.SdeEffect]) types.ModuleState {
	if defaultOfflineActivationGroup(mod.GroupID) {
		return types.StateOnline
	}
	if IsActivatableModule(mod, effects) {
		return types.StateActive
	}
	return types.StateOnline
}

// FreeHardpointsFor: hardpoint (e tetti di gruppo) ancora liberi per il modulo.
func FreeHardpointsFor(mod, ship *types.SdeType, fittedHiModules []int, dataset TypeLookup) float64 {
	turretCap, _ := readAttr(ship, engine.AttrTurretHardpoints)
	launcherCap, _ := readAttr(ship, engine.AttrLauncherHardpoints)
	turretUsed, launcherUsed := 0.0, 0.0
	for _, id := range fittedHiModules {
		t, ok := dataset.GetType(id)
		if !ok {
			continue
		}
		if IsTurretWeapon(t) {
			turretUsed++
		} else if IsMissileLauncher(t) {
			launcherUsed++
		}
	}
	var hardpointFree float64
	if IsTurretWeapon(mod) {
		hardpointFree = jsmath.Max(0, turretCap-turretUsed)
	} else if IsMissileLauncher(mod) {
		hardpointFree = jsmath.Max(0, launcherCap-launcherUsed)
	} else {
		hardpointFree = math.Inf(1)
	}
	groupFree := FreeFitGroupSlotsFor(mod, fittedHiModules, dataset)
	return jsmath.Min(hardpointFree, groupFree)
}

// DefaultModeTypeIDFor: vedi engine.DefaultModeTypeIDFor.
func DefaultModeTypeIDFor(ship *types.SdeType, dataset *types.FittingDataset) (int, bool) {
	return engine.DefaultModeTypeIDFor(ship, dataset)
}
