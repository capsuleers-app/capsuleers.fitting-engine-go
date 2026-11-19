// Porting letterale di src/effects/weapon.ts: classificazione delle armi e lettura di danno, ciclo e range.
package effects

import (
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

// WeaponClassification e' il risultato di ClassifyWeapon. DamageSource nil = nessuna sorgente di danno.
type WeaponClassification struct {
	Kind         engine.WeaponEffectKind
	EffectID     int
	DamageSource *engine.ItemState
}

// ClassifyWeapon: il primo effetto dell'item (in ordine) che compare in WEAPON_EFFECT_KIND vince.
func ClassifyWeapon(item *engine.ItemState) *WeaponClassification {
	for _, eid := range item.EffectIDs.Values() {
		kind, ok := engine.WeaponKindOf(eid)
		if !ok {
			continue
		}
		damageSource := item // smart bomb / droni portano il danno su se' stessi
		if kind == engine.WeaponTurret || kind == engine.WeaponMissile {
			damageSource = item.Charge
		}
		return &WeaponClassification{Kind: kind, EffectID: eid, DamageSource: damageSource}
	}
	return nil
}

type WeaponDamageComponents struct {
	Em, Thermal, Kinetic, Explosive, Total float64
}

// ReadDamageComponents legge i quattro tipi di danno da una sorgente (nil = tutto 0).
func ReadDamageComponents(source *engine.ItemState) WeaponDamageComponents {
	if source == nil {
		return WeaponDamageComponents{}
	}
	em := source.GetFinal(engine.AttrEmDamage, 0)
	thermal := source.GetFinal(engine.AttrThermalDamage, 0)
	kinetic := source.GetFinal(engine.AttrKineticDamage, 0)
	explosive := source.GetFinal(engine.AttrExplosiveDamage, 0)
	return WeaponDamageComponents{Em: em, Thermal: thermal, Kinetic: kinetic, Explosive: explosive, Total: em + thermal + kinetic + explosive}
}

type WeaponCycleInfo struct {
	CycleSeconds     float64
	DamageMultiplier float64
}

// ReadCycleInfo: tempo di ciclo e moltiplicatore di danno.
func ReadCycleInfo(item *engine.ItemState, kind engine.WeaponEffectKind) WeaponCycleInfo {
	cycleAttr := engine.AttrRateOfFire
	if kind == engine.WeaponSmartbomb {
		cycleAttr = engine.AttrDamageDuration
	}
	cycleMs := item.GetFinal(cycleAttr, 0)
	cycleSeconds := 0.0
	if cycleMs > 0 {
		cycleSeconds = cycleMs / 1000
	}
	damageMultiplier := 1.0
	if kind == engine.WeaponTurret {
		damageMultiplier = item.GetFinal(engine.AttrDamageMultiplier, 1)
	} else if kind == engine.WeaponMissile {
		launcherMul := item.GetFinal(engine.AttrDamageMultiplier, 1)
		chargeMul := 1.0
		if item.Charge != nil {
			chargeMul = item.Charge.GetFinal(engine.AttrMissileDamageMultiplier, 1)
		}
		damageMultiplier = launcherMul * chargeMul
	}
	return WeaponCycleInfo{CycleSeconds: cycleSeconds, DamageMultiplier: damageMultiplier}
}

type WeaponRangeInfo struct {
	Optimal, Falloff, Tracking, BurstRange               float64
	ExplosionRadius, ExplosionVelocity, Drf, FlightRange float64
}

func attrSet(p *int) bool { return p != nil && *p != 0 }

// ReadRangeInfo: range dell'arma dagli attributi dell'effetto (e della carica per i missili).
func ReadRangeInfo(item *engine.ItemState, effect *types.SdeEffect, kind engine.WeaponEffectKind) WeaponRangeInfo {
	optimal, falloff, tracking := 0.0, 0.0, 0.0
	if attrSet(effect.RangeAttributeID) {
		optimal = item.GetFinal(*effect.RangeAttributeID, 0)
	}
	if attrSet(effect.FalloffAttributeID) {
		falloff = item.GetFinal(*effect.FalloffAttributeID, 0)
	}
	if attrSet(effect.TrackingSpeedAttributeID) {
		tracking = item.GetFinal(*effect.TrackingSpeedAttributeID, 0)
	}
	burstRange := 0.0
	if kind == engine.WeaponSmartbomb && attrSet(effect.RangeAttributeID) {
		burstRange = optimal
	}
	explosionRadius, explosionVelocity, drf, flightRange := 0.0, 0.0, 0.0, 0.0
	if kind == engine.WeaponMissile && item.Charge != nil {
		explosionRadius = item.Charge.GetFinal(engine.AttrExplosionRadius, 0)
		explosionVelocity = item.Charge.GetFinal(engine.AttrExplosionVelocity, 0)
		drf = item.Charge.GetFinal(engine.AttrDrf, 0)
		velocity := item.Charge.GetFinal(engine.AttrMaxVelocity, 0)
		flightMs := item.Charge.GetFinal(engine.AttrExplosionDelay, 0)
		flightRange = velocity * flightMs / 1000
	}
	return WeaponRangeInfo{Optimal: optimal, Falloff: falloff, Tracking: tracking, BurstRange: burstRange,
		ExplosionRadius: explosionRadius, ExplosionVelocity: explosionVelocity, Drf: drf, FlightRange: flightRange}
}
