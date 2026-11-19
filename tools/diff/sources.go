package main

// Fit SOURCES for the differential harness (port of test/diff/fit-sources.mjs of eve-fit-engine).
//
//   generated  (generator.go) 4 synthetic fits for EVERY published ship: complete hull coverage, but
//              built by our own slot/charge predicates.
//   workbench  2-3 REAL fits per hull published on EVE Workbench (.workbench/corpus.json,
//              `go run ./tools/workbench`): combinations no generator proposes.
//   implants   implant SETS and BOOSTERS, which neither of the other two carries.
//
// Every source emits the same engine-agnostic fitSpec, so both engines get identical input.

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine/eft"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

const (
	attrImplantSlot = 331
	attrBoosterSlot = 1087
	boosterGroup    = 303
)

var corpusFile = filepath.Join(".workbench", "corpus.json")

type corpusRow struct {
	ID         string  `json:"id"`
	ShipTypeID int     `json:"shipTypeID"`
	ShipName   string  `json:"shipName"`
	Name       string  `json:"name"`
	Dps        float64 `json:"dps"`
	Eft        string  `json:"eft"`
}

type corpusFailure struct{ id, ship, reason string }

// resolveDroneCounts decides the active count ONCE, from bay volume and bandwidth, so the two engines
// are handed the same number (pyfa trusts amountActive verbatim, ours re-derives from bandwidth).
func resolveDroneCounts(ds *types.FittingDataset, drones []types.FitDrone, caps *types.Fitting) []specDrone {
	var bayLeft, bwLeft float64
	if caps != nil {
		bayLeft, bwLeft = caps.DroneBayMax, caps.DroneBandwidthMax
	}
	var out []specDrone
	for _, d := range drones {
		t, ok := ds.TypesByBucket[types.BucketDrones].Get(d.TypeID)
		if !ok {
			continue
		}
		vol := 5.0
		if t.Volume != nil {
			vol = *t.Volume
		} else if v, ok := attrV(t, attrVolume); ok {
			vol = v
		}
		bw := vol
		if v, ok := attrV(t, attrDroneBandwidthUsed); ok {
			bw = v
		}
		total := math.Max(1, d.CountTotal)
		byBay, byBw := total, total
		if vol > 0 {
			byBay = math.Floor(bayLeft / vol)
		}
		if bw > 0 {
			byBw = math.Floor(bwLeft / bw)
		}
		active := math.Max(0, math.Min(total, math.Min(byBay, byBw)))
		out = append(out, specDrone{TypeID: d.TypeID, Count: int(total), Active: int(active)})
		bayLeft -= vol * math.Min(total, byBay)
		bwLeft -= bw * active
	}
	return out
}

// loadWorkbenchFits reads the harvested corpus and turns each EFT into a spec. A fit whose EFT we cannot
// parse is REPORTED, never silently dropped.
func loadWorkbenchFits(ds *types.FittingDataset, skills *types.SkillProfile, perHull int) (fits []fitSpec, failures []corpusFailure, missing bool, err error) {
	raw, err := os.ReadFile(corpusFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, true, nil
	}
	if err != nil {
		return nil, nil, false, err
	}
	var corpus []corpusRow
	if err := json.Unmarshal(raw, &corpus); err != nil {
		return nil, nil, false, err
	}
	seen := map[int]int{}
	for _, row := range corpus {
		n := seen[row.ShipTypeID]
		if n >= perHull {
			continue
		}
		parsed, perr := parseEftSafe(row.Eft, ds)
		if perr != nil {
			failures = append(failures, corpusFailure{row.ID, row.ShipName, perr.Error()})
			continue
		}
		fit := parsed.Fit
		if !ds.TypesByBucket[types.BucketShips].Has(fit.ShipTypeID) {
			failures = append(failures, corpusFailure{row.ID, row.ShipName, "ship " + itoa(fit.ShipTypeID) + " not in bundle"})
			continue
		}
		subs := fit.Subsystems
		if subs == nil {
			subs = []types.FitSubsystem{}
		}
		caps := baseCaps(ds, baseFit(fit.ShipTypeID, subs), skills)
		seen[row.ShipTypeID] = n + 1
		spec := fitSpec{ShipTypeID: fit.ShipTypeID, FitType: "wb" + itoa(n+1), ModeTypeID: fit.ModeTypeID}
		for i, m := range fit.Modules {
			spec.Modules = append(spec.Modules, specModule{ID: "m" + itoa(i), Position: i, SlotType: m.SlotType,
				TypeID: m.TypeID, State: m.State, ChargeTypeID: m.ChargeTypeID})
		}
		spec.Drones = resolveDroneCounts(ds, fit.Drones, caps)
		for _, f := range fit.Fighters {
			spec.Fighters = append(spec.Fighters, specFighter{TypeID: f.TypeID, Count: int(f.Count)})
		}
		for _, s := range fit.Subsystems {
			spec.Subsystems = append(spec.Subsystems, s.TypeID)
		}
		for _, im := range fit.Implants {
			slot := im.Slot
			spec.Implants = append(spec.Implants, specImplant{TypeID: im.TypeID, Slot: &slot})
		}
		for _, b := range fit.Boosters {
			slot := b.Slot
			spec.Boosters = append(spec.Boosters, specBooster{TypeID: b.TypeID, Slot: &slot, ActiveSideEffects: []int{}})
		}
		fits = append(fits, spec)
	}
	return fits, failures, false, nil
}

