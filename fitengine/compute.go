package fitengine

// Porting letterale di src/engine.ts (parte finale): ComputeFit, deriveStats, buildProjectionReport.
// La fase dei modificatori e' in internal/engine (ApplyModifiers); qui si calcolano le statistiche derivate.

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/derived"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/effects"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

type (
	ComputeFitOptions     = engine.ComputeFitOptions
	ComputedAttribute     = engine.ComputedAttribute
	Fit                   = types.Fit
	SkillProfile          = types.SkillProfile
	DamageProfile         = types.DamageProfile
	TargetProfile         = types.TargetProfile
	ProjectedSource       = types.ProjectedSource
	ProjectedEffectReport = types.ProjectedEffectReport
	WeaponContribution    = types.WeaponContribution
)

// ModuleComputed, DroneComputed, FighterComputed (types.ts).
type ModuleComputed struct {
	FitModuleID    string
	TypeID         int
	SlotType       types.SlotType
	State          types.ModuleState
	Attributes     *omap.Map[int, ComputedAttribute]
	EffectiveCpu   float64
	EffectivePower float64
}

type DroneComputed struct {
	FitDroneID string
	TypeID     int
	Attributes *omap.Map[int, ComputedAttribute]
	Dps        float64
}

type FighterAbility struct {
	EffectID int
	Enabled  bool
	Dps      float64
}

type FighterComputed struct {
	FitFighterID string
	TypeID       int
	Attributes   *omap.Map[int, ComputedAttribute]
	Abilities    []FighterAbility
}

// ModuleAttrSnapshot: valori finali di ogni attributo del modulo e della sua carica (nil = nessuna carica).
// I Record<number, number> del TS sono in ordine di id ascendente (ordine di iterazione JS).
type ModuleAttrSnapshot struct {
	Module *omap.Map[int, float64]
	Charge *omap.Map[int, float64] `encode:"null"`
}

type DerivedStats struct {
	Fitting         types.Fitting
	Defense         types.Defense
	Offense         types.Offense
	Capacitor       types.CapacitorStats
	Tank            types.TankRates
	Navigation      types.Navigation
	Targeting       types.Targeting
	Drones          types.DroneStats
	Projected       []types.ProjectedEffectReport
	Structure       *types.StructureMeta `encode:"null"`
	ModuleSnapshots *omap.Map[string, ModuleAttrSnapshot]
}

type ComputedFit struct {
	Fit      *types.Fit
	Ship     *omap.Map[int, ComputedAttribute]
	Modules  *omap.Map[string, ModuleComputed]
	Drones   *omap.Map[string, DroneComputed]
	Fighters *omap.Map[string, FighterComputed]
	Derived  DerivedStats
}

// ComputeFit calcola il blocco completo di statistiche di un fit contro un dataset.
func ComputeFit(fit *types.Fit, dataset *types.FittingDataset, opts *ComputeFitOptions) (*ComputedFit, error) {
	o := *opts
	o.BuildProjectionReport = buildProjectionReport
	st, err := engine.ApplyModifiers(fit, dataset, &o)
	if err != nil {
		return nil, err
	}

	modules := omap.New[string, ModuleComputed]()
	for _, m := range st.Modules {
		slot, ok := m.SlotType()
		if !ok {
			slot = types.SlotLO
		}
		modules.Set(m.ID, ModuleComputed{
			FitModuleID: m.ID, TypeID: m.TypeID, SlotType: slot, State: m.State, Attributes: engine.SnapshotMap(m),
			EffectiveCpu: m.GetFinal(engine.AttrCpuUsed, 0), EffectivePower: m.GetFinal(engine.AttrPowerUsed, 0),
		})
	}
	drones := omap.New[string, DroneComputed]()
	for _, d := range st.Drones {
		drones.Set(d.ID, DroneComputed{FitDroneID: d.ID, TypeID: d.TypeID, Attributes: engine.SnapshotMap(d), Dps: 0})
	}
	fighters := omap.New[string, FighterComputed]()
	for _, f := range st.Fighters {
		fighters.Set(f.ID, FighterComputed{FitFighterID: f.ID, TypeID: f.TypeID, Attributes: engine.SnapshotMap(f), Abilities: []FighterAbility{}})
	}
	return &ComputedFit{
		Fit: fit, Ship: engine.SnapshotMap(st.Ship), Modules: modules, Drones: drones, Fighters: fighters,
		Derived: deriveStats(st.Ctx, dataset, opts.DamageProfile, fit, st.ProjectionReports),
	}, nil
}

