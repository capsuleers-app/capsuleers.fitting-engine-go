// Porting letterale di src/modifierEngine.ts: dispatcher generico `modifierInfo`, scaling delle
// skill, gruppi di stacking, applySkills, gestori legacy (applyLegacy*), proiezioni, command buff.
// Le tabelle numeriche di primo livello stanno in modifier_engine_tables.go (generato una volta
// dal TS; da allora ogni modifica va fatta in entrambi i repo).
package engine

import (
	"fmt"
	"math"
	"sort"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/gen"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

var (
	_ = fmt.Sprintf
	_ = math.Abs
	_ = sort.Ints
	_ = jsmath.Round
	_ = omap.New[int, int]
)

// noPenaltyKind: tipi di sorgente i cui modificatori NON subiscono penalita' di stacking, qualunque
// cosa dica l'effetto (NO_PENALTY_KINDS: la regola di contesto di pyfa e nient'altro).
func noPenaltyKind(k ItemKind) bool {
	return k == "skill" || k == KindCharacter || k == KindImplant || k == KindBooster
}

// percentUnitIDs: unita' dogma che marcano l'attributo come "percentuale" (PERCENT_UNIT_IDS).
var percentUnitIDs = newIntSet([]int{105, 109, 111, 121, 124, 127})

// scaleForPipeline: PostPercent e' sempre un numero percentuale grezzo (5 -> +5%), il resto passa invariato.
func scaleForPipeline(rawValue float64, _unitID *int, op types.ModifierOperation) float64 {
	if op == types.OpPostPercent {
		return rawValue / 100
	}
	return rawValue
}

// ApplyPass: quale meta' di un'applicazione a due passaggi sta facendo la chiamata.
type ApplyPass string

const (
	PassAll      ApplyPass = "all"
	PassPeer     ApplyPass = "peer"
	PassOutgoing ApplyPass = "outgoing"
	PassEarliest ApplyPass = "earliest"
)

// isPeerModifier: i target del modificatore includono un impianto o un booster?
func isPeerModifier(source *ItemState, mi *types.SdeModifierInfo, ctx *FitContext, effectID int) bool {
	eid := effectID
	for _, t := range ctx.TargetsForModifier(mi, source, &eid) {
		if t.Kind == KindImplant || t.Kind == KindBooster {
			return true
		}
	}
	return false
}

type effectMod struct {
	effect *types.SdeEffect
	mi     *types.SdeModifierInfo
}

// ApplySourceItem applica ogni modificatore di ogni effetto di un item sorgente, filtrando per lo
// stato corrente. Prima i modificatori SELF (domain itemID), poi gli OUTGOING.
func ApplySourceItem(source *ItemState, ctx *FitContext, dataset *types.FittingDataset, pass ApplyPass) {
	isLocalModule := source.Kind == KindModule
	var selfMods, outMods []effectMod
	for _, eid := range source.EffectIDs.Values() {
		// Effetti rivendicati da gestori hardcoded: non vanno riapplicati qui.
		if LegacyHandledEffectIds.Has(eid) {
			continue
		}
		// `earliest` applica SOLO gli effetti a runTime earliest; gli altri passaggi li lasciano a lui.
		if (pass == PassEarliest) != RuntimeEarliestEffectIds.Has(eid) {
			continue
		}
		// Bonus scalati sul security status del pilota: il pilota di default ha sec status 0 -> no-op.
		if SecStatusScaledEffectIds.Has(eid) {
			continue
		}
		// Parita' pyfa: uno scram/disruptor proiettato ferma gli effetti 6441 e 6442 sul bersaglio.
		if isLocalModule && ctx.StoppedLocalEffectIDs.Has(eid) {
			continue
		}
		// I drawback di un booster sono opt-in (default spenti).
		if source.Kind == KindBooster && IsBoosterSideEffectID(eid) && !source.ActiveSideEffects.Has(eid) {
			continue
		}
		effect, ok := dataset.Effects.Get(eid)
		if !ok {
			continue
		}
		if !source.AppliesAtState(effect) {
			continue
		}
		for i := range effect.ModifierInfo {
			mi := &effect.ModifierInfo[i]
			if pass == PassPeer || pass == PassOutgoing {
				peer := isPeerModifier(source, mi, ctx, effect.ID)
				if pass == PassPeer {
					if !peer {
						continue
					}
				} else if peer {
					continue
				}
			}
			if mi.Domain == "itemID" {
				selfMods = append(selfMods, effectMod{effect, mi})
			} else {
				outMods = append(outMods, effectMod{effect, mi})
			}
		}
	}
	for _, e := range selfMods {
		ApplyOneModifier(source, e.effect, e.mi, ctx, dataset)
	}
	for _, e := range outMods {
		ApplyOneModifier(source, e.effect, e.mi, ctx, dataset)
	}
}

// CollectEffectStoppers: unione degli effectID soppressi da un EffectStopper (domain 'target')
// di una sorgente proiettata ACTIVE/OVERLOAD, piu' le abilita' dei fighter 6436/6464.
func CollectEffectStoppers(projectedSources []*ItemState, dataset *types.FittingDataset) *omap.Set[int] {
	out := omap.NewSet[int]()
	for _, src := range projectedSources {
		if src.State != types.StateActive && src.State != types.StateOverload {
			continue
		}
		for _, eid := range src.EffectIDs.Values() {
			effect, ok := dataset.Effects.Get(eid)
			if !ok {
				continue
			}
			for i := range effect.ModifierInfo {
				mi := &effect.ModifierInfo[i]
				if mi.Func != "EffectStopper" {
					continue
				}
				if mi.Domain != "target" {
					continue
				}
				if mi.EffectID == nil {
					continue
				}
				out.Add(*mi.EffectID)
			}
			if eid == 6436 || eid == 6464 {
				out.Add(6441)
				out.Add(6442)
			}
		}
	}
	return out
}

// FighterProjectionReport: rapporto di un'abilita' di drain dei fighter.
type FighterProjectionReport struct {
	TypeID    int
	Kind      string // "NEUT"
	PerSecond float64
	Summary   string
}

// ApplyLegacyFighterProjection applica le abilita' di proiezione di un fighter ostile (web, neut).
func ApplyLegacyFighterProjection(fighter *ItemState, ctx *FitContext) []FighterProjectionReport {
	var reports []FighterProjectionReport
	if fighter.State != types.StateActive && fighter.State != types.StateOverload {
		return reports
	}

	if fighter.EffectIDs.Has(6435) {
		pen := fighter.GetFinal(attrFighterWebSpeedPen, 0)
		if pen != 0 {
			g := fmt.Sprintf("default:%d", attrMaxVel)
			ctx.Ship.AddAffliction(attrMaxVel, types.ModifierAffliction{
				SourceKind: "fighter", SourceID: fighter.ID, Operation: types.OpPostPercent,
				Value: pen / 100, StackingGroup: &g,
			}, 0)
		}
	}

	if fighter.EffectIDs.Has(6434) {
		amount := fighter.GetFinal(attrFighterNeutAmount, 0)
		durMs := fighter.GetFinal(attrFighterNeutDuration, 0)
		if amount > 0 && durMs > 0 {
			perSecond := amount / (durMs / 1000)
			reports = append(reports, FighterProjectionReport{
				TypeID: fighter.TypeID, Kind: "NEUT", PerSecond: perSecond,
				Summary: "Fighter cap drain: " + jsmath.ToFixed(perSecond, 1) + " GJ/s",
			})
		}
	}
	return reports
}

// modifierValue e' il risultato di computeModifierValue; scaled=true se il valore e' passato dallo
// scaling per livello di skill.
type modifierValue struct {
	value  float64
	scaled bool
}

// ApplyOneModifier applica un singolo descrittore di modificatore.
func ApplyOneModifier(source *ItemState, effect *types.SdeEffect, mi *types.SdeModifierInfo, ctx *FitContext, dataset *types.FittingDataset) {
	if mi.Func == "EffectStopper" {
		return // gestito altrove
	}

	op := types.OpPostMul
	if mi.Operation != nil {
		o, ok := OperationBySdeCode(*mi.Operation)
		if !ok {
			return // opcode sconosciuto: salta
		}
		op = o
	}

	if mi.ModifiedAttributeID == nil {
		return
	}

	computed := computeModifierValue(source, mi, ctx, dataset, op)
	if computed == nil {
		return
	}

	eid := effect.ID
	targets := ctx.TargetsForModifier(mi, source, &eid)
	if len(targets) == 0 {
		return
	}

	// Le op moltiplicative portano un moltiplicatore letterale; solo i modificatori scalati per skill
	// sono bonus frazionari da leggere come 1 + bonus.
	isMul := op == types.OpPreMul || op == types.OpPostMul || op == types.OpPreDiv || op == types.OpPostDiv
	value := computed.value
	if isMul && computed.scaled {
		value = 1 + computed.value
	}

	// Per le op moltiplicative la base dell'attributo bersaglio parte dal defaultValue SDE.
	sdeDefault := 0.0
	if isMul {
		if a, ok := dataset.Attributes.Get(*mi.ModifiedAttributeID); ok {
			sdeDefault = a.DefaultValue
		}
	}

	// Il gruppo di stacking si decide PER BERSAGLIO.
	for _, target := range targets {
		affliction := types.ModifierAffliction{
			SourceKind:            mapSourceKind(source.Kind),
			SourceID:              source.ID,
			Operation:             op,
			Value:                 value,
			StackingGroup:         computeStackingGroup(source, mi, dataset, effect, target, ctx),
			ResistanceAttributeID: effect.ResistanceAttributeID,
		}
		target.AddAffliction(*mi.ModifiedAttributeID, affliction, sdeDefault)
	}
}

// computeModifierValue: il valore numerico che il modificatore contribuisce (nil = non applicabile).
func computeModifierValue(source *ItemState, mi *types.SdeModifierInfo, ctx *FitContext, dataset *types.FittingDataset, op types.ModifierOperation) *modifierValue {
	if mi.ModifyingAttributeID == nil {
		return nil
	}
	mattr := *mi.ModifyingAttributeID
	if !source.HasAttr(mattr) {
		return nil
	}
	rawBase := source.GetFinal(mattr, 0)
	var unitID *int
	if sourceAttr, ok := dataset.Attributes.Get(mattr); ok {
		unitID = sourceAttr.UnitID
	}
	baseValue := scaleForPipeline(rawBase, unitID, op)

	// Subsystem T3C: scalano per la sotto-skill razziale corrispondente.
	if source.Kind == KindSubsystem {
		if scalingSkill, ok := subsystemBonusScalingSkill.get(mattr); ok {
			level := ctx.SkillLevel(scalingSkill)
			if level == 0 {
				return nil
			}
			return &modifierValue{value: baseValue * level, scaled: true}
		}
		if mi.Func == "LocationRequiredSkillModifier" || mi.Func == "OwnerRequiredSkillModifier" {
			if mi.SkillTypeID == nil {
				return nil
			}
			if ctx.SkillLevel(*mi.SkillTypeID) == 0 {
				return nil
			}
		}
		return &modifierValue{value: baseValue, scaled: false}
	}

	// Moduli/cariche/droni/fighter/mode: il modificatore e' guardato dalla skill richiesta ma non scalato.
	if (source.Kind == KindModule || source.Kind == KindCharge || source.Kind == KindDrone ||
		source.Kind == KindFighter || source.Kind == KindMode) &&
		(mi.Func == "LocationRequiredSkillModifier" || mi.Func == "OwnerRequiredSkillModifier") {
		if mi.SkillTypeID == nil {
			return nil
		}
		level := ctx.SkillLevel(*mi.SkillTypeID)
		if level == 0 {
			return nil
		}
		return &modifierValue{value: baseValue, scaled: false}
	}

	// Bonus di scafo "moltiplicato per il livello della skill".
	if source.Kind == KindShip {
		if scalingSkill, ok := shipBonusScalingSkill.get(mattr); ok {
			level := ctx.SkillLevel(scalingSkill)
			if level == 0 {
				return nil
			}
			return &modifierValue{value: baseValue * level, scaled: true}
		}
	}

	if mi.Func == "LocationRequiredSkillModifier" || mi.Func == "OwnerRequiredSkillModifier" {
		if mi.SkillTypeID == nil {
			return nil
		}
		skillID := *mi.SkillTypeID
		level := ctx.SkillLevel(skillID)
		if level == 0 {
			return nil // skill a 0: modificatore inattivo
		}
		if source.Kind == KindShip {
			return &modifierValue{value: baseValue, scaled: false}
		}
		if source.Kind == KindImplant || source.Kind == KindBooster {
			return &modifierValue{value: baseValue, scaled: false}
		}
		if ItemRequiresSkill(source, skillID) {
			return &modifierValue{value: baseValue, scaled: false}
		}
		return &modifierValue{value: baseValue * level, scaled: true}
	}

	return &modifierValue{value: baseValue, scaled: false}
}

// penaltyReceiver: la classe di ricevitore di un bersaglio, per le chiavi qualificate del registro.
func penaltyReceiver(target, source *ItemState, ctx *FitContext) gen.PenaltyReceiver {
	if target == ctx.Ship {
		return gen.ReceiverShip
	}
	if target == source {
		return gen.ReceiverSelf
	}
	switch target.Kind {
	case KindCharge:
		return gen.ReceiverCharge
	case KindDrone:
		return gen.ReceiverDrone
	case KindFighter:
		return gen.ReceiverFighter
	case KindCharacter:
		return gen.ReceiverCharacter
	case KindModule, KindSubsystem:
		return gen.ReceiverModule
	}
	return gen.ReceiverSelf
}

// honouredPenaltyGroup: HONOURED_PENALTY_GROUPS.
func honouredPenaltyGroup(g string) bool {
	switch g {
	case "postMul", "postDiv", "postPerc", "postPercent", "cloakingScanResolutionMultiplier":
		return true
	}
	return false
}

// computeStackingGroup: chiave di penalita' di stacking (nil = nessuna) per (sorgente, effetto, bersaglio).
func computeStackingGroup(source *ItemState, mi *types.SdeModifierInfo, dataset *types.FittingDataset, effect *types.SdeEffect, target *ItemState, ctx *FitContext) *string {
	if noPenaltyKind(source.Kind) {
		return nil
	}
	group, ok := gen.StackPenaltyGroup(effect.ID, mi.ModifiedAttributeID, penaltyReceiver(target, source, ctx), string(source.Kind))
	if !ok {
		return nil
	}
	var s string
	if !honouredPenaltyGroup(group) {
		s = fmt.Sprintf("default:%s", jsAttrID(mi.ModifiedAttributeID))
	} else {
		s = fmt.Sprintf("%s:%s", group, jsAttrID(mi.ModifiedAttributeID))
	}
	return &s
}

// jsAttrID formatta `${mi.modifiedAttributeID}` di JS: "undefined" se assente.
func jsAttrID(p *int) string {
	if p == nil {
		return "undefined"
	}
	return fmt.Sprintf("%d", *p)
}

// mapSourceKind: tipo di sorgente dell'affliction per il breakdown dell'UI.
func mapSourceKind(kind ItemKind) string {
	switch kind {
	case KindShip:
		return "ship"
	case KindModule:
		return "module"
	case KindCharge:
		return "module" // gli effetti di una carica sono attribuiti al modulo padre
	case KindDrone:
		return "drone"
	case KindFighter:
		return "drone" // fighter ~ drone per l'UI
	case KindImplant:
		return "implant"
	case KindBooster:
		return "booster"
	case KindSubsystem:
		return "subsystem"
	case KindMode:
		return "mode"
	case KindCharacter:
		return "skill"
	}
	return ""
}

// ApplySkills applica gli effetti di ogni skill addestrata (livello > 0) ai bersagli del personaggio.
func ApplySkills(ctx *FitContext, dataset *types.FittingDataset) {
	skills := &dataset.TypesByBucket[types.BucketSkills]
	for i := 0; i < ctx.SkillLevels.Len(); i++ {
		skillID, level := ctx.SkillLevels.At(i)
		if level <= 0 {
			continue
		}
		skillType, ok := skills.Get(skillID)
		if !ok {
			continue
		}
		for _, eRef := range skillType.Effects {
			if legacyDroneDmgEffectIds.Has(eRef.ID) {
				applyLegacyDroneDmgBonus(skillType, skillID, level, ctx)
				continue
			}
			if missileChargeDmgAttr, ok := legacyMissileChargeDmgEffects.get(eRef.ID); ok {
				applyLegacyMissileChargeDmg(skillType, skillID, level, ctx, missileChargeDmgAttr)
				continue
			}
			if eRef.ID == 1851 {
				applyLegacyMissileSpecRof(skillType, skillID, level, ctx)
				continue
			}
			effect, ok := dataset.Effects.Get(eRef.ID)
			if !ok {
				continue
			}
			for mi_i := range effect.ModifierInfo {
				mi := &effect.ModifierInfo[mi_i]
				if mi.Func == "EffectStopper" {
					continue
				}
				if mi.ModifiedAttributeID == nil {
					continue
				}
				if mi.ModifyingAttributeID == nil {
					continue
				}
				op := types.OpPostMul
				if mi.Operation != nil {
					o, ok := OperationBySdeCode(*mi.Operation)
					if !ok {
						continue
					}
					op = o
				}
				isMul := op == types.OpPreMul || op == types.OpPostMul || op == types.OpPreDiv || op == types.OpPostDiv
				var rawBase float64
				found := false
				for _, a := range skillType.Attributes {
					if a.ID == *mi.ModifyingAttributeID {
						rawBase, found = a.V, true
						break
					}
				}
				if !found {
					continue
				}
				var unitID *int
				if m, ok := dataset.Attributes.Get(*mi.ModifyingAttributeID); ok {
					unitID = m.UnitID
				}
				baseVal := scaleForPipeline(rawBase, unitID, op)
				var value float64
				if isMul {
					value = 1 + float64(baseVal*level)
				} else {
					value = baseVal * level
				}
				var stackingGroup *string
				if !noPenaltyKind(KindCharacter) {
					s := fmt.Sprintf("default:%d", *mi.ModifiedAttributeID)
					stackingGroup = &s
				}
				targets := resolveSkillTargets(mi, ctx)
				if len(targets) == 0 {
					continue
				}
				affliction := types.ModifierAffliction{
					SourceKind: "skill", SourceID: fmt.Sprintf("skill:%d", skillID),
					Operation: op, Value: value, StackingGroup: stackingGroup,
				}
				for _, t := range targets {
					t.AddAffliction(*mi.ModifiedAttributeID, affliction, 0)
				}
			}
		}
	}
}

// orElse e' `x || d` di JS su numeri: d se x e' 0, -0 o NaN.
func orElse(x, d float64) float64 {
	if x == 0 || math.IsNaN(x) {
		return d
	}
	return x
}

// skillAttr e' `skillType.attributes.find(a => a.id === id)?.v`.
func skillAttr(t *types.SdeType, id int) (float64, bool) {
	for _, a := range t.Attributes {
		if a.ID == id {
			return a.V, true
		}
	}
	return 0, false
}

func skillAfflictionNoPenalty(skillID int, value float64) types.ModifierAffliction {
	return types.ModifierAffliction{
		SourceKind: "skill", SourceID: fmt.Sprintf("skill:%d", skillID),
		Operation: types.OpPostPercent, Value: value, StackingGroup: nil,
	}
}

// applyLegacyDroneDmgBonus: bonus di danno dei droni dalle skill (effetto 1730).
func applyLegacyDroneDmgBonus(skillType *types.SdeType, skillID int, level float64, ctx *FitContext) {
	bonusRaw, ok := skillAttr(skillType, attrDamageMultiplierBonus)
	if !ok || bonusRaw == 0 {
		return
	}
	value := (bonusRaw / 100) * level // frazione PostPercent
	apply := func(target *ItemState) {
		target.AddAffliction(attrDamageMultiplier, skillAfflictionNoPenalty(skillID, value), 0)
	}
	for _, d := range ctx.Drones {
		if ItemRequiresSkill(d, skillID) {
			apply(d)
		}
	}
	for _, f := range ctx.Fighters {
		if ItemRequiresSkill(f, skillID) {
			apply(f)
		}
	}
}

// applyLegacyMissileChargeDmg: bonus di danno per tipo sulle cariche missile.
func applyLegacyMissileChargeDmg(skillType *types.SdeType, skillID int, level float64, ctx *FitContext, targetAttr int) {
	bonusRaw, ok := skillAttr(skillType, attrDamageMultiplierBonus)
	if !ok || bonusRaw == 0 {
		return
	}
	value := (bonusRaw / 100) * level
	for _, m := range ctx.Modules {
		if m.Charge == nil {
			continue
		}
		if !ItemRequiresSkill(m.Charge, skillID) {
			continue
		}
		m.Charge.AddAffliction(targetAttr, skillAfflictionNoPenalty(skillID, value), 0)
	}
}

// applyLegacyMissileSpecRof: bonus ROF della Specialization missili T2 (effetto 1851).
func applyLegacyMissileSpecRof(skillType *types.SdeType, skillID int, level float64, ctx *FitContext) {
	bonusRaw, ok := skillAttr(skillType, attrRofBonus)
	if !ok || bonusRaw == 0 {
		return
	}
	value := (bonusRaw / 100) * level
	for _, m := range ctx.Modules {
		if !ItemRequiresSkill(m, skillID) {
			continue
		}
		m.AddAffliction(51 /* ATTR_SPEED */, skillAfflictionNoPenalty(skillID, value), 0)
	}
}

// ApplyLegacyPropMods: AB/MWD (effetti 6730/6731) come boost di velocita' + massa + signature bloom.
func ApplyLegacyPropMods(ctx *FitContext) {
	for _, mod := range ctx.Modules {
		if mod.State != types.StateActive && mod.State != types.StateOverload {
			continue
		}
		isPropMod := false
		for _, eid := range legacyPropmodEffectIds.Values() {
			if mod.EffectIDs.Has(eid) {
				isPropMod = true
				break
			}
		}
		if !isPropMod {
			continue
		}
		speedFactor := mod.GetFinal(attrSpeedFactor, 0)
		thrust := mod.GetFinal(attrSpeedBoostFactor, 0)
		massAdd := mod.GetFinal(attrMassAddition, 0)
		if speedFactor <= 0 || thrust <= 0 {
			continue
		}
		ctx.Ship.AddAffliction(attrMass, types.ModifierAffliction{
			SourceKind: "module", SourceID: mod.ID, Operation: types.OpModAdd, Value: massAdd, StackingGroup: nil,
		}, 0)
		effectiveMass := orElse(ctx.Ship.GetFinal(attrMass, 1), 1)
		boost := (speedFactor / 100) * thrust / effectiveMass
		ctx.Ship.AddAffliction(attrMaxVelocity, types.ModifierAffliction{
			SourceKind: "module", SourceID: mod.ID, Operation: types.OpPostPercent, Value: boost, StackingGroup: nil,
		}, 0)
		sigBonus := mod.GetFinal(attrSignatureRadiusBonus, 0)
		if sigBonus > 0 {
			g := fmt.Sprintf("default:%d", attrSignatureRadius)
			ctx.Ship.AddAffliction(attrSignatureRadius, types.ModifierAffliction{
				SourceKind: "module", SourceID: mod.ID, Operation: types.OpPostPercent, Value: sigBonus / 100, StackingGroup: &g,
			}, 0)
		}
	}
}

// aff costruisce una ModifierAffliction.
func aff(kind, id string, op types.ModifierOperation, value float64, group *string) types.ModifierAffliction {
	return types.ModifierAffliction{SourceKind: kind, SourceID: id, Operation: op, Value: value, StackingGroup: group}
}

// defGroup e' la chiave `default:${attr}`.
func defGroup(attr int) *string {
	s := fmt.Sprintf("default:%d", attr)
	return &s
}

// getBaseOr e' `item.getBase(id) ?? d`.
func getBaseOr(item *ItemState, id int, d float64) float64 {
	if v, ok := item.GetBase(id); ok {
		return v
	}
	return d
}

func activeOrOverload(m *ItemState) bool {
	return m.State == types.StateActive || m.State == types.StateOverload
}

// ApplyLegacyEntosisLink: Entosis Link (effetto 6063) boosta i sensori della nave.
func ApplyLegacyEntosisLink(ctx *FitContext) {
	for _, mod := range ctx.Modules {
		if !activeOrOverload(mod) {
			continue
		}
		if !mod.EffectIDs.Has(entosisLinkEffectId) {
			continue
		}
		sensorPairs := [4][2]int{
			{1027, attrScanGravimetric},
			{1028, attrScanLadar},
			{1029, attrScanMagnetometric},
			{1030, attrScanRadar},
		}
		for _, p := range sensorPairs {
			srcAttr, tgtAttr := p[0], p[1]
			bonusPct := getBaseOr(mod, srcAttr, 0)
			if bonusPct == 0 {
				continue
			}
			ctx.Ship.AddAffliction(tgtAttr, aff("module", mod.ID, types.OpPostPercent, bonusPct/100, nil), 0)
		}
	}
}

// ApplyLegacyCapitalEhe: Capital Emergency Hull Energizer (effetto 6484).
func ApplyLegacyCapitalEhe(ctx *FitContext) {
	for _, mod := range ctx.Modules {
		if !activeOrOverload(mod) {
			continue
		}
		if !mod.EffectIDs.Has(capitalEheEffectId) {
			continue
		}
		pairs := [4][2]int{
			{attrHullEheEmRes, attrHullEmRes},
			{attrHullEheThermRes, attrHullThermRes},
			{attrHullEheKinRes, attrHullKinRes},
			{attrHullEheExpRes, attrHullExpRes},
		}
		for _, p := range pairs {
			srcAttr, tgtAttr := p[0], p[1]
			factor := getBaseOr(mod, srcAttr, 1)
			if factor == 1 {
				continue
			}
			ctx.Ship.AddAffliction(tgtAttr, aff("module", mod.ID, types.OpPostMul, factor, defGroup(tgtAttr)), 0)
		}
	}
}

// ApplyLegacyHicBubble: sfera di interdizione dei Heavy Interdiction Cruiser (effetto 3380).
func ApplyLegacyHicBubble(ctx *FitContext) {
	for _, mod := range ctx.Modules {
		if !activeOrOverload(mod) {
			continue
		}
		if !mod.EffectIDs.Has(hicBubbleEffectId) {
			continue
		}
		if mod.Charge != nil {
			continue // la variante con carica e' la bolla di warp-disruption focalizzata
		}
		ctx.Ship.AddAffliction(attrDisallowAssistance, aff("module", mod.ID, types.OpPostAssign, 1, nil), 0)

		massBonus := getBaseOr(mod, attrMassBonusPercentage, 0)
		sigBonus := getBaseOr(mod, attrSignatureRadiusBonusLegacy, 0)
		if massBonus != 0 {
			ctx.Ship.AddAffliction(attrMass, aff("module", mod.ID, types.OpPostPercent, massBonus/100, nil), 0)
		}
		if sigBonus != 0 {
			ctx.Ship.AddAffliction(attrSignatureRadius, aff("module", mod.ID, types.OpPostPercent, sigBonus/100, nil), 0)
		}
		speedFactorNerf := getBaseOr(mod, attrSpeedFactorBonus, 0)
		thrustNerf := getBaseOr(mod, attrSpeedBoostFactorBonus, 0)
		if speedFactorNerf == 0 && thrustNerf == 0 {
			continue
		}
		for _, target := range ctx.Modules {
			if target == mod {
				continue
			}
			if target.GroupID != propulsionModuleGroupId {
				continue
			}
			if speedFactorNerf != 0 {
				target.AddAffliction(attrSpeedFactor, aff("module", mod.ID, types.OpPostPercent, speedFactorNerf/100, nil), 0)
			}
			if thrustNerf != 0 {
				target.AddAffliction(attrSpeedBoostFactor, aff("module", mod.ID, types.OpPostPercent, thrustNerf/100, nil), 0)
			}
		}
	}
}

// ApplyLegacyMjdSigBloom: signature bloom del Micro Jump Drive.
func ApplyLegacyMjdSigBloom(ctx *FitContext) {
	for _, mod := range ctx.Modules {
		if !activeOrOverload(mod) {
			continue
		}
		isMjd := false
		for _, eid := range legacyMjdEffectIds.Values() {
			if mod.EffectIDs.Has(eid) {
				isMjd = true
				break
			}
		}
		if !isMjd {
			continue
		}
		bonusPct := mod.GetFinal(attrSignatureRadiusBonusPercent, 0)
		if bonusPct == 0 {
			continue
		}
		penalised := false
		sig := attrSignatureRadius
		for _, eid := range legacyMjdEffectIds.Values() {
			if mod.EffectIDs.Has(eid) {
				if _, ok := gen.StackPenaltyGroup(eid, &sig, gen.ReceiverShip, "module"); ok {
					penalised = true
					break
				}
			}
		}
		var group *string
		if penalised {
			group = defGroup(attrSignatureRadius)
		}
		ctx.Ship.AddAffliction(attrSignatureRadius, aff("module", mod.ID, types.OpPostPercent, bonusPct/100, group), 0)
	}
}

// ApplyLegacyCapBoosterInjection: nessuna operazione (la logica vive in derived/capacitor).
func ApplyLegacyCapBoosterInjection(_ *FitContext) {}

// CalculateProjectionRangeFactor: fattore di range per gli effetti proiettati con falloff.
// distance nil = non impostata (effetto pieno).
func CalculateProjectionRangeFactor(optimalRange, falloffRange float64, distance *float64) float64 {
	if distance == nil {
		return 1
	}
	if optimalRange == 0 && falloffRange == 0 {
		return 1
	}
	if falloffRange > 0 {
		if *distance > optimalRange+float64(3*falloffRange) {
			return 0
		}
		x := jsmath.Max(0, *distance-optimalRange) / falloffRange
		return math.Pow(0.5, x*x)
	}
	if *distance <= optimalRange {
		return 1
	}
	return 0
}

type projectionTargetKind int

const (
	targetShip projectionTargetKind = iota
	targetModulesRequiringSkill
	targetChargesRequiringSkill
)

type projectionDef struct {
	effectID int
	target   projectionTargetKind
	skillID  int
	pairs    [][2]int // (srcAttr, tgtAttr)
	friendly bool
}

var (
	pairsSensorDamp = [][2]int{{attrMaxTargetRangeBonus, attrMaxTargetRange}, {attrScanResolutionBonus, attrScanResolution}}
	pairsTrackDisr  = [][2]int{{attrTrackingSpeedBonus, attrTrackingSpeed}, {attrMaxRangeBonus, attrMaxRange}, {attrFalloffBonus, attrFalloff}}
	pairsGuidance   = [][2]int{{attrAoeCloudSizeBonus, attrAoeCloudSize}, {attrAoeVelocityBonus, attrAoeVelocity}, {attrMissileVelocityBonus, attrMaxVelocity}, {attrExplosionDelayBonus, attrExplosionDelay}}
	pairsPaint      = [][2]int{{attrSignatureRadiusBonusLegacy, attrSignatureRadius}}
	pairsWeb        = [][2]int{{attrSpeedFactor, attrMaxVelocity}}
)

// projectionEwarDefs: PROJECTION_EWAR_DEFS, nello stesso ordine.
var projectionEwarDefs = []projectionDef{
	{effectID: 6422, target: targetShip, pairs: pairsSensorDamp},
	{effectID: 6424, target: targetModulesRequiringSkill, skillID: skillGunnery, pairs: pairsTrackDisr},
	{effectID: 6423, target: targetChargesRequiringSkill, skillID: skillMissileLauncherOp, pairs: pairsGuidance},
	{effectID: 6425, target: targetShip, pairs: pairsPaint},
	{effectID: 6426, target: targetShip, pairs: pairsWeb},
	{effectID: 6476, target: targetShip, pairs: pairsWeb},   // AoE Web
	{effectID: 6478, target: targetShip, pairs: pairsPaint}, // AoE Paint
	{effectID: 6479, target: targetModulesRequiringSkill, skillID: skillGunnery, pairs: pairsTrackDisr},
	{effectID: 6479, target: targetChargesRequiringSkill, skillID: skillMissileLauncherOp, pairs: pairsGuidance},
	{effectID: 6481, target: targetShip, pairs: pairsSensorDamp}, // AoE Damp
	{effectID: 6682, target: targetShip, pairs: pairsWeb},        // Standup Web
	{effectID: 6683, target: targetShip, pairs: pairsPaint},      // Standup Paint
	{effectID: 6684, target: targetShip, pairs: pairsSensorDamp}, // Standup Sensor Damp
	{effectID: 6686, target: targetModulesRequiringSkill, skillID: skillGunnery, pairs: pairsTrackDisr},
	{effectID: 6686, target: targetChargesRequiringSkill, skillID: skillMissileLauncherOp, pairs: pairsGuidance},
	{effectID: 6427, target: targetShip, friendly: true, pairs: [][2]int{
		{attrMaxTargetRangeBonus, attrMaxTargetRange},
		{attrScanResolutionBonus, attrScanResolution},
		{1027 /* scanGravimetricStrengthPercent */, attrScanGravimetric},
		{1029 /* scanMagnetometricStrengthPercent */, attrScanMagnetometric},
		{1030 /* scanRadarStrengthPercent */, attrScanRadar},
		{1028 /* scanLadarStrengthPercent */, attrScanLadar},
	}},
	{effectID: 6428, target: targetModulesRequiringSkill, skillID: skillGunnery, friendly: true, pairs: pairsTrackDisr},
}

func applyProjectionEwarEffect(srcMod *ItemState, def *projectionDef, distance *float64, ctx *FitContext) {
	optimal := srcMod.GetFinal(attrRangeOptimal, 0)
	falloff := srcMod.GetFinal(attrRangeFalloff, 0)
	factor := CalculateProjectionRangeFactor(optimal, falloff, distance)
	if factor == 0 {
		return
	}

	// Il bersaglio puo' rifiutare l'effetto (disallowOffensiveModifiers / disallowAssistance).
	gateAttr := 872
	if def.friendly {
		gateAttr = attrDisallowAssistance
	}
	if ctx.Ship.GetFinal(gateAttr, 0) > 0 {
		return
	}

	pushAffliction := func(target *ItemState, srcAttr, tgtAttr int) {
		raw := srcMod.GetFinal(srcAttr, 0)
		if raw == 0 {
			return
		}
		g := fmt.Sprintf("proj:%d:%d", def.effectID, tgtAttr)
		target.AddAffliction(tgtAttr, aff("module", srcMod.ID, types.OpPostPercent, (raw/100)*factor, &g), 0)
	}

	switch def.target {
	case targetShip:
		for _, pair := range def.pairs {
			pushAffliction(ctx.Ship, pair[0], pair[1])
		}
		return
	case targetModulesRequiringSkill:
		sid := def.skillID
		for _, m := range ctx.Modules {
			if !ItemRequiresSkill(m, sid) {
				continue
			}
			for _, pair := range def.pairs {
				pushAffliction(m, pair[0], pair[1])
			}
		}
		return
	}
	sid := def.skillID
	for _, m := range ctx.Modules {
		if m.Charge == nil {
			continue
		}
		if !ItemRequiresSkill(m.Charge, sid) {
			continue
		}
		for _, pair := range def.pairs {
			pushAffliction(m.Charge, pair[0], pair[1])
		}
	}
}

// ApplyLegacyProjectionEwar applica ogni effetto EWAR proiettato che la sorgente porta.
func ApplyLegacyProjectionEwar(srcMod *ItemState, distance *float64, ctx *FitContext) {
	for i := range projectionEwarDefs {
		def := &projectionEwarDefs[i]
		if !srcMod.EffectIDs.Has(def.effectID) {
			continue
		}
		applyProjectionEwarEffect(srcMod, def, distance, ctx)
	}
}

type remoteRepDef struct {
	effectID   int
	amountAttr int
	layer      string // SHIELD | ARMOR | HULL
}

var projectionRemoteRepDefs = []remoteRepDef{
	{6186, attrShieldRepAmount, "SHIELD"},
	{6188, attrArmorRepAmount, "ARMOR"},
	{6185, attrHullRepAmount, "HULL"},
	{6651, attrArmorRepAmount, "ARMOR"},
	{6652, attrShieldRepAmount, "SHIELD"},
}

// RemoteRepReport e' il rapporto di buildRemoteRepReport.
type RemoteRepReport struct {
	Layer     string
	PerSecond float64
	EffectID  int
}

// BuildRemoteRepReport: rapporto di una riparazione remota ricevuta (nil = nessuno).
func BuildRemoteRepReport(srcMod *ItemState, distance *float64, ctx *FitContext) *RemoteRepReport {
	if ctx.Ship.GetFinal(attrDisallowAssistance, 0) > 0 {
		return nil
	}
	for _, def := range projectionRemoteRepDefs {
		if !srcMod.EffectIDs.Has(def.effectID) {
			continue
		}
		optimal := srcMod.GetFinal(attrRangeOptimal, 0)
		falloff := srcMod.GetFinal(attrRangeFalloff, 0)
		factor := CalculateProjectionRangeFactor(optimal, falloff, distance)
		if factor == 0 {
			return nil
		}
		amount := srcMod.GetFinal(def.amountAttr, 0) * factor
		cycleSec := srcMod.GetFinal(attrDurationMs, 0) / 1000
		if cycleSec <= 0 {
			return nil
		}
		return &RemoteRepReport{Layer: def.layer, PerSecond: amount / cycleSec, EffectID: def.effectID}
	}
	return nil
}

type capWarfareDef struct {
	effectID   int
	amountAttr int
	sign       float64
	gateAttr   int
	kind       string // NEUT | NOS | REMOTE_CAP
}

var projectionCapWarfareDefs = []capWarfareDef{
	{6184, attrPowerTransferAmount, -1, attrDisallowAssistance, "REMOTE_CAP"},
	{6187, attrNeutAmount, 1, attrDisallowOffensive, "NEUT"},
	{6197, attrPowerTransferAmount, 1, attrDisallowOffensive, "NOS"},
	{6216, attrNeutAmount, 1, attrDisallowOffensive, "NEUT"}, // Structure Energy Neutralizer
	{6477, attrNeutAmount, 1, attrDisallowOffensive, "NEUT"},
}

// CapWarfareReport e' il rapporto di buildCapWarfareReport.
type CapWarfareReport struct {
	Kind      string
	PerSecond float64
	EffectID  int
}

// BuildCapWarfareReport: neut/nos/remote cap ricevuto (nil = nessuno).
func BuildCapWarfareReport(srcMod *ItemState, distance *float64, ctx *FitContext) *CapWarfareReport {
	for _, def := range projectionCapWarfareDefs {
		if !srcMod.EffectIDs.Has(def.effectID) {
			continue
		}
		if ctx.Ship.GetFinal(def.gateAttr, 0) > 0 {
			return nil
		}
		optimal := srcMod.GetFinal(attrRangeOptimal, 0)
		falloff := srcMod.GetFinal(attrRangeFalloff, 0)
		factor := CalculateProjectionRangeFactor(optimal, falloff, distance)
		if factor == 0 {
			return nil
		}
		amount := srcMod.GetFinal(def.amountAttr, 0) * factor
		cycleSec := srcMod.GetFinal(attrDurationMs, 0) / 1000
		if cycleSec <= 0 {
			return nil
		}
		return &CapWarfareReport{Kind: def.kind, PerSecond: (def.sign * amount) / cycleSec, EffectID: def.effectID}
	}
	return nil
}

type sensorCandidate struct {
	kind      string
	strength  float64
	bonusAttr int
}

// pickPrimarySensor: il sensore piu' forte della nave (a parita' vince il primo).
func pickPrimarySensor(ship *ItemState) sensorCandidate {
	candidates := []sensorCandidate{
		{"radar", ship.GetFinal(attrScanRadarStr, 0), attrScanRadarBonus},
		{"ladar", ship.GetFinal(attrScanLadarStr, 0), attrScanLadarBonus},
		{"magnetometric", ship.GetFinal(attrScanMagnetometricStr, 0), attrScanMagnetometricBonus},
		{"gravimetric", ship.GetFinal(attrScanGravimetricStr, 0), attrScanGravimetricBonus},
	}
	best := candidates[0]
	for _, c := range candidates {
		if c.strength > best.strength {
			best = c
		}
	}
	return best
}

// BuildEcmProjectionReport: probabilita' di jam di un ECM proiettato (ok=false = nessun rapporto).
func BuildEcmProjectionReport(srcMod *ItemState, distance *float64, ctx *FitContext) (jamChance float64, ok bool) {
	isEcm := false
	for _, eid := range projectionEcmEffectIds.Values() {
		if srcMod.EffectIDs.Has(eid) {
			isEcm = true
			break
		}
	}
	if !isEcm {
		return 0, false
	}
	if ctx.Ship.GetFinal(attrDisallowOffensive, 0) > 0 {
		return 0, false
	}
	primary := pickPrimarySensor(ctx.Ship)
	if primary.strength <= 0 {
		return 0, false
	}
	sourceStrength := srcMod.GetFinal(primary.bonusAttr, 0)
	if sourceStrength <= 0 {
		return 0, false
	}
	optimal := srcMod.GetFinal(attrRangeOptimal, 0)
	falloff := srcMod.GetFinal(attrRangeFalloff, 0)
	factor := CalculateProjectionRangeFactor(optimal, falloff, distance)
	if factor == 0 {
		return 0, false
	}
	effectiveStrength := sourceStrength * factor
	return jsmath.Max(0, jsmath.Min(1, effectiveStrength/primary.strength)), true
}

// ApplyLegacyOverload: bonus di surriscaldamento dei moduli in OVERLOAD.
func ApplyLegacyOverload(ctx *FitContext) {
	for _, mod := range ctx.Modules {
		if mod.State != types.StateOverload {
			continue
		}
		for _, p := range overheatPairs {
			srcAttr, tgtAttr := p[0], p[1]
			bonus := getBaseOr(mod, srcAttr, 0)
			if bonus == 0 {
				continue
			}
			mod.AddAffliction(tgtAttr, aff("module", mod.ID, types.OpPostPercent, bonus/100, nil), 0)
		}
	}
}

// ApplyLegacyBastion: Bastion Module (effetto 6658).
func ApplyLegacyBastion(ctx *FitContext) {
	for _, mod := range ctx.Modules {
		if !activeOrOverload(mod) {
			continue
		}
		if !mod.EffectIDs.Has(bastionEffectId) {
			continue
		}
		ship := ctx.Ship
		sourceID := mod.ID

		// Resistenze: il Bastion le moltiplica leggendo l'attributo del proprio item.
		resPairs := [12][2]int{
			{attrShieldEmRes, attrShieldEmRes},
			{attrShieldThermRes, attrShieldThermRes},
			{attrShieldKinRes, attrShieldKinRes},
			{attrShieldExpRes, attrShieldExpRes},
			{attrArmorEmRes, attrArmorEmRes},
			{attrArmorThermRes, attrArmorThermRes},
			{attrArmorKinRes, attrArmorKinRes},
			{attrArmorExpRes, attrArmorExpRes},
			{attrHullEmRes, 974},
			{attrHullThermRes, 977},
			{attrHullKinRes, 976},
			{attrHullExpRes, 975},
		}
		isHullRes := func(a int) bool {
			return a == attrHullEmRes || a == attrHullThermRes || a == attrHullKinRes || a == attrHullExpRes
		}
		for _, p := range resPairs {
			tgt, src := p[0], p[1]
			factor := getBaseOr(mod, src, 1)
			if factor == 1 {
				continue
			}
			var g *string
			if !isHullRes(tgt) {
				g = defGroup(tgt)
			}
			ship.AddAffliction(tgt, aff("module", sourceID, types.OpPreMul, factor, g), 0)
		}

		// Torrette grandi: range, falloff e ROF.
		maxRangeBonus := mod.GetFinal(attrMaxRangeBonus, 0)
		falloffBonusV := mod.GetFinal(attrFalloffBonus, 0)
		turretRofBonus := mod.GetFinal(3109 /* bastionTurretROFBonus */, 0)
		matchTurret := func(m *ItemState) bool {
			return ItemRequiresSkill(m, skillLargeEnergyTurret) ||
				ItemRequiresSkill(m, skillLargeHybridTurret) ||
				ItemRequiresSkill(m, skillLargeProjectileTurret) ||
				ItemRequiresSkill(m, skillLargePrecursorWeapon)
		}
		matchTurretFalloff := func(m *ItemState) bool {
			return ItemRequiresSkill(m, skillLargeEnergyTurret) ||
				ItemRequiresSkill(m, skillLargeHybridTurret) ||
				ItemRequiresSkill(m, skillLargeProjectileTurret)
		}
		pp := func(target *ItemState, attr int, v float64, g *string) {
			target.AddAffliction(attr, aff("module", sourceID, types.OpPostPercent, v/100, g), 0)
		}
		for _, target := range ctx.Modules {
			if matchTurret(target) {
				if maxRangeBonus != 0 {
					pp(target, attrMaxRange, maxRangeBonus, defGroup(attrMaxRange))
				}
				if turretRofBonus != 0 {
					pp(target, attrSpeed, turretRofBonus, defGroup(attrSpeed))
				}
			}
			if matchTurretFalloff(target) && falloffBonusV != 0 {
				pp(target, attrFalloff, falloffBonusV, defGroup(attrFalloff))
			}
		}

		// Missili: velocita' delle cariche pesanti/torpedini/cruise e ROF dei lanciatori.
		missileVelocityBonus := mod.GetFinal(attrMissileVelocityBonus, 0)
		missileRofBonus := mod.GetFinal(3108 /* bastionMissileROFBonus */, 0)
		matchMissileCharge := func(c *ItemState) bool {
			return ItemRequiresSkill(c, skillTorpedoes) || ItemRequiresSkill(c, skillCruiseMissiles) || ItemRequiresSkill(c, skillHeavyMissiles)
		}
		matchMissileLauncherROF := func(m *ItemState) bool {
			return ItemRequiresSkill(m, skillCruiseMissiles) || ItemRequiresSkill(m, skillTorpedoes) ||
				ItemRequiresSkill(m, skillTorpedoSpec) || ItemRequiresSkill(m, skillCruiseMissileSpec)
		}
		for _, target := range ctx.Modules {
			if target.Charge != nil && matchMissileCharge(target.Charge) && missileVelocityBonus != 0 {
				pp(target.Charge, attrMaxVelocity, missileVelocityBonus, defGroup(attrMaxVelocity))
			}
			if matchMissileLauncherROF(target) && missileRofBonus != 0 {
				pp(target, attrSpeed, missileRofBonus, defGroup(attrSpeed))
			}
		}

		// Riparazioni armor/shield: quantita' e durata/cap.
		armorRepBonus := mod.GetFinal(attrArmorDamageAmountBonus, 0)
		shieldBoostBonus := mod.GetFinal(attrShieldBoostMultiplier, 0)
		repairDurationBonus := mod.GetFinal(attrBastionRepairDurationBonus, 0)
		for _, target := range ctx.Modules {
			isArmor := ItemRequiresSkill(target, skillRepairSystems)
			isShield := ItemRequiresSkill(target, skillShieldOperation)
			if isArmor && armorRepBonus != 0 {
				pp(target, attrArmorRepAmount, armorRepBonus, defGroup(attrArmorRepAmount))
			}
			if isShield && shieldBoostBonus != 0 {
				pp(target, attrShieldRepAmount, shieldBoostBonus, defGroup(attrShieldRepAmount))
			}
			if (isArmor || isShield) && repairDurationBonus != 0 {
				pp(target, attrDurationMs, repairDurationBonus, nil)
				pp(target, attrCapacitorNeed, repairDurationBonus, nil)
			}
		}

		// Velocita' della nave.
		speedPenalty := mod.GetFinal(attrSpeedFactor, 0)
		if speedPenalty != 0 {
			ship.AddAffliction(attrMaxVelocity, aff("module", sourceID, types.OpPostPercent, speedPenalty/100, nil), 0)
		}

		// Sensori.
		sensorBonusPairs := [4][2]int{
			{1027, attrScanGravimetric},
			{1028, attrScanLadar},
			{1029, attrScanMagnetometric},
			{1030, attrScanRadar},
		}
		for _, p := range sensorBonusPairs {
			v := mod.GetFinal(p[0], 0)
			if v == 0 {
				continue
			}
			ship.AddAffliction(p[1], aff("module", sourceID, types.OpPostPercent, v/100, defGroup(p[1])), 0)
		}

		// Resistenze agli effetti remoti.
		resistPairs := [5][2]int{
			{attrRemoteRepairImpedanceBonus, attrRemoteRepairImpedance},
			{attrRemoteAssistImpedanceBonus, attrRemoteAssistImpedance},
			{attrSensorDampResistBonus, attrSensorDampResist},
			{attrTargetPaintResistBonus, attrTargetPaintResist},
			{attrWeaponDisruptResistBonus, attrWeaponDisruptResist},
		}
		for _, p := range resistPairs {
			v := mod.GetFinal(p[0], 0)
			if v == 0 {
				continue
			}
			ship.AddAffliction(p[1], aff("module", sourceID, types.OpPostPercent, v/100, nil), 0)
		}

		// Blocco dei Micro Jump Drive.
		mjdBlock := mod.GetFinal(attrActivationBlockedStr, 0)
		if mjdBlock != 0 {
			for _, target := range ctx.Modules {
				if ItemRequiresSkill(target, skillMjdOperation) {
					target.AddAffliction(attrActivationBlocked, aff("module", sourceID, types.OpModAdd, mjdBlock, nil), 0)
				}
			}
		}

		// Stato warp scramble.
		siegeStatus := mod.GetFinal(attrSiegeWarpStatus, 0)
		if siegeStatus != 0 {
			ship.AddAffliction(attrWarpScrambleStatus, aff("module", sourceID, types.OpModAdd, siegeStatus, nil), 0)
		}
	}
}

// ApplyLegacyMutadaptiveSpool: spool del Mutadaptive Remote Armor Repairer (effetto 7166).
func ApplyLegacyMutadaptiveSpool(ctx *FitContext, spoolPercent float64) {
	if spoolPercent <= 0 {
		return
	}
	clampedSpool := jsmath.Max(0, jsmath.Min(1, spoolPercent))
	for _, mod := range ctx.Modules {
		if !activeOrOverload(mod) {
			continue
		}
		if !mod.EffectIDs.Has(mutadaptiveRarEffectId) {
			continue
		}
		maxBonusPct := mod.GetFinal(attrMutadaptiveBonusMax, 0)
		if maxBonusPct <= 0 {
			continue
		}
		mod.AddAffliction(attrArmorRepAmount, aff("module", mod.ID, types.OpPostPercent, (maxBonusPct/100)*clampedSpool, nil), 0)
	}
}

// ApplyLegacyDoomsdaySelfEffects: effetti su se' stessi dei doomsday (velocita', stato warp).
func ApplyLegacyDoomsdaySelfEffects(ctx *FitContext) {
	for _, mod := range ctx.Modules {
		if !activeOrOverload(mod) {
			continue
		}
		isDD := false
		for _, eid := range legacyDoomsdayEffectIds.Values() {
			if mod.EffectIDs.Has(eid) {
				isDD = true
				break
			}
		}
		if !isDD {
			continue
		}
		speedFactor := mod.GetFinal(attrSpeedFactor, 0)
		if speedFactor != 0 {
			ctx.Ship.AddAffliction(attrMaxVelocity, aff("module", mod.ID, types.OpPostPercent, speedFactor/100, defGroup(attrMaxVelocity)), 0)
		}
		siegeStatus := mod.GetFinal(attrSiegeWarpStatus, 0)
		if siegeStatus != 0 {
			ctx.Ship.AddAffliction(attrWarpScrambleStatus, aff("module", mod.ID, types.OpModAdd, siegeStatus, nil), 0)
		}
	}
}

// operationByDbuffName e' OPERATION_BY_DBUFF_NAME.
func operationByDbuffName(name string) (types.ModifierOperation, bool) {
	switch name {
	case "PreAssignment":
		return types.OpPreAssign, true
	case "PreMul":
		return types.OpPreMul, true
	case "PreDiv":
		return types.OpPreDiv, true
	case "ModAdd":
		return types.OpModAdd, true
	case "ModSub":
		return types.OpModSub, true
	case "PostMul":
		return types.OpPostMul, true
	case "PostDiv":
		return types.OpPostDiv, true
	case "PostPercent":
		return types.OpPostPercent, true
	case "PostAssignment":
		return types.OpPostAssign, true
	}
	return "", false
}

// ApplyLegacyCommandBursts: Command Burst (Warfare Link) caricati con una carica che porta i buff.
func ApplyLegacyCommandBursts(ctx *FitContext) {
	for _, mod := range ctx.Modules {
		if !activeOrOverload(mod) {
			continue
		}
		isBurst := false
		for _, eid := range commandBurstEffectIds.Values() {
			if mod.EffectIDs.Has(eid) {
				isBurst = true
				break
			}
		}
		if !isBurst {
			continue
		}
		charge := mod.Charge
		if charge == nil {
			continue
		}
		for _, p := range attrWarfareBuffIds {
			idAttr, valueAttr := p[0], p[1]
			buffID := getBaseOr(charge, idAttr, 0)
			if buffID == 0 {
				continue
			}
			value := mod.GetFinal(valueAttr, 0)
			if value == 0 {
				continue
			}
			applyCommandBuff(buffID, value, mod, ctx)
		}
	}
}

// applyCommandBuff applica un buff di comando (dbuff collection) a nave, moduli, droni e fighter.
func applyCommandBuff(buffID float64, value float64, sourceMod *ItemState, ctx *FitContext) {
	// buffID e' un numero JS usato come chiave di Map: un non-intero non trova nulla.
	if buffID != math.Trunc(buffID) || math.IsInf(buffID, 0) {
		return
	}
	dbuff, ok := ctx.Dataset.DbuffCollections.Get(int(buffID))
	if !ok {
		return
	}
	op, ok := operationByDbuffName(dbuff.OperationName)
	if !ok {
		return
	}
	scaledValue := value
	if op == types.OpPostPercent {
		scaledValue = value / 100
	}
	affliction := func(operation types.ModifierOperation, attrID int) types.ModifierAffliction {
		return aff("module", sourceMod.ID, operation, scaledValue, defGroup(attrID))
	}

	for _, mi := range dbuff.ItemModifiers {
		ctx.Ship.AddAffliction(mi.DogmaAttributeID, affliction(op, mi.DogmaAttributeID), 0)
	}
	for _, mi := range dbuff.LocationModifiers {
		ctx.Ship.AddAffliction(mi.DogmaAttributeID, affliction(op, mi.DogmaAttributeID), 0)
		for _, m := range ctx.Modules {
			m.AddAffliction(mi.DogmaAttributeID, affliction(op, mi.DogmaAttributeID), 0)
		}
		for _, d := range ctx.Drones {
			d.AddAffliction(mi.DogmaAttributeID, affliction(op, mi.DogmaAttributeID), 0)
		}
		for _, f := range ctx.Fighters {
			f.AddAffliction(mi.DogmaAttributeID, affliction(op, mi.DogmaAttributeID), 0)
		}
	}
	for _, mi := range dbuff.LocationGroupModifiers {
		for _, m := range ctx.Modules {
			if m.GroupID == mi.GroupID {
				m.AddAffliction(mi.DogmaAttributeID, affliction(op, mi.DogmaAttributeID), 0)
			}
			if m.Charge != nil && m.Charge.GroupID == mi.GroupID {
				m.Charge.AddAffliction(mi.DogmaAttributeID, affliction(op, mi.DogmaAttributeID), 0)
			}
		}
		for _, d := range ctx.Drones {
			if d.GroupID == mi.GroupID {
				d.AddAffliction(mi.DogmaAttributeID, affliction(op, mi.DogmaAttributeID), 0)
			}
		}
	}
	for _, mi := range dbuff.LocationRequiredSkillModifiers {
		for _, m := range ctx.Modules {
			if ItemRequiresSkill(m, mi.SkillID) {
				m.AddAffliction(mi.DogmaAttributeID, affliction(op, mi.DogmaAttributeID), 0)
			}
			if m.Charge != nil && ItemRequiresSkill(m.Charge, mi.SkillID) {
				m.Charge.AddAffliction(mi.DogmaAttributeID, affliction(op, mi.DogmaAttributeID), 0)
			}
		}
		for _, d := range ctx.Drones {
			if ItemRequiresSkill(d, mi.SkillID) {
				d.AddAffliction(mi.DogmaAttributeID, affliction(op, mi.DogmaAttributeID), 0)
			}
		}
	}
}

// itemRequiresSkillOneOf: l'item richiede almeno una delle skill (via requiredSkills()).
func itemRequiresSkillOneOf(item *ItemState, skillIDs *intSet) bool {
	for _, req := range item.RequiredSkills() {
		if req.SkillID == math.Trunc(req.SkillID) && skillIDs.Has(int(req.SkillID)) {
			return true
		}
	}
	return false
}

type receiver int

const (
	recvShip receiver = iota
	recvTurrets
	recvMissiles
	recvMissileCharges
	recvSmartbombs
	recvShieldBoosters
	recvArmorRepairers
	recvRemoteShieldTransfer
	recvRemoteArmorRep
	recvEnergyNeut
	recvEnergyNos
	recvWebifiers
	recvTargetPainters
	recvDrones
)

type systemEffectSpec struct {
	beaconAttr int
	receiver   receiver
	targetAttr int
	operation  types.ModifierOperation
}

// systemEffectSpecs: SYSTEM_EFFECT_SPECS nello stesso ordine.
var systemEffectSpecs = []systemEffectSpec{
	{1465, recvShip, 267, types.OpPostPercent},
	{1466, recvShip, 269, types.OpPostPercent},
	{1467, recvShip, 270, types.OpPostPercent},
	{1468, recvShip, 268, types.OpPostPercent},
	{1489, recvShip, 271, types.OpPostPercent},
	{1490, recvShip, 272, types.OpPostPercent},
	{1491, recvShip, 273, types.OpPostPercent},
	{1492, recvShip, 274, types.OpPostPercent},
	{984, recvShip, 113, types.OpPostPercent},
	{985, recvShip, 111, types.OpPostPercent},
	{986, recvShip, 109, types.OpPostPercent},
	{987, recvShip, 110, types.OpPostPercent},
	{146, recvShip, 263, types.OpPostMul},
	{148, recvShip, 265, types.OpPostMul},
	{169, recvShip, 70, types.OpPostMul},
	{237, recvShip, 76, types.OpPostMul},
	{652, recvShip, 552, types.OpPostMul},
	{1499, recvShip, 482, types.OpPostMul},
	{1500, recvShip, 55, types.OpPostMul},
	{1840, recvShip, 479, types.OpPostMul},
	{327, recvShip, 9, types.OpPostPercent},
	{151, recvShip, 70, types.OpPostPercent},
	{309, recvShip, 76, types.OpPostPercent},
	{566, recvShip, 564, types.OpPostPercent},
	{601, recvShip, 600, types.OpPostPercent},
	{1950, recvShip, 600, types.OpModAdd},
	{1851, recvShip, 208, types.OpPostPercent},
	{20, recvShip, 37, types.OpPostPercent},
	{306, recvShip, 37, types.OpPostPercent},
	{1482, recvTurrets, 64, types.OpPostMul},
	{244, recvTurrets, 160, types.OpPostMul},
	{767, recvTurrets, 160, types.OpPostPercent},
	{1469, recvMissileCharges, 37, types.OpPostMul},
	{1470, recvMissileCharges, 281, types.OpPostMul},
	{1483, recvMissileCharges, 653, types.OpPostMul},
	{1967, recvMissileCharges, 654, types.OpPostMul},
	{1487, recvSmartbombs, 97, types.OpPostMul},
	{1488, recvSmartbombs, 114, types.OpPostMul},
	{1488, recvSmartbombs, 116, types.OpPostMul},
	{1488, recvSmartbombs, 117, types.OpPostMul},
	{1488, recvSmartbombs, 118, types.OpPostMul},
	{1495, recvArmorRepairers, 84, types.OpPostMul},
	{1496, recvShieldBoosters, 68, types.OpPostMul},
	{1497, recvRemoteShieldTransfer, 68, types.OpPostMul},
	{1498, recvRemoteArmorRep, 84, types.OpPostMul},
	{1966, recvEnergyNeut, 97, types.OpPostMul},
	{1966, recvEnergyNos, 97, types.OpPostMul},
	{1968, recvTargetPainters, 554, types.OpPostMul},
	{1969, recvWebifiers, 20, types.OpPostMul},
	{591, recvDrones, 37, types.OpPostPercent},
	{1255, recvDrones, 64, types.OpPostPercent},
	{1485, recvShip, 1211, types.OpPostMul},
	{1229, recvShip, 1211, types.OpPostPercent},
}

// receiverPredicate: RECEIVER_PREDICATES (ship, missileCharges e drones sono gestiti a parte: sempre false).
func receiverPredicate(r receiver, m *ItemState) bool {
	switch r {
	case recvTurrets:
		for _, eid := range m.EffectIDs.Values() {
			if eid == 10 || eid == 34 || eid == 6995 || eid == 8037 {
				return true
			}
		}
		return false
	case recvMissiles:
		return m.EffectIDs.Has(9) || m.EffectIDs.Has(101)
	case recvSmartbombs:
		return m.GroupID == smartBombGroupId
	case recvShieldBoosters:
		return intersects(m.EffectIDs, shieldBoosterEffectIds)
	case recvArmorRepairers:
		return intersects(m.EffectIDs, armorRepairerEffectIds)
	case recvRemoteShieldTransfer:
		return intersects(m.EffectIDs, remoteShieldXferEffectIds)
	case recvRemoteArmorRep:
		return intersects(m.EffectIDs, remoteArmorRepEffectIds)
	case recvEnergyNeut:
		return intersects(m.EffectIDs, energyNeutralizerEffectIds)
	case recvEnergyNos:
		return intersects(m.EffectIDs, energyNosferatuEffectIds)
	case recvWebifiers:
		return intersects(m.EffectIDs, stasisWebEffectIds)
	case recvTargetPainters:
		return intersects(m.EffectIDs, targetPainterEffectIds)
	}
	return false
}

func intersects(a, b *intSet) bool {
	for i := 0; i < a.Len(); i++ {
		if b.Has(a.At(i)) {
			return true
		}
	}
	return false
}

// ApplyLegacySystemEffect applica gli effetti di sistema (Incursion/Triglavian/Drifter/Wormhole) del
// beacon scelto. beaconTypeID nil o 0 = nessun beacon.
func ApplyLegacySystemEffect(ctx *FitContext, beaconTypeID *int) {
	if beaconTypeID == nil || *beaconTypeID == 0 {
		return
	}
	beacon, ok := ctx.Dataset.GetType(*beaconTypeID)
	if !ok {
		return
	}
	// new Map(attributes.map(a => [a.id, a.v])): l'ultimo vince; solo letture per chiave.
	beaconAttrs := map[int]float64{}
	for _, a := range beacon.Attributes {
		beaconAttrs[a.ID] = a.V
	}
	ship := ctx.Ship
	sourceID := fmt.Sprintf("beacon:%d", *beaconTypeID)

	for _, spec := range systemEffectSpecs {
		raw, ok := beaconAttrs[spec.beaconAttr]
		if !ok || raw == 0 {
			continue
		}
		value := raw
		if spec.operation == types.OpPostPercent {
			value = raw / 100
		}
		a := aff("module", sourceID, spec.operation, value, nil)
		switch spec.receiver {
		case recvShip:
			ship.AddAffliction(spec.targetAttr, a, 0)
			continue
		case recvDrones:
			for _, d := range ctx.Drones {
				d.AddAffliction(spec.targetAttr, a, 0)
			}
			continue
		case recvMissileCharges:
			for _, m := range ctx.Modules {
				if m.Charge == nil {
					continue
				}
				if !m.EffectIDs.Has(9) && !m.EffectIDs.Has(101) {
					continue
				}
				m.Charge.AddAffliction(spec.targetAttr, a, 0)
			}
			continue
		}
		for _, m := range ctx.Modules {
			if !receiverPredicate(spec.receiver, m) {
				continue
			}
			m.AddAffliction(spec.targetAttr, a, 0)
		}
	}

	damageAttrs := [4]int{114, 116, 117, 118}
	for _, beaconAttrID := range [2]int{1686, 203} {
		reduction := beaconAttrs[beaconAttrID]
		if reduction == 0 {
			continue
		}
		a := aff("module", sourceID, types.OpPostPercent, reduction/100, nil)
		for _, m := range ctx.Modules {
			if m.GroupID == smartBombGroupId {
				for _, attr := range damageAttrs {
					m.AddAffliction(attr, a, 0)
				}
				continue
			}
			if m.Charge != nil && ItemRequiresSkill(m.Charge, skillMissileLauncherOp) {
				for _, attr := range damageAttrs {
					m.Charge.AddAffliction(attr, a, 0)
				}
			}
		}
	}

	smallWeaponMul := beaconAttrs[1493]
	if smallWeaponMul != 0 && smallWeaponMul != 1 {
		a := aff("module", sourceID, types.OpPostMul, smallWeaponMul, nil)
		for _, m := range ctx.Modules {
			isSmallTurret := (m.EffectIDs.Has(10) || m.EffectIDs.Has(34) || m.EffectIDs.Has(6995)) &&
				itemRequiresSkillOneOf(m, smallTurretSkillIds)
			isSmallMissile := (m.EffectIDs.Has(9) || m.EffectIDs.Has(101)) &&
				itemRequiresSkillOneOf(m, smallMissileSkillIds)
			if isSmallTurret {
				m.AddAffliction(64 /* damageMultiplier */, a, 0)
			} else if isSmallMissile && m.Charge != nil {
				for _, attr := range damageAttrs {
					m.Charge.AddAffliction(attr, a, 0)
				}
			}
		}
	}

	durationBonus := beaconAttrs[66]
	if durationBonus != 0 {
		a := aff("module", sourceID, types.OpPostPercent, durationBonus/100, nil)
		for _, m := range ctx.Modules {
			m.AddAffliction(73, a, 0)
			m.AddAffliction(51, a, 0)
		}
	}
}

// ApplyLegacyFighterAbilities: abilita' passive dei fighter (AB, MWD, Evasive Maneuvers).
func ApplyLegacyFighterAbilities(ctx *FitContext) {
	for _, fighter := range ctx.Fighters {
		pp := func(attr int, pct float64, g *string) {
			fighter.AddAffliction(attr, aff("fighter", fighter.ID, types.OpPostPercent, pct/100, g), 0)
		}
		if fighter.EffectIDs.Has(fighterAbEffectId) {
			speedPct := fighter.GetFinal(attrFighterAbSpeedBonus, 0)
			if speedPct != 0 {
				pp(attrMaxVelocity, speedPct, defGroup(attrMaxVelocity))
			}
		}
		if fighter.EffectIDs.Has(fighterMwdEffectId) {
			speedPct := fighter.GetFinal(attrFighterMwdSpeedBonus, 0)
			if speedPct != 0 {
				pp(attrMaxVelocity, speedPct, defGroup(attrMaxVelocity))
			}
			sigPct := fighter.GetFinal(attrFighterMwdSigBonus, 0)
			if sigPct != 0 {
				pp(attrSignatureRadius, sigPct, defGroup(attrSignatureRadius))
			}
		}
		if fighter.EffectIDs.Has(fighterEvasiveEffectId) {
			speedPct := fighter.GetFinal(attrFighterEvasiveSpeedBonus, 0)
			if speedPct != 0 {
				pp(attrMaxVelocity, speedPct, defGroup(attrMaxVelocity))
			}
			sigPct := fighter.GetFinal(attrFighterEvasiveSigBonus, 0)
			if sigPct != 0 {
				pp(attrSignatureRadius, sigPct, defGroup(attrSignatureRadius))
			}
			resonances := [4][2]int{
				{attrFighterEvasiveEmRes, attrShieldEmReso},
				{attrFighterEvasiveThmRes, attrShieldThmReso},
				{attrFighterEvasiveKinRes, attrShieldKinReso},
				{attrFighterEvasiveExpRes, attrShieldExpReso},
			}
			for _, p := range resonances {
				mul := fighter.GetFinal(p[0], 0)
				if mul == 0 || mul == 1 {
					continue
				}
				fighter.AddAffliction(p[1], aff("fighter", fighter.ID, types.OpPostMul, mul, defGroup(p[1])), 0)
			}
		}
	}
}

// DisintegratorSpoolBonus: bonus di danno al grado di spool dato.
func DisintegratorSpoolBonus(maxBonus, spoolPercent float64) float64 {
	if maxBonus <= 0 {
		return 0
	}
	p := jsmath.Max(0, jsmath.Min(1, spoolPercent))
	return maxBonus * p
}

// DisintegratorCyclesToFullSpool: cicli per arrivare a spool pieno.
func DisintegratorCyclesToFullSpool(maxBonus, bonusPerCycle float64) float64 {
	if maxBonus <= 0 || bonusPerCycle <= 0 {
		return 0
	}
	return math.Ceil(maxBonus / bonusPerCycle)
}

// ApplyLegacyDisintegratorSpool: spool dei disintegratori Triglavian (effetto 6995).
func ApplyLegacyDisintegratorSpool(ctx *FitContext, spoolPercent float64) {
	p := jsmath.Max(0, jsmath.Min(1, spoolPercent))
	if p <= 0 {
		return
	}
	for _, mod := range ctx.Modules {
		if !activeOrOverload(mod) {
			continue
		}
		if !mod.EffectIDs.Has(disintegratorEffectId) {
			continue
		}
		if !mod.HasAttr(attrDmgMultBonusMax) {
			continue
		}
		maxBonus := mod.GetFinal(attrDmgMultBonusMax, 0)
		if maxBonus <= 0 {
			continue
		}
		mod.AddAffliction(attrDamageMultiplier, aff("module", mod.ID, types.OpPostPercent, p*maxBonus, nil), 0)
	}
}

// ApplyLegacyRAH: Reactive Armor Hardener (effetto 4928): simula il ciclo di adattamento contro il
// profilo di danno e applica la media delle resistenze del ciclo come PreMul alla nave.
func ApplyLegacyRAH(ctx *FitContext, damageProfile *types.DamageProfile) {
	if damageProfile == nil {
		return
	}
	if damageProfile.Em+damageProfile.Thermal+damageProfile.Kinetic+damageProfile.Explosive <= 0 {
		return
	}
	for _, mod := range ctx.Modules {
		if !activeOrOverload(mod) {
			continue
		}
		if !mod.EffectIDs.Has(rahEffectId) {
			continue
		}
		ship := ctx.Ship

		// Danno ricevuto per tipo con le resistenze correnti della nave (senza la RAH).
		baseDamageTaken := [4]float64{
			damageProfile.Em * ship.GetFinal(attrArmorEmRes, 1),
			damageProfile.Thermal * ship.GetFinal(attrArmorThermRes, 1),
			damageProfile.Kinetic * ship.GetFinal(attrArmorKinRes, 1),
			damageProfile.Explosive * ship.GetFinal(attrArmorExpRes, 1),
		}
		resistanceShift := orElse(mod.GetFinal(attrRahResistanceShift, 0), 6) / 100
		rahResistance := [4]float64{
			getBaseOr(mod, attrArmorEmRes, 1),
			getBaseOr(mod, attrArmorThermRes, 1),
			getBaseOr(mod, attrArmorKinRes, 1),
			getBaseOr(mod, attrArmorExpRes, 1),
		}

		var cycles [][4]float64
		loopStart := -1
		for cycle := 0; cycle < 50; cycle++ {
			// [indice, danno, resistenza]
			tuples := [4][3]float64{
				{0, baseDamageTaken[0] * rahResistance[0], rahResistance[0]},
				{3, baseDamageTaken[3] * rahResistance[3], rahResistance[3]},
				{2, baseDamageTaken[2] * rahResistance[2], rahResistance[2]},
				{1, baseDamageTaken[1] * rahResistance[1], rahResistance[1]},
			}
			sort.SliceStable(tuples[:], func(i, j int) bool {
				return jsmath.CmpResult(tuples[i][1]-tuples[j][1]) < 0
			})

			var change0, change1, change2, change3 float64
			if tuples[2][1] == 0 {
				change0 = 1 - tuples[0][2]
				change1 = 1 - tuples[1][2]
				change2 = 1 - tuples[2][2]
				change3 = -(change0 + change1 + change2)
			} else if tuples[1][1] == 0 {
				change0 = 1 - tuples[0][2]
				change1 = 1 - tuples[1][2]
				change2 = -(change0 + change1) / 2
				change3 = -(change0 + change1) / 2
			} else {
				change0 = jsmath.Min(resistanceShift, 1-tuples[0][2])
				change1 = jsmath.Min(resistanceShift, 1-tuples[1][2])
				change2 = -(change0 + change1) / 2
				change3 = -(change0 + change1) / 2
			}
			rahResistance[int(tuples[0][0])] = tuples[0][2] + change0
			rahResistance[int(tuples[1][0])] = tuples[1][2] + change1
			rahResistance[int(tuples[2][0])] = tuples[2][2] + change2
			rahResistance[int(tuples[3][0])] = tuples[3][2] + change3

			// Cerca un ciclo gia' visto (a meno della tolleranza): da li' inizia il loop.
			const tolerance = 1e-6
			for i := 0; i < len(cycles); i++ {
				v := cycles[i]
				if math.Abs(rahResistance[0]-v[0]) <= tolerance &&
					math.Abs(rahResistance[1]-v[1]) <= tolerance &&
					math.Abs(rahResistance[2]-v[2]) <= tolerance &&
					math.Abs(rahResistance[3]-v[3]) <= tolerance {
					loopStart = i
					break
				}
			}
			if loopStart >= 0 {
				break
			}
			cycles = append(cycles, rahResistance)
		}

		var loopCycles [][4]float64
		if loopStart >= 0 {
			loopCycles = cycles[loopStart:]
		} else {
			start := len(cycles) - 20
			if start < 0 {
				start = 0
			}
			loopCycles = cycles[start:]
		}
		if len(loopCycles) == 0 {
			continue
		}

		avg := [4]float64{0, 0, 0, 0}
		for _, c := range loopCycles {
			avg[0] += c[0]
			avg[1] += c[1]
			avg[2] += c[2]
			avg[3] += c[3]
		}
		for i := 0; i < 4; i++ {
			avg[i] = avg[i] / float64(len(loopCycles))
		}

		targets := [4][2]float64{
			{avg[0], attrArmorEmRes},
			{avg[1], attrArmorThermRes},
			{avg[2], attrArmorKinRes},
			{avg[3], attrArmorExpRes},
		}
		for _, p := range targets {
			factor, tgtAttr := p[0], int(p[1])
			ctx.Ship.AddAffliction(tgtAttr, aff("module", mod.ID, types.OpPreMul, factor, defGroup(tgtAttr)), 0)
		}
	}
}

// ApplyLegacySubsystemSlots: gli slot/hardpoint dei subsystem T3C si sommano alla nave.
func ApplyLegacySubsystemSlots(ctx *FitContext) {
	for _, sub := range ctx.Subsystems {
		for _, p := range subsystemSlotBonusMap {
			sourceAttr, targetAttr := p[0], p[1]
			bonus := getBaseOr(sub, sourceAttr, 0)
			if bonus == 0 {
				continue
			}
			ctx.Ship.AddAffliction(targetAttr, aff("subsystem", sub.ID, types.OpModAdd, bonus, nil), 0)
		}
	}
}

// subsystemAddPassiveEffects: (effectID, srcAttr, tgtAttr, nome descrittivo).
var subsystemAddPassiveEffects = []struct {
	effectID, srcAttr, tgtAttr int
	name                       string
}{
	{3831, 263, 263, "shieldCapacity"},
	{3771, 1159, 265, "armorHP"},
	{6920, 2688, 9, "structureHP"},
	{3810, 2689, 38, "cargoCapacity"},
	{3811, 482, 482, "capacitorCapacity"},
	{3808, 552, 552, "signatureRadius"},
}

// LegacyHandledEffectIds: effetti rivendicati da gestori hardcoded, che il dispatcher generico salta
// (LEGACY_HANDLED_PASSIVE_ADD_EFFECTS, poi LEGACY_HANDLED_HARDCODED_EFFECTS).
var LegacyHandledEffectIds = newLegacyHandled()

func newLegacyHandled() *intSet {
	s := newIntSet(nil)
	for _, e := range subsystemAddPassiveEffects {
		s.Add(e.effectID)
	}
	for _, id := range legacyHandledHardcodedEffects.Values() {
		s.Add(id)
	}
	return s
}

// ApplyLegacySubsystemAddPassive: bonus "AddPassive" dei subsystem T3C (HP, capacita', ecc.).
func ApplyLegacySubsystemAddPassive(ctx *FitContext) {
	for _, sub := range ctx.Subsystems {
		for _, e := range subsystemAddPassiveEffects {
			if !sub.EffectIDs.Has(e.effectID) {
				continue
			}
			value := getBaseOr(sub, e.srcAttr, 0)
			if value == 0 {
				continue
			}
			ctx.Ship.AddAffliction(e.tgtAttr, aff("subsystem", sub.ID, types.OpModAdd, value, nil), 0)
		}
	}
}

// resolveSkillTargets: i bersagli di un modificatore di skill (sorgente = il personaggio).
func resolveSkillTargets(mi *types.SdeModifierInfo, ctx *FitContext) []*ItemState {
	return ctx.TargetsForModifier(mi, ctx.Character, nil)
}
