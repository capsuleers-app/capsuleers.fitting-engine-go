package main

// Attribute drift between OUR SDE bundle and the pinned pyfa's staticdata (same rule as
// test/diff/attr-drift.mjs of eve-fit-engine).
//
// pyfa's staticdata trails CCP's. A type our bundle has and pyfa doesn't is caught by the oracle
// (`dropped`); a type BOTH know but describe differently (CCP rebalanced it after the pin) used to come
// back as an engine diff: SDE 3720952 cut 97955's CPU from 25 to 16 and four fits read exactly -9 CPU.
//
// The rule, per type used by a compared fit that the pinned pyfa knows:
//   - compared: every attribute of OUR bundle's type (the bundle carries exactly what the engine
//     reads), plus the type-level mass / volume / capacity (pyfa stores them as attributes 4 / 161 / 38);
//   - drift: pyfa lacks the attribute, or |ours - pyfa| > 1e-6 * max(1, |pyfa|);
//   - not compared: attributes only pyfa has (its eve.db keeps some our bundle drops, e.g.
//     cloneGradeRestriction), and 1367 maxSubSystems, which pyfa's db_update.py forces from 5 to 4 on
//     purpose (not staticdata drift; not a compared stat).
//
// A fit using a drifted type is skipped and reported as oracle SDE drift, like an absent type.

import (
	"math"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

const attrMaxSubSystems = 1367

type attrDiff struct {
	Attr int      `json:"attr"`
	Name string   `json:"name"`
	Ours float64  `json:"ours"`
	Pyfa *float64 `json:"pyfa"`
}

type attrDriftEntry struct {
	TypeID int        `json:"typeID"`
	Name   string     `json:"name"`
	Attrs  []attrDiff `json:"attrs"`
	Fits   []string   `json:"fits"`
}

// specTypes lists every typeID a spec uses, in first-seen order.
func specTypes(spec fitSpec) []int {
	seen := map[int]bool{}
	var out []int
	add := func(id int) {
		if id != 0 && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	add(spec.ShipTypeID)
	for _, s := range spec.Subsystems {
		add(s)
	}
	if spec.ModeTypeID != nil {
		add(*spec.ModeTypeID)
	}
	for _, m := range spec.Modules {
		add(m.TypeID)
		if m.ChargeTypeID != nil {
			add(*m.ChargeTypeID)
		}
	}
	for _, d := range spec.Drones {
		add(d.TypeID)
	}
	for _, f := range spec.Fighters {
		add(f.TypeID)
	}
	for _, im := range spec.Implants {
		add(im.TypeID)
	}
	for _, b := range spec.Boosters {
		add(b.TypeID)
	}
	return out
}

// oracleAttrs asks the oracle for the pinned staticdata's base attributes of these types.
func oracleAttrs(ids []int) map[int]map[int]float64 {
	var res struct {
		Attrs map[string]map[string]float64 `json:"attrs"`
	}
	callOracle(map[string]any{"op": "attrs", "typeIDs": ids}, &res)
	out := map[int]map[int]float64{}
	for t, attrs := range res.Attrs {
		m := map[int]float64{}
		for a, v := range attrs {
			m[atoi(a)] = v
		}
		out[atoi(t)] = m
	}
	return out
}

func atoi(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		n = n*10 + int(s[i]-'0')
	}
	return n
}

func differs(ours, pyfa float64) bool {
	return math.Abs(ours-pyfa) > 1e-6*math.Max(1, math.Abs(pyfa))
}

// typeAttrDrift: the attributes of one type that differ from the pinned pyfa (nil when none, or when
// pyfa doesn't know the type: that case is the oracle's `dropped`).
func typeAttrDrift(ds *types.FittingDataset, t *types.SdeType, py map[int]float64) []attrDiff {
	if py == nil {
		return nil
	}
	name := func(id int) string {
		if a, ok := ds.Attributes.Get(id); ok && a.Name != "" {
			return a.Name
		}
		return jsmath.NumberToString(float64(id))
	}
	var out []attrDiff
	check := func(id int, label string, ours float64) {
		pv, ok := py[id]
		if !ok {
			out = append(out, attrDiff{id, label, ours, nil})
		} else if differs(ours, pv) {
			p := pv
			out = append(out, attrDiff{id, label, ours, &p})
		}
	}
	own := map[int]bool{}
	for _, a := range t.Attributes {
		own[a.ID] = true
		if a.ID == attrMaxSubSystems {
			continue
		}
		check(a.ID, name(a.ID), a.V)
	}
	for _, f := range []struct {
		id    int
		label string
		v     *float64
	}{{4, "mass", t.Mass}, {161, "volume", t.Volume}, {38, "capacity", t.Capacity}} {
		if f.v != nil && !own[f.id] {
			check(f.id, f.label, *f.v)
		}
	}
	return out
}

func fmtNum(v *float64) string {
	if v == nil {
		return "—"
	}
	return jsmath.NumberToString(*v)
}