// buildProjectionReport e' buildProjectionReport di engine.ts (nessun attributo viene modificato).
func buildProjectionReport(source *engine.ItemState, ctx *engine.FitContext) types.ProjectedEffectReport {
	cls := effects.ClassifyEwar(source)
	if cls == nil {
		return types.ProjectedEffectReport{TypeID: source.TypeID, Kind: "OTHER", Summary: string(source.Kind) + " module projected"}
	}
	if cls.Kind == "ECM" {
		chance := effects.EcmJamChance(source, ctx.Ship)
		return types.ProjectedEffectReport{
			TypeID: source.TypeID, Kind: "ECM", JamChance: &chance,
			Summary: "Per-cycle jam chance " + jsmath.ToFixed(chance*100, 1) + "%",
		}
	}
	summary := string(cls.Kind)
	switch cls.Kind {
	case "SENSOR_DAMP":
		summary = "Sensor strength / lock range reduced"
	case "TRACKING_DISRUPT":
		summary = "Turret tracking / range reduced"
	case "WEB":
		summary = "Max velocity reduced"
	case "WARP_SCRAM":
		summary = "Warp ability disabled"
	case "WARP_DISRUPT":
		summary = "Warp disrupted (point)"
	case "NEUT":
		summary = "Capacitor drain incoming"
	case "NOS":
		summary = "Capacitor siphoned"
	}
	return types.ProjectedEffectReport{TypeID: source.TypeID, Kind: string(cls.Kind), Summary: summary}
}

// Slot base che ogni capsuleer ha prima delle skill di targeting.
const charBaseMaxLockedTargets = 2

func isTurretEffect(eid int) bool   { return eid == 10 || eid == 34 || eid == 6995 || eid == 8037 }
func isLauncherEffect(eid int) bool { return eid == 9 || eid == 101 }

