// Porting letterale di src/fitContext.ts: contesto di un fit durante il calcolo
// (domini dei modificatori, target, filtri per skill del proprietario).
package engine

import (
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

// MissileBoostSkills: skill che marcano una carica come MISSILE per i boost di danno dei moduli.
var missileBoostSkills = [2]int{3319, 3323}

// pyfaGroupFilteredEffects: PYFA_GROUP_FILTERED_EFFECTS (effectID -> insieme di groupID).
func pyfaGroupFilteredEffects(effectID int) (func(groupID int) bool, bool) {
	if effectID == 3495 {
		return func(g int) bool { return g == 52 || g == 65 }, true
	}
	return nil, false
}

// FitContextInit e' FitContextInit. Mode/Target/Structure nil = assenti.
type FitContextInit struct {
	Ship                      *ItemState
	Character                 *ItemState
	SkillLevels               *omap.Map[int, float64]
	Modules                   []*ItemState
	Drones                    []*ItemState
	Fighters                  []*ItemState
	Implants                  []*ItemState
	Boosters                  []*ItemState
	Subsystems                []*ItemState
	Mode                      *ItemState
	Target                    *ItemState
	Structure                 *ItemState
	SkillProfile              *types.SkillProfile
	Dataset                   *types.FittingDataset
	DisintegratorSpoolPercent float64
}

type FitContext struct {
	Ship                      *ItemState
	Character                 *ItemState
	SkillLevels               *omap.Map[int, float64]
	Modules                   []*ItemState
	Drones                    []*ItemState
	Fighters                  []*ItemState
	Implants                  []*ItemState
	Boosters                  []*ItemState
	Subsystems                []*ItemState
	Mode                      *ItemState
	Target                    *ItemState
	Structure                 *ItemState
	SkillProfile              *types.SkillProfile
	Dataset                   *types.FittingDataset
	DisintegratorSpoolPercent float64
	// ProjectedSources: sorgenti ostili proiettate in corso (riempito dai passaggi di proiezione).
	ProjectedSources []*ItemState
	// StoppedLocalEffectIDs: effetti locali che il dispatcher deve saltare (EffectStopper).
	StoppedLocalEffectIDs *omap.Set[int]
}

func NewFitContext(init FitContextInit) *FitContext {
	return &FitContext{
		Ship: init.Ship, Character: init.Character, SkillLevels: init.SkillLevels,
		Modules: init.Modules, Drones: init.Drones, Fighters: init.Fighters, Implants: init.Implants,
		Boosters: init.Boosters, Subsystems: init.Subsystems, Mode: init.Mode, Target: init.Target,
		Structure: init.Structure, SkillProfile: init.SkillProfile, Dataset: init.Dataset,
		DisintegratorSpoolPercent: init.DisintegratorSpoolPercent,
		StoppedLocalEffectIDs:     omap.NewSet[int](),
	}
}

// AllItems: tutti gli item che portano effetti o ricevono modifiche, nell'ordine del generatore TS.
func (c *FitContext) AllItems() []*ItemState {
	out := []*ItemState{c.Ship, c.Character}
	out = append(out, c.Modules...)
	for _, m := range c.Modules {
		if m.Charge != nil {
			out = append(out, m.Charge)
		}
	}
	out = append(out, c.Drones...)
	out = append(out, c.Fighters...)
	out = append(out, c.Implants...)
	out = append(out, c.Boosters...)
	out = append(out, c.Subsystems...)
	if c.Mode != nil {
		out = append(out, c.Mode)
	}
	if c.Target != nil {
		out = append(out, c.Target)
	}
	if c.Structure != nil {
		out = append(out, c.Structure)
	}
	return out
}

// SkillLevel: livello della skill, default 0 per le skill non addestrate.
func (c *FitContext) SkillLevel(skillTypeID int) float64 {
	if v, ok := c.SkillLevels.Get(skillTypeID); ok {
		return v
	}
	return 0
}

// ResolveDomain risolve la stringa di dominio SDE nell'item corrispondente (nil = nessuno).
func (c *FitContext) ResolveDomain(domain string, source *ItemState) *ItemState {
	switch domain {
	case "self", "itemID":
		return source
	case "shipID":
		return c.Ship
	case "char", "charID":
		return c.Character
	case "otherID":
		if source.Kind == KindModule {
			return source.Charge
		}
		if source.Kind == KindCharge {
			return c.FindChargeParent(source)
		}
		return nil
	case "targetID", "target":
		return c.Target
	case "structureID":
		return c.Structure
	}
	return nil
}

// FindChargeParent: il modulo che ha caricato questa carica.
func (c *FitContext) FindChargeParent(charge *ItemState) *ItemState {
	for _, m := range c.Modules {
		if m.Charge == charge {
			return m
		}
	}
	return nil
}

// TargetsForModifier: gli item su cui il modificatore agisce. effectID nil = assente.
func (c *FitContext) TargetsForModifier(modifier *types.SdeModifierInfo, source *ItemState, effectID *int) []*ItemState {
	root := c.ResolveDomain(modifier.Domain, source)
	switch modifier.Func {
	case "ItemModifier":
		// Il range di controllo droni e' una caratteristica della nave anche se la skill lo dichiara su charID.
		if modifier.Domain == "charID" && modifier.ModifiedAttributeID != nil && *modifier.ModifiedAttributeID == AttrDroneControlRange {
			if c.Ship != nil {
				return []*ItemState{c.Ship}
			}
			return nil
		}
		// Skill del personaggio che potenziano il danno missili: vanno alle cariche missile caricate.
		if modifier.Domain == "charID" && source.Kind == KindModule &&
			modifier.ModifiedAttributeID != nil && *modifier.ModifiedAttributeID == 212 {
			var out []*ItemState
			for _, m := range c.Modules {
				if m.Charge == nil {
					continue
				}
				some := false
				for _, sid := range missileBoostSkills {
					if ItemRequiresSkill(m.Charge, sid) {
						some = true
						break
					}
				}
				if !some {
					continue
				}
				out = append(out, m.Charge)
			}
			return out
		}
		if root != nil {
			return []*ItemState{root}
		}
		return nil

	case "LocationModifier":
		if root != nil {
			return []*ItemState{root}
		}
		return nil

	case "LocationGroupModifier":
		if root == nil || modifier.GroupID == nil {
			return nil
		}
		var out []*ItemState
		for _, it := range c.itemsInLocation(root) {
			if it.GroupID == *modifier.GroupID {
				out = append(out, it)
			}
		}
		return out

	case "LocationRequiredSkillModifier":
		if root == nil {
			return nil
		}
		if effectID != nil {
			if has, ok := pyfaGroupFilteredEffects(*effectID); ok {
				var out []*ItemState
				for _, it := range c.itemsInLocation(root) {
					if has(it.GroupID) {
						out = append(out, it)
					}
				}
				return out
			}
		}
		if modifier.SkillTypeID == nil {
			return nil
		}
		sid := *modifier.SkillTypeID
		var out []*ItemState
		for _, it := range c.itemsInLocation(root) {
			if ItemRequiresSkill(it, sid) {
				out = append(out, it)
			}
		}
		return out

	case "OwnerRequiredSkillModifier":
		if modifier.SkillTypeID == nil {
			return nil
		}
		skillID := *modifier.SkillTypeID
		var out []*ItemState
		for _, m := range c.Modules {
			if m.Charge != nil && ItemRequiresSkill(m.Charge, skillID) {
				out = append(out, m.Charge)
			}
		}
		for _, d := range c.Drones {
			if ItemRequiresSkill(d, skillID) {
				out = append(out, d)
			}
		}
		for _, f := range c.Fighters {
			if ItemRequiresSkill(f, skillID) {
				out = append(out, f)
			}
		}
		return out

	case "EffectStopper":
		return nil
	}
	return nil
}

// itemsInLocation: gli item "dentro" root (la nave contiene moduli/cariche/droni/fighter/subsystem/mode;
// il personaggio contiene impianti e booster).
func (c *FitContext) itemsInLocation(root *ItemState) []*ItemState {
	if root == c.Ship {
		out := []*ItemState{c.Ship}
		for _, m := range c.Modules {
			out = append(out, m)
			if m.Charge != nil {
				out = append(out, m.Charge)
			}
		}
		out = append(out, c.Drones...)
		out = append(out, c.Fighters...)
		out = append(out, c.Subsystems...)
		if c.Mode != nil {
			out = append(out, c.Mode)
		}
		return out
	}
	if root == c.Character {
		out := []*ItemState{c.Character}
		out = append(out, c.Implants...)
		out = append(out, c.Boosters...)
		return out
	}
	return []*ItemState{root}
}

// ItemRequiresSkill: l'item dichiara skillID fra i prerequisiti?
func ItemRequiresSkill(item *ItemState, skillID int) bool {
	for _, pair := range RequiredSkillPairs {
		if b, ok := item.GetBase(pair[0]); ok && b == float64(skillID) {
			return true
		}
	}
	return false
}

// ModuleAcceptsCharge: il groupID della carica deve combaciare con uno degli attributi di gruppo carica del modulo.
func ModuleAcceptsCharge(module, charge *ItemState) bool {
	for _, attrID := range ChargeGroupAttrs {
		allowed, _ := module.GetBase(attrID)
		if allowed != 0 && allowed == float64(charge.GroupID) {
			return true
		}
	}
	return false
}
