package main

// Shared stat schema + diff (port of test/diff/stat-schema.mjs of eve-fit-engine). Both engines emit the
// SAME nested stat object (same keys + units): ours via oursToSchema, pyfa (oracle/pyfa_oracle.py)
// already in this shape. Units: resists in %, targeting/drone range in km, cap stable in %, everything
// else in raw EVE units. Comparison: relative tolerance (default 1%) plus an absolute epsilon.

import (
	"math"
	"sort"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

// flat is a flattened stat block ("group.path" -> float64 | bool | nil) that keeps its key order.
type flat struct {
	keys []string
	vals map[string]any
}

func newFlat() *flat { return &flat{vals: map[string]any{}} }

func (f *flat) set(k string, v any) {
	if _, ok := f.vals[k]; !ok {
		f.keys = append(f.keys, k)
	}
	f.vals[k] = v
}

func (f *flat) filter(keep func(string) bool) *flat {
	out := newFlat()
	for _, k := range f.keys {
		if keep(k) {
			out.set(k, f.vals[k])
		}
	}
	return out
}

func pct(resonanceResist float64) float64 { return resonanceResist * 100 }

// oursToSchema maps our engine's derived block onto the common schema (matches pyfa_oracle).
func oursToSchema(d *fitengine.DerivedStats) *flat {
	f := newFlat()
	fit, def, off, cp, nav, tgt := d.Fitting, d.Defense, d.Offense, d.Capacitor, d.Navigation, d.Targeting
	f.set("fitting.cpuUsed", fit.CpuUsed)
	f.set("fitting.cpuMax", fit.CpuMax)
	f.set("fitting.powerUsed", fit.PowerUsed)
	f.set("fitting.powerMax", fit.PowerMax)
	f.set("fitting.calibrationUsed", fit.CalibrationUsed)
	f.set("fitting.calibrationMax", fit.CalibrationMax)
	f.set("defense.shieldHp", def.Shield.Hp)
	f.set("defense.armorHp", def.Armor.Hp)
	f.set("defense.hullHp", def.Hull.Hp)
	for _, l := range []struct {
		name  string
		layer types.LayerEhp
	}{{"shieldResist", def.Shield}, {"armorResist", def.Armor}, {"hullResist", def.Hull}} {
		r := l.layer.Resistances
		f.set("defense."+l.name+".em", pct(r.Em))
		f.set("defense."+l.name+".thermal", pct(r.Thermal))
		f.set("defense."+l.name+".kinetic", pct(r.Kinetic))
		f.set("defense."+l.name+".explosive", pct(r.Explosive))
	}
	// The oracle uses a uniform 25/25/25/25 damage pattern; ehpAgainstProfile is the comparable value.
	f.set("defense.ehpShield", def.Shield.EhpAgainstProfile)
	f.set("defense.ehpArmor", def.Armor.EhpAgainstProfile)
	f.set("defense.ehpHull", def.Hull.EhpAgainstProfile)
	f.set("defense.ehpTotal", def.EhpTotalAgainstProfile)
	f.set("offense.weaponDps", off.WeaponDps)
	f.set("offense.droneDps", off.DroneDps)
	f.set("offense.totalDps", off.TotalDps)
	// pyfa's getWeaponVolley() counts every weapon EXCEPT drones/fighters: subtract only those.
	alpha := 0.0
	for _, b := range off.Breakdown {
		if b.Kind != "DRONE" && b.Kind != "FIGHTER" {
			alpha += b.Alpha
		}
	}
	f.set("offense.alphaStrike", alpha)
	f.set("capacitor.capacity", cp.Capacity)
	f.set("capacitor.stable", cp.Stable)
	if cp.Stable {
		f.set("capacitor.stablePercent", cp.StablePercent*100)
		f.set("capacitor.secondsToEmpty", nil)
	} else {
		f.set("capacitor.stablePercent", nil)
		if cp.SecondsToEmpty != nil {
			f.set("capacitor.secondsToEmpty", *cp.SecondsToEmpty)
		} else {
			f.set("capacitor.secondsToEmpty", nil)
		}
	}
	f.set("navigation.maxVelocity", nav.MaxVelocity)
	f.set("navigation.alignTime", nav.AlignTimeSeconds)
	f.set("navigation.warpSpeed", nav.WarpSpeed)
	f.set("navigation.mass", nav.Mass)
	f.set("navigation.agility", nav.Agility)
	f.set("navigation.signatureRadius", tgt.SignatureRadius)
	f.set("targeting.maxTargetingRange", tgt.MaxTargetingRange/1000)
	f.set("targeting.scanResolution", tgt.ScanResolution)
	f.set("targeting.sensorStrength", tgt.SensorStrength)
	f.set("targeting.maxLockedTargets", tgt.MaxLockedTargets)
	// `x / 1000 || null`: 0 (and NaN) become null.
	if dcr := d.Drones.ControlRange / 1000; dcr != 0 && !math.IsNaN(dcr) {
		f.set("targeting.droneControlRange", dcr)
	} else {
		f.set("targeting.droneControlRange", nil)
	}
	return f
}

// flattenAny flattens the oracle's nested stats; keys follow `order`'s key order first (so a diff lists
// in the same order as the TS harness), then any key only pyfa has, sorted.
func flattenAny(stats map[string]any, order *flat) *flat {
	tmp := map[string]any{}
	var walk func(prefix string, m map[string]any)
	walk = func(prefix string, m map[string]any) {
		for k, v := range m {
			key := k
			if prefix != "" {
				key = prefix + "." + k
			}
			if sub, ok := v.(map[string]any); ok {
				walk(key, sub)
			} else {
				tmp[key] = v
			}
		}
	}
	walk("", stats)
	out := newFlat()
	for _, k := range order.keys {
		if v, ok := tmp[k]; ok {
			out.set(k, v)
			delete(tmp, k)
		}
	}
	rest := make([]string, 0, len(tmp))
	for k := range tmp {
		rest = append(rest, k)
	}
	sort.Strings(rest)
	for _, k := range rest {
		out.set(k, tmp[k])
	}
	return out
}

type statDiff struct {
	Ship     string   `json:"ship"`
	ShipID   int      `json:"shipId"`
	FitType  string   `json:"fitType"`
	Key      string   `json:"key"`
	Ours     any      `json:"ours"`
	Pyfa     any      `json:"pyfa"`
	AbsDelta *float64 `json:"absDelta"`
	PctDelta *float64 `json:"pctDelta"`
}

func num(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case int:
		return float64(x), true
	}
	return 0, false
}