func deriveStats(ctx *engine.FitContext, dataset *types.FittingDataset, damageProfile *types.DamageProfile, fit *types.Fit, projectionReports []types.ProjectedEffectReport) DerivedStats {
	ship := ctx.Ship

	var slotUsed types.SlotTable
	cpuUsed, powerUsed, calibUsed := 0.0, 0.0, 0.0
	turretHardpointsUsed, launcherHardpointsUsed := 0.0, 0.0
	for _, m := range ctx.Modules {
		if m.State == types.StateOffline {
			continue // i moduli offline non contano per il fitting
		}
		slot, hasSlot := m.SlotType()
		if hasSlot {
			switch slot {
			case types.SlotHI:
				slotUsed.HI.Used++
			case types.SlotMED:
				slotUsed.MED.Used++
			case types.SlotLO:
				slotUsed.LO.Used++
			case types.SlotRIG:
				slotUsed.RIG.Used++
			case types.SlotSUBSYSTEM:
				slotUsed.SUBSYSTEM.Used++
			case types.SlotSERVICE:
				slotUsed.SERVICE.Used++
			}
		}
		cpuUsed += m.GetFinal(engine.AttrCpuUsed, 0)
		powerUsed += m.GetFinal(engine.AttrPowerUsed, 0)
		if hasSlot && slot == types.SlotRIG {
			calibUsed += m.GetFinal(engine.AttrUpgradeCost, 0)
		}
		for _, eid := range m.EffectIDs.Values() {
			if isTurretEffect(eid) {
				turretHardpointsUsed += 1
				break
			}
			if isLauncherEffect(eid) {
				launcherHardpointsUsed += 1
				break
			}
		}
	}

	droneBayUsed, droneBwUsed := 0.0, 0.0
	fitDronesByID := omap.New[string, *types.FitDrone]()
	for i := range fit.Drones {
		fitDronesByID.Set(fit.Drones[i].ID, &fit.Drones[i])
	}
	for _, droneItem := range ctx.Drones {
		fd, ok := fitDronesByID.Get(droneItem.ID)
		if !ok {
			continue
		}
		t, ok := dataset.GetType(droneItem.TypeID)
		if !ok {
			continue
		}
		volume := 0.0
		if t.Volume != nil {
			volume = *t.Volume
		}
		bandwidth := 0.0
		if b, ok := droneItem.GetBase(engine.AttrDroneBandwidthUsed); ok {
			bandwidth = b
		}
		droneBayUsed += float64(volume * fd.CountTotal)
		droneBwUsed += float64(bandwidth * fd.CountActive)
	}

	shieldEhp := derived.ComputeLayerEhp(ship, "SHIELD", damageProfile)
	armorEhp := derived.ComputeLayerEhp(ship, "ARMOR", damageProfile)
	hullEhp := derived.ComputeLayerEhp(ship, "HULL", damageProfile)
	capStats := derived.ComputeCapacitor(ctx, dataset)
	tank := derived.ComputeTank(ctx)
	offense := derived.ComputeOffense(ctx, dataset, fit)
	sensorValue, sensorType := pickSensorStrength(ship)

	slotMax := func(a int) float64 { return ship.GetFinal(a, 0) }
	slotUsed.HI.Max = slotMax(engine.AttrHiSlots)
	slotUsed.MED.Max = slotMax(engine.AttrMedSlots)
	slotUsed.LO.Max = slotMax(engine.AttrLowSlots)
	slotUsed.RIG.Max = slotMax(engine.AttrRigSlots)
	slotUsed.SUBSYSTEM.Max = slotMax(engine.AttrSubsystemSlots)
	slotUsed.SERVICE.Max = slotMax(engine.AttrServiceSlots)

	activeDrones := 0.0
	for _, d := range fit.Drones {
		activeDrones = activeDrones + d.CountActive
	}

	return DerivedStats{
		Fitting: types.Fitting{
			CpuUsed: cpuUsed, CpuMax: ship.GetFinal(engine.AttrCpuOutput, 0),
			PowerUsed: powerUsed, PowerMax: ship.GetFinal(engine.AttrPowerOutput, 0),
			CalibrationUsed: calibUsed, CalibrationMax: ship.GetFinal(engine.AttrUpgradeCapacity, 0),
			DroneBandwidthUsed: droneBwUsed, DroneBandwidthMax: ship.GetFinal(engine.AttrDroneBandwidth, 0),
			DroneBayUsed: droneBayUsed, DroneBayMax: ship.GetFinal(engine.AttrDroneCapacity, 0),
			Slots: slotUsed,
			Hardpoints: types.Hardpoints{
				Turret:   types.SlotUse{Used: turretHardpointsUsed, Max: ship.GetFinal(engine.AttrTurretHardpoints, 0)},
				Launcher: types.SlotUse{Used: launcherHardpointsUsed, Max: ship.GetFinal(engine.AttrLauncherHardpoints, 0)},
			},
		},
		Defense: types.Defense{
			Shield: shieldEhp, Armor: armorEhp, Hull: hullEhp,
			EhpTotalAgainstProfile: derived.ComputeTotalEhp(shieldEhp, armorEhp, hullEhp, damageProfile != nil),
		},
		Offense:   offense,
		Capacitor: capStats,
		Tank:      tank,
		Navigation: types.Navigation{
			MaxVelocity: ship.GetFinal(engine.AttrMaxVelocity, 0), Mass: ship.GetFinal(engine.AttrMass, 0),
			Agility: ship.GetFinal(engine.AttrAgility, 0), AlignTimeSeconds: computeAlignTime(ship),
			WarpSpeed: ship.GetFinal(engine.AttrWarpSpeedMultiplier, 1),
		},
		Targeting: types.Targeting{
			MaxTargetingRange: jsmath.Min(ship.GetFinal(engine.AttrMaxTargetRange, 0), ship.GetFinal(engine.AttrMaxTargetRangeCap, 300000)),
			MaxLockedTargets: math.Ceil(jsmath.Min(
				charBaseMaxLockedTargets+ctx.Character.GetFinal(engine.AttrMaxLockedTargets, 0),
				ship.GetFinal(engine.AttrMaxLockedTargets, 0),
			)),
			SignatureRadius: ship.GetFinal(engine.AttrSignatureRadius, 0), ScanResolution: ship.GetFinal(engine.AttrScanResolution, 0),
			SensorStrength: sensorValue, SensorType: sensorType,
		},
		Drones: types.DroneStats{
			BayUsed: droneBayUsed, BayMax: ship.GetFinal(engine.AttrDroneCapacity, 0),
			BandwidthUsed: droneBwUsed, BandwidthMax: ship.GetFinal(engine.AttrDroneBandwidth, 0),
			Active: activeDrones, ControlRange: ship.GetFinal(engine.AttrDroneControlRange, 20000),
		},
		Projected:       projectionReportsOrEmpty(projectionReports),
		Structure:       derived.ComputeStructureMeta(ctx),
		ModuleSnapshots: snapshotModules(ctx),
	}
}

