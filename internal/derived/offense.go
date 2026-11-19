// Porting letterale di src/derived/offense.ts: DPS, alpha e breakdown per arma, droni e fighter.
package derived

import (
	"fmt"
	"math"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/effects"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

type OffenseReport = types.Offense

// legacyReload1sEffect: LEGACY_RELOAD_1S_EFFECT_IDS in ordine (10, 34, 67, 101, 6995).
var legacyReload1sEffectIDs = [5]int{10, 34, 67, 101, 6995}

// minPositive e' Math.min(...values.filter(v => v > 0)): +Infinity se l'elenco e' vuoto.
func minPositive(vals []float64) float64 {
	m := math.Inf(1)
	for _, v := range vals {
		if v > 0 {
			m = jsmath.Min(m, v)
		}
	}
	return m
}

func maxPositive(vals []float64) float64 {
	m := math.Inf(-1)
	for _, v := range vals {
		if v > 0 {
			m = jsmath.Max(m, v)
		}
	}
	return m
}

func finiteOr0(v float64) float64 {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return 0
	}
	return v
}

// ComputeOffense: DPS e alpha di armi, droni e fighter con il breakdown per gruppo di armi.
func ComputeOffense(ctx *engine.FitContext, dataset *types.FittingDataset, fit *types.Fit) OffenseReport {
	var breakdown []types.WeaponContribution
	weaponDps, weaponSustainedDps, droneDps, fighterDps, alphaStrike := 0.0, 0.0, 0.0, 0.0, 0.0

	// Armi di moduli: raggruppate per (type, carica, stato, ciclo) in ordine di inserimento.
	weaponGroups := omap.New[string, *types.WeaponContribution]()
	for _, m := range ctx.Modules {
		if m.State != types.StateActive && m.State != types.StateOverload {
			continue
		}
		contribution := weaponContributionFor(m, ctx, dataset)
		if contribution == nil {
			continue
		}
		weaponDps += contribution.Dps
		if contribution.SustainedDps != nil {
			weaponSustainedDps += *contribution.SustainedDps
		} else {
			weaponSustainedDps += contribution.Dps
		}
		alphaStrike += contribution.Alpha

		chargeKey := "-"
		if contribution.ChargeTypeID != nil {
			chargeKey = fmt.Sprintf("%d", *contribution.ChargeTypeID)
		}
		key := fmt.Sprintf("%d|%s|%s|%s", contribution.TypeID, chargeKey, m.State, jsmath.ToFixed(contribution.CycleSeconds, 3))
		if existing, ok := weaponGroups.Get(key); ok {
			existing.Dps += contribution.Dps
			existing.Alpha += contribution.Alpha
			if existing.SustainedDps != nil && contribution.SustainedDps != nil {
				v := *existing.SustainedDps + *contribution.SustainedDps
				existing.SustainedDps = &v
			}
			if existing.ChainDpsMax != nil && contribution.ChainDpsMax != nil {
				v := *existing.ChainDpsMax + *contribution.ChainDpsMax
				existing.ChainDpsMax = &v
			}
			existing.Damages = types.DamageSplit{
				Em:        existing.Damages.Em + contribution.Damages.Em,
				Thermal:   existing.Damages.Thermal + contribution.Damages.Thermal,
				Kinetic:   existing.Damages.Kinetic + contribution.Damages.Kinetic,
				Explosive: existing.Damages.Explosive + contribution.Damages.Explosive,
				Total:     existing.Damages.Total + contribution.Damages.Total,
			}
			existing.Count += 1
		} else {
			c := *contribution
			c.Count = 1
			weaponGroups.Set(key, &c)
		}
	}
	for i := 0; i < weaponGroups.Len(); i++ {
		_, row := weaponGroups.At(i)
		breakdown = append(breakdown, *row)
	}

	// Droni.
	fitDronesByItemID := omap.New[string, *types.FitDrone]()
	for i := range fit.Drones {
		fitDronesByItemID.Set(fit.Drones[i].ID, &fit.Drones[i])
	}
	for _, droneItem := range ctx.Drones {
		fd, ok := fitDronesByItemID.Get(droneItem.ID)
		if !ok || fd.CountActive <= 0 {
			continue
		}
		contribution := droneContributionFor(droneItem, fd.CountActive, ctx, dataset)
		if contribution == nil {
			continue
		}
		breakdown = append(breakdown, *contribution)
		droneDps += contribution.Dps
		alphaStrike += contribution.Alpha
	}

	// Fighter.
	fitFightersByItemID := omap.New[string, *types.FitFighter]()
	for i := range fit.Fighters {
		fitFightersByItemID.Set(fit.Fighters[i].ID, &fit.Fighters[i])
	}
	for _, fighterItem := range ctx.Fighters {
		ff, ok := fitFightersByItemID.Get(fighterItem.ID)
		if !ok || ff.Count <= 0 {
			continue
		}
		contribution := fighterContributionFor(fighterItem, ff.Count, ctx, dataset)
		if contribution == nil {
			continue
		}
		breakdown = append(breakdown, *contribution)
		fighterDps += contribution.Dps
		alphaStrike += contribution.Alpha
	}

	var turretContribs, missileContribs []types.WeaponContribution
	for _, b := range breakdown {
		if b.Kind == "TURRET" {
			turretContribs = append(turretContribs, b)
		}
		if b.Kind == "MISSILE" {
			missileContribs = append(missileContribs, b)
		}
	}
	col := func(cs []types.WeaponContribution, f func(*types.WeaponContribution) float64) []float64 {
		out := make([]float64, len(cs))
		for i := range cs {
			out[i] = f(&cs[i])
		}
		return out
	}
	weaponOptimal, weaponFalloff := 0.0, 0.0
	var weaponTracking, explosionVelocity, explosionRadius, missileRange *float64
	if len(turretContribs) > 0 {
		weaponOptimal = minPositive(col(turretContribs, func(c *types.WeaponContribution) float64 { return c.Range.Optimal }))
		weaponFalloff = minPositive(col(turretContribs, func(c *types.WeaponContribution) float64 { return c.Range.Falloff }))
		wt := minPositive(col(turretContribs, func(c *types.WeaponContribution) float64 { return c.Range.Tracking }))
		weaponTracking = &wt // senza isFinite: puo' essere +Infinity, come nel TS
	}
	if len(missileContribs) > 0 {
		ev, er := missileContribs[0].Range.ExplosionVelocity, missileContribs[0].Range.ExplosionRadius
		explosionVelocity, explosionRadius = &ev, &er
		mr := maxPositive(col(missileContribs, func(c *types.WeaponContribution) float64 { return c.Range.FlightRange }))
		if !math.IsInf(mr, 0) && !math.IsNaN(mr) {
			missileRange = &mr
		}
	}

	return OffenseReport{
		WeaponDps: weaponDps, WeaponSustainedDps: weaponSustainedDps, DroneDps: droneDps, FighterDps: fighterDps,
		TotalDps: weaponDps + droneDps + fighterDps, TotalSustainedDps: weaponSustainedDps + droneDps + fighterDps,
		AlphaStrike:   alphaStrike,
		WeaponOptimal: finiteOr0(weaponOptimal), WeaponFalloff: finiteOr0(weaponFalloff),
		WeaponTracking: weaponTracking, ExplosionVelocity: explosionVelocity, ExplosionRadius: explosionRadius,
		MissileRange: missileRange, Breakdown: breakdown,
	}
}

