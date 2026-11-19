// Porting letterale di src/derived/tank.ts: velocita' di riparazione (shield/armor/hull) e regen passivo.
package derived

import (
	"math"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

const (
	nanitRepairPasteTypeID      = 28668
	ancillaryRepPasteMultiplier = 3
)

// pasteFueledRepEffects: PASTE_FUELED_REP_EFFECTS (solo Has).
func pasteFueledRepEffect(eid int) bool { return eid == 5275 || eid == 6651 }

type layerAccumulator struct {
	amountTotal, perSecondTotal, perSecondSustainedTotal float64
	cycleSecondsList                                     []float64
}

func averageCycle(a *layerAccumulator) float64 {
	if len(a.cycleSecondsList) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range a.cycleSecondsList {
		sum += v
	}
	return sum / float64(len(a.cycleSecondsList))
}

// ComputeTank: riparazioni attive per layer e picco di regen passivo dello shield.
func ComputeTank(ctx *engine.FitContext) types.TankRates {
	acc := map[string]*layerAccumulator{"SHIELD": {}, "ARMOR": {}, "HULL": {}}
	for _, m := range ctx.Modules {
		if m.State != types.StateActive && m.State != types.StateOverload {
			continue
		}
		for _, eid := range m.EffectIDs.Values() {
			repInfo, ok := engine.RepairEffectAmountAttr(eid)
			if !ok {
				continue
			}
			baseAmount := m.GetFinal(repInfo.AmountAttr, 0)
			if baseAmount <= 0 {
				continue
			}
			amount := baseAmount
			fueled := pasteFueledRepEffect(eid) && m.Charge != nil && m.Charge.TypeID == nanitRepairPasteTypeID
			if fueled {
				amount *= ancillaryRepPasteMultiplier
			}
			durationMs := m.GetFinal(73 /* duration attr id */, 0)
			if durationMs <= 0 {
				continue
			}
			cycleSeconds := durationMs / 1000
			a := acc[repInfo.Layer]
			a.amountTotal += amount
			a.perSecondTotal += amount / cycleSeconds
			a.cycleSecondsList = append(a.cycleSecondsList, cycleSeconds)
			if fueled && m.Charge != nil {
				capacity := m.GetFinal(attrCapacity, 0)
				chargeVol := m.Charge.GetFinal(attrVolume, 1)
				charges := 0.0
				if chargeVol > 0 {
					charges = math.Floor(capacity / chargeVol)
				}
				reloadMs := m.GetFinal(attrReloadTime, 60000) // AAR: default 60 s
				reloadSec := reloadMs / 1000
				if charges > 0 {
					damagePerLoad := amount * charges
					timePerLoad := float64(charges*cycleSeconds) + reloadSec
					a.perSecondSustainedTotal += damagePerLoad / timePerLoad
				} else {
					a.perSecondSustainedTotal += amount / cycleSeconds
				}
			} else {
				a.perSecondSustainedTotal += amount / cycleSeconds
			}
		}
	}

	ship := ctx.Ship
	shieldCap := ship.GetFinal(engine.AttrShieldCapacity, 0)
	shieldRechargeMs := ship.GetFinal(engine.AttrShieldRechargeRate, 0)
	passiveShieldRegenPeak := PeakRecharge(shieldCap, shieldRechargeMs)

	s, a, h := acc["SHIELD"], acc["ARMOR"], acc["HULL"]
	return types.TankRates{
		ShieldRepairAmount: s.amountTotal, ShieldRepairDuration: averageCycle(s),
		ShieldRepairPerSecond: s.perSecondTotal, ShieldRepairPerSecondSustained: s.perSecondSustainedTotal,
		ArmorRepairAmount: a.amountTotal, ArmorRepairDuration: averageCycle(a),
		ArmorRepairPerSecond: a.perSecondTotal, ArmorRepairPerSecondSustained: a.perSecondSustainedTotal,
		HullRepairAmount: h.amountTotal, HullRepairDuration: averageCycle(h),
		HullRepairPerSecond: h.perSecondTotal, HullRepairPerSecondSustained: h.perSecondSustainedTotal,
		PassiveShieldRegenPeak: passiveShieldRegenPeak,
	}
}
