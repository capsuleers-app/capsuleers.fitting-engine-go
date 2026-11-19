// Porting letterale di src/effects/ewar.ts: classificazione EWAR e probabilita' di jam ECM.
package effects

import (
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
)

type EwarKind string

// ewarEffectKind e' EWAR_EFFECT_KIND (solo lookup per chiave).
func ewarEffectKind(effectID int) (EwarKind, bool) {
	switch effectID {
	case 1786:
		return "ECM", true
	case 1130:
		return "SENSOR_DAMP", true
	case 1799:
		return "TRACKING_DISRUPT", true
	case 14:
		return "WEB", true
	case 19:
		return "WARP_SCRAM", true
	case 52:
		return "WARP_DISRUPT", true
	case 28:
		return "NEUT", true
	case 1:
		return "NOS", true
	}
	return "", false
}

type EwarClass struct {
	Kind     EwarKind
	EffectID int
}

// ClassifyEwar: il primo effetto EWAR noto dell'item (nil = nessuno).
func ClassifyEwar(item *engine.ItemState) *EwarClass {
	for _, eid := range item.EffectIDs.Values() {
		if kind, ok := ewarEffectKind(eid); ok {
			return &EwarClass{Kind: kind, EffectID: eid}
		}
	}
	return nil
}

// EcmJamChance: probabilita' di jam del singolo ciclo (0..1).
func EcmJamChance(source, target *engine.ItemState) float64 {
	pairs := [4][2]int{
		{engine.AttrScanRadarStrength, engine.AttrScanRadarStrength},
		{engine.AttrScanLadarStrength, engine.AttrScanLadarStrength},
		{engine.AttrScanMagnetometricStrength, engine.AttrScanMagnetometricStrength},
		{engine.AttrScanGravimetricStrength, engine.AttrScanGravimetricStrength},
	}
	best := 0.0
	for _, p := range pairs {
		srcStrength := source.GetFinal(p[0], 0)
		if srcStrength <= 0 {
			continue
		}
		tgtStrength := target.GetFinal(p[1], 0)
		if tgtStrength <= 0 {
			continue
		}
		chance := srcStrength / tgtStrength
		if chance > best {
			best = chance
		}
	}
	return jsmath.Max(0, jsmath.Min(1, best))
}

// CombineJamChances: probabilita' combinata di piu' ECM indipendenti.
func CombineJamChances(chances []float64) float64 {
	probNotJammed := 1.0
	for _, p := range chances {
		if p <= 0 {
			continue
		}
		probNotJammed *= 1 - jsmath.Max(0, jsmath.Min(1, p))
	}
	return 1 - probNotJammed
}