func fighterContributionFor(fighter *engine.ItemState, count float64, _ *engine.FitContext, _ *types.FittingDataset) *types.WeaponContribution {
	totalAlpha, totalDps, primaryCycle := 0.0, 0.0, 0.0
	damages := types.DamageSplit{}

	// Attack missile (effetto 6465) e attacco secondario (6431): stesso schema, attributi diversi.
	ability := func(effectID, aEm, aThm, aKin, aExp, aMult int, setPrimary func(durMs float64)) {
		if !fighter.EffectIDs.Has(effectID) {
			return
		}
		em := fighter.GetFinal(aEm, 0)
		thm := fighter.GetFinal(aThm, 0)
		kin := fighter.GetFinal(aKin, 0)
		exp := fighter.GetFinal(aExp, 0)
		mult := fighter.GetFinal(aMult, 1)
		durMs := fighter.GetFinal(2233, 0)
		if durMs > 0 {
			sumDmg := em + thm + kin + exp
			alpha := float64(sumDmg*mult) * count
			dps := alpha / (durMs / 1000)
			totalAlpha += alpha
			totalDps += dps
			damages.Em += float64((em * mult) * count)
			damages.Thermal += float64((thm * mult) * count)
			damages.Kinetic += float64((kin * mult) * count)
			damages.Explosive += float64((exp * mult) * count)
			damages.Total += float64((sumDmg * mult) * count)
			setPrimary(durMs)
		}
	}
	ability(6465, 2227, 2228, 2229, 2230, 2226, func(durMs float64) { primaryCycle = durMs / 1000 })
	ability(6431, 2131, 2132, 2133, 2134, 2130, func(durMs float64) {
		if primaryCycle == 0 {
			primaryCycle = durMs / 1000
		}
	})

	if totalDps == 0 {
		return nil
	}
	return &types.WeaponContribution{
		SourceID: fighter.ID, TypeID: fighter.TypeID, Kind: "DRONE", Alpha: totalAlpha, Dps: totalDps,
		CycleSeconds: primaryCycle, Damages: damages,
		Range: types.WeaponRange{
			Optimal: fighter.GetFinal(2236, 0), Falloff: fighter.GetFinal(2237, 0), Tracking: 0, BurstRange: 0,
			ExplosionRadius: fighter.GetFinal(2125, 0), ExplosionVelocity: fighter.GetFinal(2126, 0),
			Drf: fighter.GetFinal(2127, 0), FlightRange: 0,
		},
		Count: count,
	}
}