func parseEftSafe(text string, ds *types.FittingDataset) (res *eft.ParseResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.New("parser panic")
		}
	}()
	return eft.ParseEft(text, ds)
}

type implantSet struct {
	name    string
	members []*types.SdeType
}

var setNameRe = regexp.MustCompile(`^((?:Low-grade|Mid-grade|High-grade)\s+\S+)\s+(?:Alpha|Beta|Gamma|Delta|Epsilon|Omega)$`)
var gradeRe = regexp.MustCompile(`(Low|Mid|High)-grade`)
var spacesRe = regexp.MustCompile(`\s+`)

// discoverImplantSets: implants named "<grade> <Family> <Greek letter>", grouped by family, one member
// per slot. Sets are a mechanic (attr 802 set multiplier), discovered from the data, not hardcoded.
func discoverImplantSets(ds *types.FittingDataset) []implantSet {
	var famOrder []string
	byFamily := map[string][]*types.SdeType{}
	for _, t := range bucket(ds, types.BucketImplants) {
		if t.GroupID == boosterGroup {
			continue
		}
		m := setNameRe.FindStringSubmatch(t.Name)
		if m == nil {
			continue
		}
		if _, ok := byFamily[m[1]]; !ok {
			famOrder = append(famOrder, m[1])
		}
		byFamily[m[1]] = append(byFamily[m[1]], t)
	}
	var sets []implantSet
	for _, fam := range famOrder {
		bySlot := map[float64]*types.SdeType{}
		var slotsSeen []float64
		for _, t := range byFamily[fam] {
			if slot, ok := attrV(t, attrImplantSlot); ok {
				if _, dup := bySlot[slot]; !dup {
					bySlot[slot] = t
					slotsSeen = append(slotsSeen, slot)
				}
			}
		}
		if len(slotsSeen) >= 2 {
			sort.SliceStable(slotsSeen, func(i, j int) bool { return slotsSeen[i] < slotsSeen[j] })
			set := implantSet{name: fam}
			for _, s := range slotsSeen {
				set.members = append(set.members, bySlot[s])
			}
			sets = append(sets, set)
		}
	}
	sort.SliceStable(sets, func(i, j int) bool { return jsmath.LocaleCompare(sets[i].name, sets[j].name) < 0 })
	return sets
}