// diffStats returns every leaf that differs beyond tolerance. Booleans must match exactly; null on one
// side but a real number on the other is a difference (when the present side is non-trivial).
func diffStats(ours, pyfa *flat, tol, eps float64) []statDiff {
	var diffs []statDiff
	keys := append([]string(nil), ours.keys...)
	for _, k := range pyfa.keys {
		if _, ok := ours.vals[k]; !ok {
			keys = append(keys, k)
		}
	}
	for _, key := range keys {
		a, b := ours.vals[key], pyfa.vals[key]
		_, aBool := a.(bool)
		_, bBool := b.(bool)
		if aBool || bBool {
			if a != b {
				diffs = append(diffs, statDiff{Key: key, Ours: a, Pyfa: b})
			}
			continue
		}
		an, aok := num(a)
		bn, bok := num(b)
		if !aok && !bok {
			continue
		}
		if !aok || !bok || math.IsNaN(an) || math.IsNaN(bn) {
			present, has := an, aok
			if !aok {
				present, has = bn, bok
			}
			if has && math.Abs(present) > eps {
				diffs = append(diffs, statDiff{Key: key, Ours: a, Pyfa: b})
			}
			continue
		}
		abs := math.Abs(an - bn)
		if abs > math.Max(tol*math.Abs(bn), eps) {
			d := statDiff{Key: key, Ours: an, Pyfa: bn, AbsDelta: &abs}
			if bn != 0 {
				p := (an - bn) / math.Abs(bn) * 100
				d.PctDelta = &p
			}
			diffs = append(diffs, d)
		}
	}
	return diffs
}