const (
	breacherPodEffectID        = 12174
	breacherPodDefaultTargetHP = 100000
)

// readReloadInfo: tempo di reload e cariche per caricamento del modulo.
func readReloadInfo(module *engine.ItemState) (reloadSeconds, chargesPerLoad float64) {
	reloadMs := module.GetFinal(attrReloadTime, 0)
	if reloadMs <= 0 {
		for _, eid := range legacyReload1sEffectIDs {
			if module.EffectIDs.Has(eid) {
				reloadMs = 1000
				break
			}
		}
	}
	reloadSeconds = reloadMs / 1000
	charge := module.Charge
	if charge == nil {
		return 0, 1
	}
	capacity := module.GetFinal(attrCapacity, 0)
	chargeVolume := charge.GetFinal(attrVolume, 0)
	if capacity <= 0 || chargeVolume <= 0 {
		return reloadSeconds, 1
	}
	chargesPerLoad = math.Floor(capacity / chargeVolume)
	return reloadSeconds, jsmath.Max(1, chargesPerLoad)
}

func computeSustainedDps(alpha, cycleSeconds, reloadSeconds, chargesPerLoad float64) float64 {
	if cycleSeconds <= 0 {
		return 0
	}
	if reloadSeconds <= 0 || chargesPerLoad <= 0 {
		return alpha / cycleSeconds
	}
	damagePerLoad := chargesPerLoad * alpha
	timePerLoad := float64(chargesPerLoad*cycleSeconds) + reloadSeconds
	return damagePerLoad / timePerLoad
}

func chargeTypeIDOf(m *engine.ItemState) *int {
	if m.Charge == nil {
		return nil
	}
	id := m.Charge.TypeID
	return &id
}

