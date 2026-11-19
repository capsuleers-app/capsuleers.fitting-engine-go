package main

// Deterministic 4-fits-per-ship generator (port of test/diff/fit-generator.mjs of eve-fit-engine; the
// two repos must build the SAME corpus, so this stays a literal port, RNG consumption included).
//
// For each ship:
//   - bonused      weapons of the ship's primary hardpoint system + matching mods
//   - non-bonused  the other weapon system / off-bonus modules
//   - t2           every slot filled with T2 (metaGroup 2) modules
//   - mixed        weapons + a spread of T1/T2/faction/deadspace/officer mods
//
// Slot/hardpoint/drone capacities are read from OUR engine (ComputeFit on a base ship+subsystems fit)
// so it works for normal hulls AND T3Cs. Fits are best-effort and may over/under-fill: both engines
// compute the same thing, and the diff is what matters. The RNG is seeded by shipTypeID.

import (
	"math"
	"regexp"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine/checks"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

const (
	metaT1        = 1
	metaT2        = 2
	metaFaction   = 4
	metaOfficer   = 5
	metaDeadspace = 6

	strategicCruiserGroup  = 963
	propulsionModuleGroup  = 46
	attrRigSize            = 1547
	attrCPU                = 50
	attrPower              = 30
	attrVolume             = 283
	attrDroneBandwidth     = 1271
	attrDroneBandwidthUsed = 1272
	attrFitsToShipType     = 1380
)

var slots = []types.SlotType{types.SlotHI, types.SlotMED, types.SlotLO, types.SlotRIG}

var subsystemGroups = []int{954, 955, 956, 957, 958} // defensive/electronic/offensive/propulsion/core

// excludedModuleGroups: target-dependent ewar pyfa models optimistically without a target (Nosferatu),
// non-combat industrial/utility modules, remote assistance, capital-special/fleet modules and
// ship-restricted "mode" modules (Siege/Triage/Bastion/Industrial Core + Entosis). See the TS source
// for the per-group rationale.
var excludedModuleGroups = setOf(
	68,
	47, 48,
	54, 464, 483, 538, 546,
	650, 1122, 737, 4138,
	658,
	41, 67, 209, 290, 325, 585, 1697, 1698, 2018,
	588, 842, 1815,
	590, 4127, 4184, 815,
	1706, 4174,
	899, 1533, 1770,
	515, 1313,
)

// specialWeaponGroups: Defender, Bomb, Breacher Pod, Festival launchers (bespoke damage semantics).
var specialWeaponGroups = setOf(512, 862, 4807, 501)

var polarizedOrCivilian = regexp.MustCompile(`(?i)Polarized|Civilian`)

func setOf(ids ...int) map[int]bool {
	m := make(map[int]bool, len(ids))
	for _, id := range ids {
		m[id] = true
	}
	return m
}

// attrV is `t.attributes.find(a => a.id === id)?.v`.
func attrV(t *types.SdeType, id int) (float64, bool) {
	if t == nil {
		return 0, false
	}
	for _, a := range t.Attributes {
		if a.ID == id {
			return a.V, true
		}
	}
	return 0, false
}

func metaOf(t *types.SdeType) int {
	if t.MetaGroupID != nil {
		return *t.MetaGroupID
	}
	return 1
}

// sizeOk rejects size-inappropriate modules: a rig whose rigSize != the ship's, or any module that
// alone exceeds the ship's CPU or PG output.
func sizeOk(t, ship *types.SdeType, cpuMax, pgMax float64) bool {
	if rs, ok := attrV(t, attrRigSize); ok {
		if srs, ok2 := attrV(ship, attrRigSize); ok2 && rs != srs {
			return false
		}
	}
	if cpu, ok := attrV(t, attrCPU); ok && truthy(cpuMax) && cpu > cpuMax {
		return false
	}
	if pg, ok := attrV(t, attrPower); ok && truthy(pgMax) && pg > pgMax {
		return false
	}
	return true
}

func truthy(f float64) bool { return f != 0 && !math.IsNaN(f) }

// rng is mulberry32, bit-exact with the JS one.
func rng(seed uint32) func() float64 {
	a := seed
	return func() float64 {
		a += 0x6D2B79F5
		t := (a ^ a>>15) * (1 | a)
		t = (t + (t^t>>7)*(61|t)) ^ t
		return float64(t^t>>14) / 4294967296
	}
}

func pick(arr []*types.SdeType, r func() float64) *types.SdeType {
	if len(arr) == 0 {
		return nil // consumes no RNG, like the JS
	}
	return arr[int(math.Floor(r()*float64(len(arr))))]
}

func shuffle(arr []*types.SdeType, r func() float64) []*types.SdeType {
	a := append([]*types.SdeType(nil), arr...)
	for i := len(a) - 1; i > 0; i-- {
		j := int(math.Floor(r() * float64(i+1)))
		a[i], a[j] = a[j], a[i]
	}
	return a
}

func filter(arr []*types.SdeType, keep func(*types.SdeType) bool) []*types.SdeType {
	var out []*types.SdeType
	for _, t := range arr {
		if keep(t) {
			out = append(out, t)
		}
	}
	return out
}

func bucket(ds *types.FittingDataset, b types.TypeBucket) []*types.SdeType {
	tbl := ds.TypesByBucket[b]
	out := make([]*types.SdeType, tbl.Len())
	for i := range out {
		_, out[i] = tbl.At(i)
	}
	return out
}

type pool struct {
	bySlot             map[types.SlotType][]*types.SdeType
	turrets, launchers []*types.SdeType
	charges            []*types.SdeType
}

var globalPool *pool

func buildPool(ds *types.FittingDataset) *pool {
	if globalPool != nil {
		return globalPool
	}
	p := &pool{bySlot: map[types.SlotType][]*types.SdeType{}}
	for _, t := range bucket(ds, types.BucketModules) {
		for _, s := range slots {
			if checks.TypeFitsSlotType(t, s) {
				p.bySlot[s] = append(p.bySlot[s], t)
			}
		}
	}
	normal := func(t *types.SdeType) bool {
		return !polarizedOrCivilian.MatchString(t.Name) && !specialWeaponGroups[t.GroupID]
	}
	p.turrets = filter(p.bySlot[types.SlotHI], func(t *types.SdeType) bool { return checks.IsTurretWeapon(t) && normal(t) })
	p.launchers = filter(p.bySlot[types.SlotHI], func(t *types.SdeType) bool { return checks.IsMissileLauncher(t) && normal(t) })
	p.charges = bucket(ds, types.BucketCharges)
	globalPool = p
	return p
}

func chargeFor(p *pool, mod *types.SdeType, r func() float64) *int {
	if !checks.ModuleAcceptsAnyCharge(mod) {
		return nil
	}
	// Validate group AND size: an oversized charge is rejected by pyfa but trusted by our engine.
	c := pick(filter(p.charges, func(c *types.SdeType) bool { return checks.ModuleAcceptsChargeType(mod, c) }), r)
	if c == nil {
		return nil
	}
	id := c.ID
	return &id
}

type specModule struct {
	ID           string
	Position     int
	SlotType     types.SlotType
	TypeID       int
	State        types.ModuleState
	ChargeTypeID *int
}

type specDrone struct{ TypeID, Count, Active int }

type specFighter struct{ TypeID, Count int }

type specImplant struct {
	TypeID int
	Slot   *float64
}

type specBooster struct {
	TypeID            int
	Slot              *float64
	ActiveSideEffects []int
}

// fitSpec is the engine-agnostic spec every source emits.
type fitSpec struct {
	ShipTypeID int
	FitType    string
	Modules    []specModule
	Drones     []specDrone
	Fighters   []specFighter
	Subsystems []int
	Implants   []specImplant
	Boosters   []specBooster
	ModeTypeID *int
}

func weaponList(p *pool, kind string) []*types.SdeType {
	if kind == "launcher" {
		return p.launchers
	}
	return p.turrets
}

// The TS generator also filters with `canFitModuleOnShip(ship, t, dataset) && ...`, but with its
// arguments in the wrong order the call returns an object, which is always truthy: that filter has never
// excluded anything. It is left out here so both repos keep generating the SAME corpus; fix it in both
// at once if it is ever fixed.

func fillWeapons(ds *types.FittingDataset, ship *types.SdeType, p *pool, kind string, count, meta int, r func() float64, cpuMax, pgMax float64) []specModule {
	if count <= 0 {
		return nil
	}
	list := filter(weaponList(p, kind), func(t *types.SdeType) bool { return sizeOk(t, ship, cpuMax, pgMax) })
	preferred := filter(list, func(t *types.SdeType) bool { return metaOf(t) == meta })
	base := list
	if len(preferred) > 0 {
		base = preferred
	}
	if len(base) == 0 {
		return nil
	}
	w := pick(base, r)
	out := make([]specModule, 0, count)
	for i := 0; i < count; i++ {
		state := checks.DefaultStateForModule(w, &ds.Effects)
		out = append(out, specModule{SlotType: types.SlotHI, TypeID: w.ID, State: state, ChargeTypeID: chargeFor(p, w, r)})
	}
	return out
}

func fillSlot(ds *types.FittingDataset, ship *types.SdeType, p *pool, slot types.SlotType, count int, metaPicker func(int) int, r func() float64, cpuMax, pgMax float64) []specModule {
	if count <= 0 {
		return nil
	}
	list := filter(p.bySlot[slot], func(t *types.SdeType) bool {
		return sizeOk(t, ship, cpuMax, pgMax) && !excludedModuleGroups[t.GroupID]
	})
	// Leftover HIGH slots are utility: never fill them with weapons.
	if slot == types.SlotHI {
		list = filter(list, func(t *types.SdeType) bool { return !checks.IsTurretWeapon(t) && !checks.IsMissileLauncher(t) })
	}
	var out []specModule
	usedGroups := map[int]int{}
	propUsed := 0
	for i := 0; i < count; i++ {
		meta := metaPicker(i)
		cands := shuffle(filter(list, func(t *types.SdeType) bool { return metaOf(t) == meta }), r)
		if len(cands) == 0 {
			cands = shuffle(list, r)
		}
		var m *types.SdeType
		for _, t := range cands {
			if usedGroups[t.GroupID] < 2 && !(t.GroupID == propulsionModuleGroup && propUsed >= 1) {
				m = t
				break
			}
		}
		if m == nil && len(cands) > 0 {
			m = cands[0]
		}
		if m == nil {
			break
		}
		if m.GroupID == propulsionModuleGroup {
			if propUsed >= 1 { // would be a 2nd prop mod: pick a non-prop instead
				var alt *types.SdeType
				for _, t := range cands {
					if t.GroupID != propulsionModuleGroup && usedGroups[t.GroupID] < 2 {
						alt = t
						break
					}
				}
				if alt == nil {
					continue
				}
				m = alt
			}
			if m.GroupID == propulsionModuleGroup {
				propUsed++
			}
		}
		usedGroups[m.GroupID]++
		out = append(out, specModule{SlotType: slot, TypeID: m.ID, State: checks.DefaultStateForModule(m, &ds.Effects), ChargeTypeID: chargeFor(p, m, r)})
	}
	return out
}

func droneSet(ds *types.FittingDataset, bayMax, bwMax float64, r func() float64) []specDrone {
	if !truthy(bayMax) || bayMax <= 0 {
		return nil
	}
	for _, d := range shuffle(bucket(ds, types.BucketDrones), r) {
		vol := 5.0
		if d.Volume != nil {
			vol = *d.Volume
		} else if v, ok := attrV(d, attrVolume); ok {
			vol = v
		}
		bw := vol
		if v, ok := attrV(d, attrDroneBandwidth); ok {
			bw = v
		}
		if !truthy(vol) || !truthy(bw) || bw > bwMax {
			continue
		}
		byBay := math.Floor(bayMax / vol)
		byBw := math.Floor(bwMax / bw)
		n := int(math.Max(1, math.Min(5, math.Min(byBay, byBw))))
		return []specDrone{{TypeID: d.ID, Count: n, Active: n}}
	}
	return nil
}

func chooseSubsystems(ds *types.FittingDataset, ship *types.SdeType) []int {
	subs := bucket(ds, types.BucketSubsystems)
	var out []int
	for _, g := range subsystemGroups {
		for _, t := range subs {
			if v, ok := attrV(t, attrFitsToShipType); t.GroupID == g && ok && v == float64(ship.ID) {
				out = append(out, t.ID)
				break
			}
		}
	}
	return out
}

func baseFit(shipTypeID int, subsystems []types.FitSubsystem) *types.Fit {
	return &types.Fit{
		ShipTypeID: shipTypeID, Name: "base", Visibility: "PRIVATE", Tags: []string{},
		Modules: []types.FitModule{}, Drones: []types.FitDrone{}, Fighters: []types.FitFighter{},
		Cargo: []types.FitCargo{}, Implants: []types.FitImplant{}, Boosters: []types.FitBooster{},
		Subsystems: subsystems,
	}
}

// baseCaps resolves the hull's real slot/hardpoint/drone capacities through our engine.
func baseCaps(ds *types.FittingDataset, fit *types.Fit, skills *types.SkillProfile) (cap *types.Fitting) {
	defer func() {
		if recover() != nil {
			cap = nil
		}
	}()
	c, err := fitengine.ComputeFit(fit, ds, &fitengine.ComputeFitOptions{SkillProfile: skills})
	if err != nil {
		return nil
	}
	return &c.Derived.Fitting
}

// generateFits returns the 4 fit-specs for a ship. skills MUST be a real All-V profile so the resolved
// CPU/PG/drone caps match what both engines compute.
func generateFits(ds *types.FittingDataset, ship *types.SdeType, skills *types.SkillProfile) []fitSpec {
	p := buildPool(ds)
	var subsystems []int
	if ship.GroupID == strategicCruiserGroup {
		subsystems = chooseSubsystems(ds, ship)
	}
	var modeTypeID *int
	if m, ok := checks.DefaultModeTypeIDFor(ship, ds); ok {
		modeTypeID = &m
	}
	fitSubs := make([]types.FitSubsystem, len(subsystems))
	for i, s := range subsystems {
		fitSubs[i] = types.FitSubsystem{ID: "s" + itoa(i), Slot: i + 1, TypeID: s}
	}
	cap := baseCaps(ds, baseFit(ship.ID, fitSubs), skills)
	var hi, med, lo, rig, turretMax, launcherMax, bayMax, bwMax, cpuMax, pgMax float64
	if cap != nil {
		hi, med, lo, rig = cap.Slots.HI.Max, cap.Slots.MED.Max, cap.Slots.LO.Max, cap.Slots.RIG.Max
		turretMax, launcherMax = cap.Hardpoints.Turret.Max, cap.Hardpoints.Launcher.Max
		bayMax, bwMax = cap.DroneBayMax, cap.DroneBandwidthMax
		cpuMax, pgMax = cap.CpuMax, cap.PowerMax
	}
	primaryKind, otherKind := "turret", "launcher"
	if launcherMax > turretMax {
		primaryKind, otherKind = "launcher", "turret"
	}
	wpCount := int(math.Max(turretMax, launcherMax))

	assemble := func(fitType string, weapons func(r func() float64) []specModule, metaForSlot func(int) int) fitSpec {
		r := rng(uint32(ship.ID*31 + len(fitType)*7))
		w := weapons(r)
		mods := append([]specModule(nil), w...)
		mods = append(mods, fillSlot(ds, ship, p, types.SlotHI, int(math.Max(0, hi-float64(len(w)))), metaForSlot, r, cpuMax, pgMax)...)
		mods = append(mods, fillSlot(ds, ship, p, types.SlotMED, int(med), metaForSlot, r, cpuMax, pgMax)...)
		mods = append(mods, fillSlot(ds, ship, p, types.SlotLO, int(lo), metaForSlot, r, cpuMax, pgMax)...)
		mods = append(mods, fillSlot(ds, ship, p, types.SlotRIG, int(rig), metaForSlot, r, cpuMax, pgMax)...)
		for i := range mods {
			mods[i].ID = "m" + itoa(i)
			mods[i].Position = i
		}
		return fitSpec{ShipTypeID: ship.ID, FitType: fitType, Modules: mods, Drones: droneSet(ds, bayMax, bwMax, r),
			Subsystems: subsystems, ModeTypeID: modeTypeID}
	}
	const2 := func(v int) func(int) int { return func(int) int { return v } }
	mixed := []int{metaT1, metaT2, metaFaction, metaDeadspace, metaOfficer}
	return []fitSpec{
		assemble("bonused", func(r func() float64) []specModule {
			return fillWeapons(ds, ship, p, primaryKind, wpCount, metaT2, r, cpuMax, pgMax)
		}, const2(metaT2)),
		assemble("non-bonused", func(r func() float64) []specModule {
			return fillWeapons(ds, ship, p, otherKind, wpCount, metaT1, r, cpuMax, pgMax)
		}, const2(metaT1)),
		assemble("t2", func(r func() float64) []specModule {
			return fillWeapons(ds, ship, p, primaryKind, wpCount, metaT2, r, cpuMax, pgMax)
		}, const2(metaT2)),
		assemble("mixed", func(r func() float64) []specModule {
			return fillWeapons(ds, ship, p, primaryKind, wpCount, metaFaction, r, cpuMax, pgMax)
		}, func(i int) int { return mixed[i%5] }),
	}
}