func projectionReportsOrEmpty(r []types.ProjectedEffectReport) []types.ProjectedEffectReport {
	if r == nil {
		return []types.ProjectedEffectReport{}
	}
	return r
}

// snapshotModules: valori finali per modulo e carica. Gli oggetti JS hanno le chiavi intere in ordine
// ascendente, poi le altre in ordine di inserimento: vale per gli id di attributo (sempre interi) e per
// gli id dei moduli (stringhe).
func snapshotModules(ctx *engine.FitContext) *omap.Map[string, ModuleAttrSnapshot] {
	snap := func(it *engine.ItemState) *omap.Map[int, float64] {
		ids := make([]int, 0, it.Attrs.Len())
		for i := 0; i < it.Attrs.Len(); i++ {
			id, _ := it.Attrs.At(i)
			ids = append(ids, id)
		}
		sort.Ints(ids)
		out := omap.New[int, float64]()
		for _, id := range ids {
			out.Set(id, it.GetFinal(id, 0))
		}
		return out
	}
	byID := omap.New[string, ModuleAttrSnapshot]()
	for _, m := range ctx.Modules {
		s := ModuleAttrSnapshot{Module: snap(m)}
		if m.Charge != nil {
			s.Charge = snap(m.Charge)
		}
		byID.Set(m.ID, s)
	}
	// Ordine di chiavi JS: indici di array ascendenti, poi le altre in ordine di inserimento.
	var intKeys []string
	var otherKeys []string
	for i := 0; i < byID.Len(); i++ {
		k, _ := byID.At(i)
		if n, err := strconv.ParseUint(k, 10, 32); err == nil && n <= 4294967294 && strconv.FormatUint(n, 10) == k {
			intKeys = append(intKeys, k)
		} else {
			otherKeys = append(otherKeys, k)
		}
	}
	sort.Slice(intKeys, func(i, j int) bool {
		a, _ := strconv.ParseUint(intKeys[i], 10, 32)
		b, _ := strconv.ParseUint(intKeys[j], 10, 32)
		return a < b
	})
	out := omap.New[string, ModuleAttrSnapshot]()
	for _, k := range append(intKeys, otherKeys...) {
		v, _ := byID.Get(k)
		out.Set(k, v)
	}
	return out
}

// computeAlignTime: t = ln(4) * massa * agilita' / 1e6.
func computeAlignTime(ship *engine.ItemState) float64 {
	mass := ship.GetFinal(engine.AttrMass, 0)
	agility := ship.GetFinal(engine.AttrAgility, 0)
	if mass == 0 || math.IsNaN(mass) || agility == 0 || math.IsNaN(agility) {
		return 0
	}
	return math.Log(4) * mass * agility / 1000000
}

// pickSensorStrength: il sensore piu' forte (a parita' "multispectral").
func pickSensorStrength(ship *engine.ItemState) (float64, string) {
	candidates := [4]struct {
		attr int
		typ  string
	}{
		{engine.AttrScanMagnetometricStrength, "magnetometric"},
		{engine.AttrScanLadarStrength, "ladar"},
		{engine.AttrScanRadarStrength, "radar"},
		{engine.AttrScanGravimetricStrength, "gravimetric"},
	}
	best := -1.0
	typ := "unknown"
	for _, c := range candidates {
		v := ship.GetFinal(c.attr, 0)
		if v > best {
			best = v
			typ = c.typ
		} else if v == best {
			typ = "multispectral"
		}
	}
	if best <= 0 {
		t := "unknown"
		if best == 0 {
			t = typ
		}
		return jsmath.Max(best, 0), t
	}
	return best, typ
}

var _ = fmt.Sprintf