func weaponContributionFor(module *engine.ItemState, ctx *engine.FitContext, dataset *types.FittingDataset) *types.WeaponContribution {
	cls := effects.ClassifyWeapon(module)
	if cls == nil {
		return nil
	}
	effect, ok := dataset.Effects.Get(cls.EffectID)
	if !ok {
		return nil
	}

	// Breacher Pod: danno a tick (DoT), non il modello alpha/ciclo.
	if module.Charge != nil && module.Charge.EffectIDs.Has(breacherPodEffectID) {
		cycle := effects.ReadCycleInfo(module, cls.Kind)
		return breacherContributionFor(module, cycle.CycleSeconds, effect, cls)
	}

	damages := effects.ReadDamageComponents(cls.DamageSource)
	split := func(d effects.WeaponDamageComponents) types.DamageSplit {
		return types.DamageSplit{Em: d.Em, Thermal: d.Thermal, Kinetic: d.Kinetic, Explosive: d.Explosive, Total: d.Total}
	}
	rangeOf := func() types.WeaponRange {
		r := effects.ReadRangeInfo(module, effect, cls.Kind)
		return types.WeaponRange{Optimal: r.Optimal, Falloff: r.Falloff, Tracking: r.Tracking, BurstRange: r.BurstRange,
			ExplosionRadius: r.ExplosionRadius, ExplosionVelocity: r.ExplosionVelocity, Drf: r.Drf, FlightRange: r.FlightRange}
	}
	if damages.Total <= 0 {
		return &types.WeaponContribution{
			SourceID: module.ID, TypeID: module.TypeID, Kind: types.WeaponKind(cls.Kind), Alpha: 0, Dps: 0, CycleSeconds: 0,
			Damages: split(damages), Range: rangeOf(), ChargeTypeID: chargeTypeIDOf(module), Count: 1,
		}
	}
	cycle := effects.ReadCycleInfo(module, cls.Kind)
	if cycle.CycleSeconds <= 0 {
		return &types.WeaponContribution{
			SourceID: module.ID, TypeID: module.TypeID, Kind: types.WeaponKind(cls.Kind),
			Alpha: damages.Total * cycle.DamageMultiplier, Dps: 0, CycleSeconds: 0,
			Damages: split(damages), Range: rangeOf(), ChargeTypeID: chargeTypeIDOf(module), Count: 1,
		}
	}

	alpha := damages.Total * cycle.DamageMultiplier
	dps := alpha / cycle.CycleSeconds
	reloadSeconds, chargesPerLoad := readReloadInfo(module)
	sustainedDps := computeSustainedDps(alpha, cycle.CycleSeconds, reloadSeconds, chargesPerLoad)

	// Vorton Projector: catena di bersagli (serie geometrica).
	vortonChainSpecs := [2][3]int{{8037, 3037, 1353}, {6447, 2104, 2106}}
	var chainDpsMax, chainTargetCount *float64
	for _, spec := range vortonChainSpecs {
		if !module.EffectIDs.Has(spec[0]) {
			continue
		}
		targets := jsmath.Max(1, jsmath.Round(module.GetFinal(spec[1], 1)))
		reductionFactor := module.GetFinal(spec[2], 0)
		var v float64
		if reductionFactor > 0 && reductionFactor < 1 {
			r := 1 - reductionFactor
			multiplier := (1 - math.Pow(r, targets)) / (1 - r)
			v = dps * multiplier
		} else {
			v = dps * targets
		}
		chainDpsMax = &v
		t := targets
		chainTargetCount = &t
		break
	}

	// Disintegratore Triglavian: info di spool.
	var disintegrator *types.DisintegratorInfo
	if module.EffectIDs.Has(6995) {
		maxBonus, bonusPerCycle := 0.0, 0.0
		if module.HasAttr(2734) {
			maxBonus = module.GetFinal(2734, 0)
		}
		if module.HasAttr(2733) {
			bonusPerCycle = module.GetFinal(2733, 0)
		}
		if maxBonus > 0 {
			spoolFactor := 1 + float64(ctx.DisintegratorSpoolPercent*maxBonus)
			baseDps := dps
			if spoolFactor > 0 {
				baseDps = dps / spoolFactor
			}
			disintegrator = &types.DisintegratorInfo{MaxBonus: maxBonus, BonusPerCycle: bonusPerCycle, BaseDps: baseDps}
		}
	}

	var reloadSecondsPtr *float64
	if reloadSeconds != 0 && !math.IsNaN(reloadSeconds) {
		reloadSecondsPtr = &reloadSeconds
	}
	return &types.WeaponContribution{
		SourceID: module.ID, TypeID: module.TypeID, Kind: types.WeaponKind(cls.Kind), Alpha: alpha, Dps: dps,
		SustainedDps: &sustainedDps, ReloadSeconds: reloadSecondsPtr, ChargesPerLoad: &chargesPerLoad,
		ChainDpsMax: chainDpsMax, ChainTargetCount: chainTargetCount, CycleSeconds: cycle.CycleSeconds,
		Damages: scaledDamage(damages, cycle.DamageMultiplier), Range: rangeOf(), ChargeTypeID: chargeTypeIDOf(module),
		Count: 1, Disintegrator: disintegrator,
	}
}

