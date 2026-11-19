// Porting letterale di src/engine.ts, prima meta': costruzione degli item e applicazione di tutti i
// modificatori (fino ai report di proiezione). Le statistiche derivate e ComputeFit stanno in
// fitengine/compute.go (internal/derived importa questo package: niente cicli).
package engine

import (
	"fmt"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

// ComputeFitOptions e' ComputeFitOptions (nil / assente dove il TS ha `?`).
type ComputeFitOptions struct {
	SkillProfile              *types.SkillProfile
	DamageProfile             *types.DamageProfile
	TargetProfile             *types.TargetProfile
	Projected                 []types.ProjectedSource
	ProjectedSkillLevels      *omap.Map[int, float64]
	SpoolPercent              *float64
	DisintegratorSpoolPercent *float64
	SystemEffectTypeID        *int
	// BuildProjectionReport e' buildProjectionReport (effects/ewar.ts): iniettato da fitengine perche'
	// internal/effects dipende da internal/engine (niente cicli).
	BuildProjectionReport func(source *ItemState, ctx *FitContext) ProjectedEffectReport
}

type ProjectedEffectReport = types.ProjectedEffectReport

const (
	attrImplantness = 331
	attrBoosterness = 1087
)

// syntheticCharacterType e' SYNTHETIC_CHARACTER_TYPE: il personaggio non ha un type nell'SDE.
var syntheticCharacterType = types.SdeType{ID: -1, Name: "Character", GroupID: -1, CategoryID: -1,
	Attributes: []types.SdeTypeAttr{}, Effects: []types.SdeTypeEffect{}}

type slotItem struct {
	slot float64
	item *ItemState
}

// dropSlotCollisions tiene il PRIMO item di ogni slot e scarta gli altri.
func dropSlotCollisions(entries []slotItem) []*ItemState {
	taken := omap.NewSet[float64]()
	var out []*ItemState
	for _, e := range entries {
		if taken.Has(e.slot) {
			continue
		}
		taken.Add(e.slot)
		out = append(out, e.item)
	}
	return out
}

// FitState e' lo stato del fit dopo l'applicazione dei modificatori.
type FitState struct {
	Fit               *types.Fit
	Ctx               *FitContext
	Ship              *ItemState
	Character         *ItemState
	Mode              *ItemState
	Modules           []*ItemState
	Drones            []*ItemState
	Fighters          []*ItemState
	Implants          []*ItemState
	Boosters          []*ItemState
	Subsystems        []*ItemState
	ProjectionReports []ProjectedEffectReport
}

func hasCharge(id *int) bool { return id != nil && *id != 0 }

func projectedModule(ps *types.ProjectedSource, dataset *types.FittingDataset) (*ItemState, bool) {
	t, ok := dataset.GetType(ps.TypeID)
	if !ok {
		return nil, false
	}
	var chargeType *types.SdeType
	if hasCharge(ps.ChargeTypeID) {
		if ct, ok := dataset.GetType(*ps.ChargeTypeID); ok {
			chargeType = ct
		}
	}
	return MakeModuleState(&types.FitModule{
		ID: ps.ID, TypeID: ps.TypeID, SlotType: types.SlotHI, Position: 0, State: ps.State,
		ChargeTypeID: ps.ChargeTypeID, Mutator: ps.Mutator,
	}, t, chargeType), true
}

// ApplyModifiers e' la prima parte di computeFit: costruisce gli item e applica, nell'ordine del TS,
// skill, nave, subsystem, bonus del personaggio, moduli, droni, fighter, mode, gestori legacy e
// proiezioni.
func ApplyModifiers(fit *types.Fit, dataset *types.FittingDataset, opts *ComputeFitOptions) (*FitState, error) {
	shipType, ok := dataset.GetType(fit.ShipTypeID)
	if !ok {
		return nil, fmt.Errorf("computeFit: unknown ship typeID %d", fit.ShipTypeID)
	}

	ship := MakeShipState(shipType)
	character := MakeCharacterState(&syntheticCharacterType)

	var modules []*ItemState
	for i := range fit.Modules {
		fm := &fit.Modules[i]
		t, ok := dataset.GetType(fm.TypeID)
		if !ok {
			continue
		}
		var charge *types.SdeType
		if hasCharge(fm.ChargeTypeID) {
			if ct, ok := dataset.GetType(*fm.ChargeTypeID); ok {
				charge = ct
			}
		}
		modules = append(modules, MakeModuleState(fm, t, charge))
	}
	var drones []*ItemState
	for i := range fit.Drones {
		fd := &fit.Drones[i]
		if t, ok := dataset.GetType(fd.TypeID); ok {
			drones = append(drones, MakeDroneState(fd, t))
		}
	}
	var fighters []*ItemState
	for i := range fit.Fighters {
		ff := &fit.Fighters[i]
		if t, ok := dataset.GetType(ff.TypeID); ok {
			fighters = append(fighters, MakeFighterState(ff, t))
		}
	}
	// Due impianti nello stesso slot non possono coesistere: si tiene il primo.
	var implantEntries []slotItem
	for i := range fit.Implants {
		fi := &fit.Implants[i]
		if t, ok := dataset.GetType(fi.TypeID); ok {
			slot := fi.Slot
			if v, ok := skillAttr(t, attrImplantness); ok {
				slot = v
			}
			implantEntries = append(implantEntries, slotItem{slot, MakeImplantState(fi, t)})
		}
	}
	implants := dropSlotCollisions(implantEntries)
	var boosterEntries []slotItem
	for i := range fit.Boosters {
		fb := &fit.Boosters[i]
		if t, ok := dataset.GetType(fb.TypeID); ok {
			slot := fb.Slot
			if v, ok := skillAttr(t, attrBoosterness); ok {
				slot = v
			}
			boosterEntries = append(boosterEntries, slotItem{slot, MakeBoosterState(fb, t)})
		}
	}
	boosters := dropSlotCollisions(boosterEntries)
	var subsystems []*ItemState
	for i := range fit.Subsystems {
		fs := &fit.Subsystems[i]
		if t, ok := dataset.GetType(fs.TypeID); ok {
			subsystems = append(subsystems, MakeSubsystemState(fs, t))
		}
	}

	// Modo T3D/T3C: quello del fit, altrimenti quello di default per la nave.
	var mode *ItemState
	var modeTypeID int
	hasMode := false
	if fit.ModeTypeID != nil {
		modeTypeID, hasMode = *fit.ModeTypeID, true
	} else if id, ok := DefaultModeTypeIDFor(shipType, dataset); ok {
		modeTypeID, hasMode = id, true
	}
	if hasMode {
		if t, ok := dataset.GetType(modeTypeID); ok {
			mode = MakeModeState(modeTypeID, t)
		}
	}

	// Il range di controllo droni parte dal default SDE (20 km se assente).
	droneControlDefault := 20000.0
	if a, ok := dataset.Attributes.Get(AttrDroneControlRange); ok {
		droneControlDefault = a.DefaultValue
	}
	ship.Attr(AttrDroneControlRange, droneControlDefault)

	setSchema := func(it *ItemState) {
		if it == nil {
			return
		}
		it.AttrSchema = &dataset.Attributes
		if it.Charge != nil {
			it.Charge.AttrSchema = &dataset.Attributes
		}
	}
	setSchema(ship)
	setSchema(character)
	setSchema(mode)
	for _, group := range [][]*ItemState{modules, drones, fighters, implants, boosters, subsystems} {
		for _, it := range group {
			setSchema(it)
		}
	}

	// Skill: Object.entries di un oggetto a chiavi intere -> id ascendente (gia' cosi' in SkillProfile).
	skillLevels := omap.New[int, float64]()
	if opts.SkillProfile != nil && opts.SkillProfile.Skills != nil {
		for i := 0; i < opts.SkillProfile.Skills.Len(); i++ {
			k, v := opts.SkillProfile.Skills.At(i)
			skillLevels.Set(k, float64(v))
		}
	}

	disintegratorSpool := 1.0
	if opts.DisintegratorSpoolPercent != nil {
		disintegratorSpool = *opts.DisintegratorSpoolPercent
	}
	ctx := NewFitContext(FitContextInit{
		Ship: ship, Character: character, SkillLevels: skillLevels, Modules: modules, Drones: drones,
		Fighters: fighters, Implants: implants, Boosters: boosters, Subsystems: subsystems, Mode: mode,
		SkillProfile: opts.SkillProfile, Dataset: dataset, DisintegratorSpoolPercent: disintegratorSpool,
	})

	// 1. Skill, nave e subsystem.
	ApplySkills(ctx, dataset)
	ApplySourceItem(ship, ctx, dataset, PassAll)
	for _, s := range subsystems {
		ApplySourceItem(s, ctx, dataset, PassAll)
	}
	ApplyLegacySubsystemSlots(ctx)
	ApplyLegacySubsystemAddPassive(ctx)

	// 2. Pre-passata EffectStopper dalle sorgenti proiettate.
	if len(opts.Projected) > 0 {
		var earlyProjected []*ItemState
		for i := range opts.Projected {
			if it, ok := projectedModule(&opts.Projected[i], dataset); ok {
				earlyProjected = append(earlyProjected, it)
			}
		}
		ctx.StoppedLocalEffectIDs = CollectEffectStoppers(earlyProjected, dataset)
	}

	// 3. Lato personaggio: prima i bonus fra pari, poi quelli in uscita.
	var charSide []*ItemState
	charSide = append(charSide, implants...)
	charSide = append(charSide, boosters...)
	for _, it := range charSide {
		ApplySourceItem(it, ctx, dataset, PassPeer)
	}
	for _, it := range charSide {
		ApplySourceItem(it, ctx, dataset, PassOutgoing)
	}

	// 4. Effetti a runTime earliest, cariche prima dei moduli, poi moduli, droni, fighter, mode.
	for _, m := range modules {
		ApplySourceItem(m, ctx, dataset, PassEarliest)
	}
	for _, m := range modules {
		if m.Charge != nil {
			ApplySourceItem(m.Charge, ctx, dataset, PassAll)
		}
	}
	for _, m := range modules {
		ApplySourceItem(m, ctx, dataset, PassAll)
	}
	for _, d := range drones {
		ApplySourceItem(d, ctx, dataset, PassAll)
	}
	for _, f := range fighters {
		ApplySourceItem(f, ctx, dataset, PassAll)
	}
	if mode != nil {
		ApplySourceItem(mode, ctx, dataset, PassAll)
	}

	// 5. Gestori legacy.
	spool := 1.0
	if opts.SpoolPercent != nil {
		spool = *opts.SpoolPercent
	}
	ApplyLegacyOverload(ctx)
	ApplyLegacyPropMods(ctx)
	ApplyLegacyEntosisLink(ctx)
	ApplyLegacyCapitalEhe(ctx)
	ApplyLegacyHicBubble(ctx)
	ApplyLegacyMjdSigBloom(ctx)
	ApplyLegacyCommandBursts(ctx)
	ApplyLegacyCapBoosterInjection(ctx)
	ApplyLegacyDoomsdaySelfEffects(ctx)
	ApplyLegacyBastion(ctx)
	ApplyLegacyMutadaptiveSpool(ctx, spool)
	ApplyLegacyDisintegratorSpool(ctx, disintegratorSpool)
	ApplyLegacyFighterAbilities(ctx)
	ApplyLegacySystemEffect(ctx, opts.SystemEffectTypeID)
	ApplyLegacyRAH(ctx, opts.DamageProfile)

	// 6. Sorgenti proiettate sul fit.
	var projectionReports []ProjectedEffectReport
	offensiveModifiersDisallowed := ctx.Ship.GetFinal(872, 0) > 0
	if len(opts.Projected) > 0 {
		previousTarget := ctx.Target
		ctx.Target = ctx.Ship
		for i := range opts.Projected {
			ps := &opts.Projected[i]
			projectedItem, ok := projectedModule(ps, dataset)
			if !ok {
				continue
			}
			ctx.ProjectedSources = append(ctx.ProjectedSources, projectedItem)
			if projectedItem.State == types.StateActive || projectedItem.State == types.StateOverload {
				isFriendly := projectedItem.EffectIDs.Has(6427) || projectedItem.EffectIDs.Has(6428)
				if !isFriendly && offensiveModifiersDisallowed {
					continue
				}
				ApplySourceItem(projectedItem, ctx, dataset, PassAll)
				if projectedItem.Charge != nil {
					ApplySourceItem(projectedItem.Charge, ctx, dataset, PassAll)
				}
				ApplyLegacyProjectionEwar(projectedItem, ps.ProjectionRange, ctx)
				for _, fr := range ApplyLegacyFighterProjection(projectedItem, ctx) {
					ps := fr.PerSecond
					projectionReports = append(projectionReports, ProjectedEffectReport{
						TypeID: fr.TypeID, Kind: fr.Kind, PerSecond: &ps, Summary: fr.Summary,
					})
				}
				if rep := BuildRemoteRepReport(projectedItem, ps.ProjectionRange, ctx); rep != nil {
					kind := "REMOTE_REP_HULL"
					if rep.Layer == "SHIELD" {
						kind = "REMOTE_REP_SHIELD"
					} else if rep.Layer == "ARMOR" {
						kind = "REMOTE_REP_ARMOR"
					}
					per := rep.PerSecond
					projectionReports = append(projectionReports, ProjectedEffectReport{
						TypeID: projectedItem.TypeID, Kind: kind, PerSecond: &per,
						Summary: rep.Layer + " rep incoming: " + jsmath.ToFixed(rep.PerSecond, 1) + " HP/s",
					})
					continue
				}
				if cap := BuildCapWarfareReport(projectedItem, ps.ProjectionRange, ctx); cap != nil {
					sign := "injection"
					if cap.PerSecond >= 0 {
						sign = "drain"
					}
					per := cap.PerSecond
					abs := cap.PerSecond
					if abs < 0 {
						abs = -abs
					}
					projectionReports = append(projectionReports, ProjectedEffectReport{
						TypeID: projectedItem.TypeID, Kind: cap.Kind, PerSecond: &per,
						Summary: "Cap " + sign + ": " + jsmath.ToFixed(abs, 1) + " GJ/s",
					})
					continue
				}
				if jam, ok := BuildEcmProjectionReport(projectedItem, ps.ProjectionRange, ctx); ok {
					j := jam
					projectionReports = append(projectionReports, ProjectedEffectReport{
						TypeID: projectedItem.TypeID, Kind: "ECM", JamChance: &j,
						Summary: "Per-cycle jam chance " + jsmath.ToFixed(jam*100, 1) + "%",
					})
					continue
				}
				if opts.BuildProjectionReport != nil {
					projectionReports = append(projectionReports, opts.BuildProjectionReport(projectedItem, ctx))
				}
			}
		}
		ctx.Target = previousTarget
	}

	return &FitState{
		Fit: fit, Ctx: ctx, Ship: ship, Character: character, Mode: mode, Modules: modules, Drones: drones,
		Fighters: fighters, Implants: implants, Boosters: boosters, Subsystems: subsystems,
		ProjectionReports: projectionReports,
	}, nil
}

// SnapshotMap e' snapshotMap: attributi di un item con il valore calcolato SENZA tetti (ma.compute()).
func SnapshotMap(item *ItemState) *omap.Map[int, ComputedAttribute] {
	out := omap.New[int, ComputedAttribute]()
	for i := 0; i < item.Attrs.Len(); i++ {
		id, ma := item.Attrs.At(i)
		final := ma.Compute(nil)
		out.Set(id, ComputedAttribute{ID: id, Base: ma.Base, Final: final, Afflictions: ma.Afflictions})
	}
	return out
}
