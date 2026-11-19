// Porting letterale di src/derived/ehp.ts: EHP per layer e contro un profilo di danno.
package derived

import (
	"math"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

type DefenseLayerKind string

type layerAttributeMap struct{ hp, em, thermal, kinetic, explosive int }

func layerAttrs(layer DefenseLayerKind) layerAttributeMap {
	switch layer {
	case "SHIELD":
		return layerAttributeMap{engine.AttrShieldCapacity, engine.AttrShieldEmRes, engine.AttrShieldThermalRes, engine.AttrShieldKineticRes, engine.AttrShieldExplosiveRes}
	case "ARMOR":
		return layerAttributeMap{engine.AttrArmorHp, engine.AttrArmorEmRes, engine.AttrArmorThermalRes, engine.AttrArmorKineticRes, engine.AttrArmorExplosiveRes}
	}
	return layerAttributeMap{engine.AttrHp, engine.AttrStructureEmRes, engine.AttrStructureThermalRes, engine.AttrStructureKineticRes, engine.AttrStructureExplosiveRes}
}

var uniformProfile = types.DamageProfile{Name: "Uniform", Em: 0.25, Thermal: 0.25, Kinetic: 0.25, Explosive: 0.25}

// ComputeLayerEhp: HP, EHP uniforme e contro il profilo, resistenze di un layer. profile nil = uniforme.
func ComputeLayerEhp(ship *engine.ItemState, layer DefenseLayerKind, profile *types.DamageProfile) types.LayerEhp {
	attrs := layerAttrs(layer)
	hp := ship.GetFinal(attrs.hp, 0)
	resEm := ship.GetFinal(attrs.em, 1)
	resTh := ship.GetFinal(attrs.thermal, 1)
	resKi := ship.GetFinal(attrs.kinetic, 1)
	resEx := ship.GetFinal(attrs.explosive, 1)
	p := &uniformProfile
	if profile != nil {
		p = profile
	}
	return types.LayerEhp{
		Hp:                hp,
		EhpUniform:        EhpUnderProfile(hp, resEm, resTh, resKi, resEx, &uniformProfile),
		EhpAgainstProfile: EhpUnderProfile(hp, resEm, resTh, resKi, resEx, p),
		Resistances:       types.Resistances{Em: 1 - resEm, Thermal: 1 - resTh, Kinetic: 1 - resKi, Explosive: 1 - resEx},
	}
}

// EhpUnderProfile: HP / risonanza efficace pesata sul profilo.
func EhpUnderProfile(hp, resonanceEm, resonanceThermal, resonanceKinetic, resonanceExplosive float64, profile *types.DamageProfile) float64 {
	if hp <= 0 {
		return 0
	}
	totalWeight := profile.Em + profile.Thermal + profile.Kinetic + profile.Explosive
	if totalWeight <= 0 {
		return hp // profilo degenere: HP grezzi
	}
	effectiveResonance := (float64(profile.Em*resonanceEm) +
		float64(profile.Thermal*resonanceThermal) +
		float64(profile.Kinetic*resonanceKinetic) +
		float64(profile.Explosive*resonanceExplosive)) / totalWeight
	if effectiveResonance <= 0 {
		return math.Inf(1) // immune
	}
	return hp / effectiveResonance
}

// ComputeTotalEhp: somma dei tre layer (contro il profilo o uniforme); non finito -> MAX_SAFE_INTEGER.
func ComputeTotalEhp(shield, armor, hull types.LayerEhp, useProfile bool) float64 {
	pick := func(l types.LayerEhp) float64 {
		if useProfile {
			return l.EhpAgainstProfile
		}
		return l.EhpUniform
	}
	total := pick(shield) + pick(armor) + pick(hull)
	if math.IsInf(total, 0) || math.IsNaN(total) {
		return 9007199254740991
	}
	return total
}