func breacherContributionFor(module *engine.ItemState, cycleSeconds float64, effect *types.SdeEffect, cls *effects.WeaponClassification) *types.WeaponContribution {
	charge := module.Charge
	maxPerTick := charge.GetFinal(engine.AttrDotMaxDamagePerTick, 0)
	pctPerTickRaw := charge.GetFinal(engine.AttrDotMaxHpPercentagePerTick, 0)
	dotDurationMs := charge.GetFinal(engine.AttrDotDuration, 0)
	pctDamage := (pctPerTickRaw / 100) * breacherPodDefaultTargetHP
	tickDamage := jsmath.Min(maxPerTick, pctDamage)
	damages := types.DamageSplit{Em: 0, Thermal: tickDamage, Kinetic: 0, Explosive: 0, Total: tickDamage}
	fallback := 1.0
	if dotDurationMs > 0 {
		fallback = dotDurationMs / 1000
	}
	r := effects.ReadRangeInfo(module, effect, cls.Kind)
	cid := charge.TypeID
	return &types.WeaponContribution{
		SourceID: module.ID, TypeID: module.TypeID, Kind: types.WeaponKind(cls.Kind), Alpha: tickDamage, Dps: tickDamage,
		CycleSeconds: orElse(cycleSeconds, fallback), Damages: damages,
		Range: types.WeaponRange{Optimal: r.Optimal, Falloff: r.Falloff, Tracking: r.Tracking, BurstRange: r.BurstRange,
			ExplosionRadius: r.ExplosionRadius, ExplosionVelocity: r.ExplosionVelocity, Drf: r.Drf, FlightRange: r.FlightRange},
		ChargeTypeID: &cid, Count: 1,
	}
}

func droneContributionFor(drone *engine.ItemState, countActive float64, _ *engine.FitContext, dataset *types.FittingDataset) *types.WeaponContribution {
	cls := effects.ClassifyWeapon(drone)
	if cls == nil {
		return nil
	}
	effect, ok := dataset.Effects.Get(cls.EffectID)
	if !ok {
		return nil
	}
	damages := readDamageComponentsFromDrone(drone)
	if damages.Total <= 0 {
		return nil
	}
	cycleMs := orElse(drone.GetFinal(engine.AttrRateOfFire, 0), drone.GetFinal(engine.AttrDamageDuration, 0))
	if cycleMs <= 0 {
		return nil
	}
	cycleSeconds := cycleMs / 1000
	damageMultiplier := drone.GetFinal(engine.AttrDamageMultiplier, 1)
	alphaSingle := damages.Total * damageMultiplier
	dpsSingle := alphaSingle / cycleSeconds
	r := effects.ReadRangeInfo(drone, effect, cls.Kind)
	return &types.WeaponContribution{
		SourceID: drone.ID, TypeID: drone.TypeID, Kind: "DRONE", Alpha: alphaSingle * countActive, Dps: dpsSingle * countActive,
		CycleSeconds: cycleSeconds, Damages: scaledDamage(damages, damageMultiplier*countActive),
		Range: types.WeaponRange{Optimal: r.Optimal, Falloff: r.Falloff, Tracking: r.Tracking, BurstRange: r.BurstRange,
			ExplosionRadius: r.ExplosionRadius, ExplosionVelocity: r.ExplosionVelocity, Drf: r.Drf, FlightRange: r.FlightRange},
		Count: countActive,
	}
}

func readDamageComponentsFromDrone(drone *engine.ItemState) effects.WeaponDamageComponents {
	em := drone.GetFinal(engine.AttrEmDamage, 0)
	thermal := drone.GetFinal(engine.AttrThermalDamage, 0)
	kinetic := drone.GetFinal(engine.AttrKineticDamage, 0)
	explosive := drone.GetFinal(engine.AttrExplosiveDamage, 0)
	return effects.WeaponDamageComponents{Em: em, Thermal: thermal, Kinetic: kinetic, Explosive: explosive, Total: em + thermal + kinetic + explosive}
}

func scaledDamage(d effects.WeaponDamageComponents, factor float64) types.DamageSplit {
	return types.DamageSplit{Em: d.Em * factor, Thermal: d.Thermal * factor, Kinetic: d.Kinetic * factor, Explosive: d.Explosive * factor, Total: d.Total * factor}
}