// hardwiringsBySlot: the non-set implants, grouped by slot, each list sorted by id.
func hardwiringsBySlot(ds *types.FittingDataset) map[float64][]*types.SdeType {
	bySlot := map[float64][]*types.SdeType{}
	for _, t := range bucket(ds, types.BucketImplants) {
		if t.GroupID == boosterGroup || gradeRe.MatchString(t.Name) {
			continue
		}
		slot, ok := attrV(t, attrImplantSlot)
		if !ok {
			continue
		}
		bySlot[slot] = append(bySlot[slot], t)
	}
	for _, list := range bySlot {
		sort.SliceStable(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	}
	return bySlot
}

// shipAffectingBoosters: boosters any of whose effects modifies the SHIP (the combat boosters plus a few
// speciality ones); cerebral accelerators and the like only move character attributes.
func shipAffectingBoosters(ds *types.FittingDataset) []*types.SdeType {
	var out []*types.SdeType
	for _, t := range bucket(ds, types.BucketImplants) {
		if t.GroupID != boosterGroup {
			continue
		}
		touches := false
		for _, e := range t.Effects {
			if eff, ok := ds.Effects.Get(e.ID); ok {
				for _, mi := range eff.ModifierInfo {
					if mi.Domain == "shipID" {
						touches = true
					}
				}
			}
		}
		if touches {
			out = append(out, t)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func sideEffectIDs(booster *types.SdeType) []int {
	var out []int
	for _, e := range booster.Effects {
		if engine.IsBoosterSideEffectID(e.ID) {
			out = append(out, e.ID)
		}
	}
	return out
}

func f64(v float64) *float64 { return &v }

func setFit(ship *types.SdeType, set implantSet) fitSpec {
	spec := fitSpec{ShipTypeID: ship.ID, FitType: "set:" + spacesRe.ReplaceAllString(set.name, "-")}
	for _, t := range set.members {
		var slot *float64
		if v, ok := attrV(t, attrImplantSlot); ok {
			slot = f64(v)
		}
		spec.Implants = append(spec.Implants, specImplant{TypeID: t.ID, Slot: slot})
	}
	return spec
}

// boosterFits: a booster with its drawbacks off and on; the pair isolates the opt-in path.
func boosterFits(ship *types.SdeType, b *types.SdeType) []fitSpec {
	slot := 1.0
	if v, ok := attrV(b, attrBoosterSlot); ok {
		slot = v
	}
	tag := spacesRe.ReplaceAllString(b.Name, "-")
	if b.Name == "" {
		tag = itoa(b.ID)
	}
	side := sideEffectIDs(b)
	fits := []fitSpec{{ShipTypeID: ship.ID, FitType: "boost:" + tag,
		Boosters: []specBooster{{TypeID: b.ID, Slot: f64(slot), ActiveSideEffects: []int{}}}}}
	if len(side) > 0 {
		fits = append(fits, fitSpec{ShipTypeID: ship.ID, FitType: "boost:" + tag + ":side",
			Boosters: []specBooster{{TypeID: b.ID, Slot: f64(slot), ActiveSideEffects: side}}})
	}
	return fits
}

// hardwiringFit: a full hardwiring rack, one per slot (plain bonuses that must not penalise each other).
func hardwiringFit(ship *types.SdeType, bySlot map[float64][]*types.SdeType) []fitSpec {
	var rack []specImplant
	for _, slot := range []float64{6, 7, 8, 9, 10} {
		if list := bySlot[slot]; len(list) > 0 {
			rack = append(rack, specImplant{TypeID: list[len(list)/2].ID, Slot: f64(slot)})
		}
	}
	if len(rack) == 0 {
		return nil
	}
	return []fitSpec{{ShipTypeID: ship.ID, FitType: "hardwirings", Implants: rack}}
}

// generateImplantMatrix: the FULL implant/booster matrix on one (bare) hull: every set, every
// ship-affecting booster with drawbacks off and on, plus the hardwiring rack.
func generateImplantMatrix(ds *types.FittingDataset, ship *types.SdeType) []fitSpec {
	var out []fitSpec
	for _, set := range discoverImplantSets(ds) {
		out = append(out, setFit(ship, set))
	}
	out = append(out, hardwiringFit(ship, hardwiringsBySlot(ds))...)
	for _, b := range shipAffectingBoosters(ds) {
		out = append(out, boosterFits(ship, b)...)
	}
	return out
}

// generateImplantFits: a deterministic rotating SLICE of the matrix for one hull.
func generateImplantFits(ds *types.FittingDataset, ship *types.SdeType, rotation int) []fitSpec {
	sets := discoverImplantSets(ds)
	boosters := shipAffectingBoosters(ds)
	var out []fitSpec
	if len(sets) > 0 {
		out = append(out, setFit(ship, sets[rotation%len(sets)]))
		out = append(out, setFit(ship, sets[(rotation*7+3)%len(sets)]))
	}
	out = append(out, hardwiringFit(ship, hardwiringsBySlot(ds))...)
	if len(boosters) > 0 {
		out = append(out, boosterFits(ship, boosters[rotation%len(boosters)])...)
	}
	return out
}
