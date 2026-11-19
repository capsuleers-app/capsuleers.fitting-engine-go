// Porting di src/constants.ts. GENERATO UNA VOLTA dal TS con scripts di conversione e poi mantenuto a mano:
// i valori sono doppi con eve-fit-engine (src/constants.ts): ogni modifica va fatta in entrambi i repo.
// Regola dei nomi: ATTR.MAX_VELOCITY -> AttrMaxVelocity, CATEGORY.SHIP -> CategoryShip, ...
package engine

import "github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"

// CATEGORY
const (
	CategoryShip            = 6
	CategoryModule          = 7
	CategoryCharge          = 8
	CategorySkill           = 16
	CategoryDrone           = 18
	CategoryImplant         = 20
	CategorySubsystem       = 32
	CategoryFighter         = 87
	CategoryStructureModule = 66
)

// SLOT_EFFECT_ID
const (
	SlotEffectLoPower     = 11
	SlotEffectHiPower     = 12
	SlotEffectMedPower    = 13
	SlotEffectRigSlot     = 2663
	SlotEffectSubsystem   = 3772
	SlotEffectServiceSlot = 6306
)

// ACTIVATION_EFFECT_ID
const (
	ActivationEffectOnlineForStructures  = 16
	ActivationEffectOnline               = 16
	ActivationEffectLaserTurret          = 10
	ActivationEffectProjectileTurret     = 34
	ActivationEffectHybridTurret         = 35
	ActivationEffectMissileLaunch        = 87
	ActivationEffectMissileLaunchDumb    = 4947
	ActivationEffectDroneDamageAmp       = 5379
	ActivationEffectShieldBoosting       = 4
	ActivationEffectShieldBoostingFueled = 4936
	ActivationEffectArmorRepair          = 27
	ActivationEffectArmorRepairFueled    = 5275
	ActivationEffectHullRepair           = 26
	ActivationEffectAbThrust             = 6
	ActivationEffectMwdThrust            = 8
	ActivationEffectShieldTransfer       = 18
	ActivationEffectArmorTransfer        = 592
	ActivationEffectRemoteCapTransfer    = 31
	ActivationEffectNosferatu            = 1
	ActivationEffectWebTargeted          = 14
	ActivationEffectWarpScrambler        = 19
	ActivationEffectSensorDamp           = 1130
	ActivationEffectTrackingDisruptor    = 1799
	ActivationEffectEcm                  = 1786
	ActivationEffectEnergyNeutralize     = 28
)

// ATTR
const (
	AttrMass                      = 4
	AttrHp                        = 9
	AttrAgility                   = 70
	AttrVolume                    = 161
	AttrCapacity                  = 38
	AttrPowerOutput               = 11
	AttrPowerUsed                 = 30
	AttrCpuOutput                 = 48
	AttrCpuUsed                   = 50
	AttrUpgradeCapacity           = 1132
	AttrUpgradeCost               = 1153
	AttrDroneBandwidth            = 1271
	AttrDroneBandwidthUsed        = 1272
	AttrDroneCapacity             = 283
	AttrHiSlots                   = 14
	AttrMedSlots                  = 13
	AttrLowSlots                  = 12
	AttrRigSlots                  = 1137
	AttrSubsystemSlots            = 1367
	AttrServiceSlots              = 2056
	AttrLauncherHardpoints        = 101
	AttrTurretHardpoints          = 102
	AttrRigSize                   = 1547
	AttrMaxGroupFitted            = 1544
	AttrMaxTypeFitted             = 2487
	AttrCapacitorCapacity         = 482
	AttrCapacitorRechargeRate     = 55
	AttrShieldCapacity            = 263
	AttrShieldRechargeRate        = 479
	AttrShieldEmRes               = 271
	AttrShieldThermalRes          = 274
	AttrShieldKineticRes          = 273
	AttrShieldExplosiveRes        = 272
	AttrArmorHp                   = 265
	AttrArmorEmRes                = 267
	AttrArmorThermalRes           = 270
	AttrArmorKineticRes           = 269
	AttrArmorExplosiveRes         = 268
	AttrStructureEmRes            = 113
	AttrStructureThermalRes       = 110
	AttrStructureKineticRes       = 109
	AttrStructureExplosiveRes     = 111
	AttrMaxTargetRange            = 76
	AttrMaxTargetRangeCap         = 797
	AttrMaxLockedTargets          = 192
	AttrSignatureRadius           = 552
	AttrScanResolution            = 564
	AttrScanRadarStrength         = 208
	AttrScanLadarStrength         = 209
	AttrScanMagnetometricStrength = 210
	AttrScanGravimetricStrength   = 211
	AttrDroneControlRange         = 458
	AttrMaxVelocity               = 37
	AttrWarpSpeedMultiplier       = 600
	AttrDamageMultiplier          = 64
	AttrEmDamage                  = 114
	AttrThermalDamage             = 118
	AttrKineticDamage             = 117
	AttrExplosiveDamage           = 116
	AttrOptimalRange              = 54
	AttrFalloffRange              = 158
	AttrTrackingSpeed             = 160
	AttrRateOfFire                = 51
	AttrDamageDuration            = 73
	AttrMissileDamageMultiplier   = 212
	AttrExplosionVelocity         = 653
	AttrExplosionRadius           = 654
	AttrDrf                       = 858
	AttrExplosionDelay            = 281
	AttrDotDuration               = 5735
	AttrDotMaxDamagePerTick       = 5736
	AttrDotMaxHpPercentagePerTick = 5737
	AttrChargeGroup1              = 604
	AttrChargeGroup2              = 605
	AttrChargeGroup3              = 606
	AttrChargeGroup4              = 609
	AttrChargeGroup5              = 610
	AttrChargeSize                = 128
	AttrRequiredSkill1            = 182
	AttrRequiredSkill2            = 183
	AttrRequiredSkill3            = 184
	AttrRequiredSkill4            = 1285
	AttrRequiredSkill5            = 1289
	AttrRequiredSkill6            = 1290
	AttrRequiredSkill1Level       = 277
	AttrRequiredSkill2Level       = 278
	AttrRequiredSkill3Level       = 279
	AttrRequiredSkill4Level       = 1286
	AttrRequiredSkill5Level       = 1287
	AttrRequiredSkill6Level       = 1288
	AttrMaxActiveDrones           = 352
	AttrMaxVelocityLimit          = 192
	AttrMaxRangeLimit             = 192
	AttrCanFitShipGroup1          = 1298
	AttrCanFitShipGroup2          = 1299
	AttrCanFitShipGroup3          = 1300
	AttrCanFitShipGroup4          = 1301
	AttrCanFitShipGroup5          = 1872
	AttrCanFitShipGroup6          = 1879
	AttrCanFitShipGroup7          = 1880
	AttrCanFitShipGroup8          = 1881
	AttrCanFitShipGroup9          = 2065
	AttrCanFitShipType1           = 1302
	AttrCanFitShipType2           = 1303
	AttrCanFitShipType3           = 1304
	AttrCanFitShipType4           = 1305
)

var CanFitShipGroupAttrs = []int{1298, 1299, 1300, 1301, 1872, 1879, 1880, 1881, 2065}

var CanFitShipTypeAttrs = []int{1302, 1303, 1304, 1305}

var ChargeGroupAttrs = []int{AttrChargeGroup1, AttrChargeGroup2, AttrChargeGroup3, AttrChargeGroup4, AttrChargeGroup5}

// RequiredSkillPairs: (attributo skill, attributo livello).
var RequiredSkillPairs = [6][2]int{
	{AttrRequiredSkill1, AttrRequiredSkill1Level},
	{AttrRequiredSkill2, AttrRequiredSkill2Level},
	{AttrRequiredSkill3, AttrRequiredSkill3Level},
	{AttrRequiredSkill4, AttrRequiredSkill4Level},
	{AttrRequiredSkill5, AttrRequiredSkill5Level},
	{AttrRequiredSkill6, AttrRequiredSkill6Level},
}

// StackingPenaltyK e' STACKING_PENALTY_K.
const StackingPenaltyK = 7.1289

// OperationBySdeCode e' OPERATION_BY_SDE_CODE (ok=false se il codice non e' mappato: undefined in TS).
func OperationBySdeCode(code int) (types.ModifierOperation, bool) {
	switch code {
	case -1:
		return types.OpPreAssign, true
	case 0:
		return types.OpPreMul, true
	case 1:
		return types.OpPreDiv, true
	case 2:
		return types.OpModAdd, true
	case 3:
		return types.OpModSub, true
	case 4:
		return types.OpPostMul, true
	case 5:
		return types.OpPostDiv, true
	case 6:
		return types.OpPostPercent, true
	case 7:
		return types.OpPostAssign, true
	}
	return "", false
}

// SlotEffectToSlotType e' SLOT_EFFECT_TO_SLOT_TYPE (solo lookup per chiave).
func SlotEffectToSlotType(effectID int) (types.SlotType, bool) {
	switch effectID {
	case SlotEffectLoPower:
		return types.SlotLO, true
	case SlotEffectHiPower:
		return types.SlotHI, true
	case SlotEffectMedPower:
		return types.SlotMED, true
	case SlotEffectRigSlot:
		return types.SlotRIG, true
	case SlotEffectSubsystem:
		return types.SlotSUBSYSTEM, true
	case SlotEffectServiceSlot:
		return types.SlotSERVICE, true
	}
	return "", false
}

// WeaponEffectKind e' WEAPON_EFFECT_KIND (solo lookup per chiave).
type WeaponEffectKind string

const (
	WeaponTurret    WeaponEffectKind = "TURRET"
	WeaponMissile   WeaponEffectKind = "MISSILE"
	WeaponSmartbomb WeaponEffectKind = "SMARTBOMB"
	WeaponDoomsday  WeaponEffectKind = "DOOMSDAY"
)

func WeaponKindOf(effectID int) (WeaponEffectKind, bool) {
	switch effectID {
	case 10:
		return WeaponTurret, true
	case 34:
		return WeaponTurret, true
	case 6995:
		return WeaponTurret, true
	case 8037:
		return WeaponTurret, true
	case 9:
		return WeaponMissile, true
	case 101:
		return WeaponMissile, true
	case 38:
		return WeaponSmartbomb, true
	case 6447:
		return WeaponSmartbomb, true
	}
	return "", false
}

// RepairInfo: voce di REPAIR_EFFECT_AMOUNT_ATTR.
type RepairInfo struct {
	AmountAttr int
	Layer      string
}

func RepairEffectAmountAttr(effectID int) (RepairInfo, bool) {
	switch effectID {
	case ActivationEffectShieldBoosting:
		return RepairInfo{68, "SHIELD"}, true
	case ActivationEffectShieldBoostingFueled:
		return RepairInfo{68, "SHIELD"}, true
	case ActivationEffectArmorRepair:
		return RepairInfo{84, "ARMOR"}, true
	case ActivationEffectArmorRepairFueled:
		return RepairInfo{84, "ARMOR"}, true
	case ActivationEffectHullRepair:
		return RepairInfo{83, "HULL"}, true
	}
	return RepairInfo{}, false
}

// LegacyEffectEntry: id effetto SDE, nome atteso (nil = non esportato) e descrizione del gestore.
type LegacyEffectEntry struct {
	ID      int
	Name    *string
	Handler string
}

func strp(s string) *string { return &s }

// LEGACY_EFFECT_IDS (ordinata come nel TS).
var LegacyEffectIDs = []LegacyEffectEntry{
	{4, nil, "derived/tank::SHIELD_BOOSTING (active rep dispatch)"},
	{26, nil, "derived/tank::HULL_REPAIR (structureRepair dispatch)"},
	{27, nil, "derived/tank::ARMOR_REPAIR (active rep dispatch)"},
	{48, nil, "derived/capacitor::boosterDrainEntry"},
	{1730, nil, "modifierEngine::applyLegacyDroneDmgBonus"},
	{1851, nil, "modifierEngine::applyLegacyMissileSpecRof"},
	{660, nil, "modifierEngine::applyLegacyMissileChargeDmg (em)"},
	{661, nil, "modifierEngine::applyLegacyMissileChargeDmg (explosive)"},
	{662, nil, "modifierEngine::applyLegacyMissileChargeDmg (thermal)"},
	{668, nil, "modifierEngine::applyLegacyMissileChargeDmg (kinetic)"},
	{3380, nil, "modifierEngine::applyLegacyHicBubble"},
	{3771, nil, "modifierEngine::applyLegacySubsystemAddPassive (armorHP)"},
	{3808, nil, "modifierEngine::applyLegacySubsystemAddPassive (signatureRadius)"},
	{3810, nil, "modifierEngine::applyLegacySubsystemAddPassive (capacity)"},
	{3811, nil, "modifierEngine::applyLegacySubsystemAddPassive (capacitorCapacity)"},
	{3831, nil, "modifierEngine::applyLegacySubsystemAddPassive (shieldCapacity)"},
	{4489, nil, "modifierEngine::applyLegacyDoomsdaySelfEffects (legacy DD)"},
	{4490, nil, "modifierEngine::applyLegacyDoomsdaySelfEffects (legacy DD)"},
	{4491, nil, "modifierEngine::applyLegacyDoomsdaySelfEffects (legacy DD)"},
	{4492, nil, "modifierEngine::applyLegacyDoomsdaySelfEffects (legacy DD)"},
	{4928, nil, "modifierEngine::applyLegacyRAH"},
	{4936, nil, "derived/tank::SHIELD_BOOSTING_FUELED"},
	{4921, nil, "modifierEngine::applyLegacyMjdSigBloom (Battleship MJD)"},
	{5275, nil, "derived/tank::ARMOR_REPAIR_FUELED"},
	{6063, nil, "modifierEngine::applyLegacyEntosisLink"},
	{6184, nil, "modifierEngine::buildCapWarfareReport (REMOTE_CAP)"},
	{6185, nil, "modifierEngine::buildRemoteRepReport (HULL)"},
	{6186, nil, "modifierEngine::buildRemoteRepReport (SHIELD)"},
	{6187, nil, "modifierEngine::buildCapWarfareReport (NEUT)"},
	{6188, nil, "modifierEngine::buildRemoteRepReport (ARMOR)"},
	{6197, nil, "modifierEngine::buildCapWarfareReport (NOS)"},
	{6201, nil, "modifierEngine::applyLegacyDoomsdaySelfEffects (Reaper slash)"},
	{6208, nil, "modifierEngine::applyLegacyMjdSigBloom (Cruiser MJD)"},
	{6216, nil, "modifierEngine::buildCapWarfareReport (Structure NEUT)"},
	{6422, nil, "modifierEngine::PROJECTION_EWAR_DEFS (sensor damp)"},
	{6423, nil, "modifierEngine::PROJECTION_EWAR_DEFS (guidance disruptor)"},
	{6424, nil, "modifierEngine::PROJECTION_EWAR_DEFS (tracking disruptor)"},
	{6425, nil, "modifierEngine::PROJECTION_EWAR_DEFS (target painter)"},
	{6426, nil, "modifierEngine::PROJECTION_EWAR_DEFS (web/grappler)"},
	{6427, nil, "modifierEngine::PROJECTION_EWAR_DEFS (remote sensor booster, friendly)"},
	{6428, nil, "modifierEngine::PROJECTION_EWAR_DEFS (remote tracking computer, friendly)"},
	{6439, nil, "modifierEngine::applyLegacyFighterAbilities (Evasive Maneuvers)"},
	{6440, nil, "modifierEngine::applyLegacyFighterAbilities (Afterburner)"},
	{6441, nil, "modifierEngine::applyLegacyFighterAbilities (MicroWarpDrive)"},
	{6431, nil, "derived/offense::fighterContributionFor (missile attack family)"},
	{6465, nil, "derived/offense::fighterContributionFor (turret attack family)"},
	{6434, nil, "modifierEngine::applyLegacyFighterProjection (Energy Neutralizer)"},
	{6435, nil, "modifierEngine::applyLegacyFighterProjection (Stasis Webifier)"},
	{6436, nil, "modifierEngine::collectEffectStoppers (Warp Disruption — stops MWD/MJD)"},
	{6464, nil, "modifierEngine::collectEffectStoppers (Tackle — stops MWD/MJD)"},
	{6470, nil, "modifierEngine::buildEcmProjectionReport"},
	{6476, nil, "modifierEngine::PROJECTION_EWAR_DEFS (AoE web)"},
	{6477, nil, "modifierEngine::buildCapWarfareReport (Burst Projector NEUT)"},
	{6478, nil, "modifierEngine::PROJECTION_EWAR_DEFS (AoE paint)"},
	{6479, nil, "modifierEngine::PROJECTION_EWAR_DEFS (AoE track)"},
	{6481, nil, "modifierEngine::PROJECTION_EWAR_DEFS (AoE damp)"},
	{6472, nil, "modifierEngine::applyLegacyDoomsdaySelfEffects (DD beam)"},
	{6473, nil, "modifierEngine::applyLegacyDoomsdaySelfEffects (DD cone)"},
	{6474, nil, "modifierEngine::applyLegacyDoomsdaySelfEffects (DD HOG)"},
	{6484, nil, "modifierEngine::applyLegacyCapitalEhe"},
	{6651, nil, "modifierEngine::buildRemoteRepReport (Remote AAR)"},
	{6652, nil, "modifierEngine::buildRemoteRepReport (Remote ASB)"},
	{6658, nil, "modifierEngine::applyLegacyBastion"},
	{6682, nil, "modifierEngine::PROJECTION_EWAR_DEFS (Standup web)"},
	{6683, nil, "modifierEngine::PROJECTION_EWAR_DEFS (Standup paint)"},
	{6684, nil, "modifierEngine::PROJECTION_EWAR_DEFS (Standup damp)"},
	{6685, nil, "modifierEngine::buildEcmProjectionReport (Standup ECM)"},
	{6686, nil, "modifierEngine::PROJECTION_EWAR_DEFS (Standup track / guidance)"},
	{6714, nil, "modifierEngine::buildEcmProjectionReport (ECM Burst)"},
	{6730, nil, "modifierEngine::applyLegacyPropMods (MWD)"},
	{6731, nil, "modifierEngine::applyLegacyPropMods (AB)"},
	{6732, nil, "modifierEngine::applyLegacyCommandBursts"},
	{6733, nil, "modifierEngine::applyLegacyCommandBursts"},
	{6734, nil, "modifierEngine::applyLegacyCommandBursts"},
	{6735, nil, "modifierEngine::applyLegacyCommandBursts"},
	{6736, nil, "modifierEngine::applyLegacyCommandBursts"},
	{6995, nil, "modifierEngine::applyLegacyDisintegratorSpool"},
	{7166, nil, "modifierEngine::applyLegacyMutadaptiveSpool"},
	{8037, nil, "derived/offense::Vorton ChainLightning (ship)"},
	{6447, nil, "derived/offense::Vorton ChainLightning (Standup XL)"},
	{4728, nil, "modifierEngine::applyLegacySystemEffect"},
	{3773, nil, "modifierEngine::applyLegacySubsystemSlots (marker)"},
	{3774, nil, "modifierEngine::applyLegacySubsystemSlots (marker)"},
	{12174, nil, "derived/offense::breacherContributionFor"},
	{12126, nil, "modifierEngine::applyLegacyMjdSigBloom (Frigate MJD)"},
}

// OUT_OF_SCOPE_EFFECT_IDS (ordinata come nel TS).
var OutOfScopeEffectIDs = []LegacyEffectEntry{
	{40, nil, "fitChecks::isMissileLauncher (hardpoint marker)"},
	{42, nil, "fitChecks::isTurretWeapon (hardpoint marker, also on miners)"},
	{263, nil, "fitChecks::reload-cycle marker (barrage)"},
	{54, nil, "OUT-OF-SCOPE :: targetPassively (passive lock, no stat impact)"},
	{55, nil, "OUT-OF-SCOPE :: targetHostiles (target marker)"},
	{46, nil, "OUT-OF-SCOPE :: shipScan (utility, no ship stat)"},
	{47, nil, "OUT-OF-SCOPE :: cargoScan (utility, no ship stat)"},
	{1738, nil, "OUT-OF-SCOPE :: doHacking (data/relic analyzers, no stat impact)"},
	{2255, nil, "OUT-OF-SCOPE :: tractorBeamCan (utility)"},
	{3793, nil, "OUT-OF-SCOPE :: probeLaunching (utility)"},
	{2757, nil, "OUT-OF-SCOPE :: salvaging (utility)"},
	{5163, nil, "OUT-OF-SCOPE :: salvageDroneEffect (drone utility)"},
	{8093, nil, "OUT-OF-SCOPE :: cloneRespawnBay (no stat impact)"},
	{8364, nil, "OUT-OF-SCOPE :: industrialItemCompression (no stat impact)"},
	{6719, nil, "OUT-OF-SCOPE :: moduleBonusIndustrialInvulnerability (industrial-only state)"},
	{848, nil, "OUT-OF-SCOPE :: cloakingTargetingDelayBonus (UI hint, no stat)"},
	{2413, nil, "OUT-OF-SCOPE :: snowBallLaunching (firework charges)"},
	{103, nil, "OUT-OF-SCOPE :: defenderMissileLaunching (missile defender, no stat)"},
	{127, nil, "OUT-OF-SCOPE :: torpedoLaunching (charge marker; launchers use 101/9)"},
	{104, nil, "OUT-OF-SCOPE :: fofMissileLaunching (charge marker; launchers use 101)"},
	{2971, nil, "OUT-OF-SCOPE :: bombLaunching (bomb charge marker; launchers use 101)"},
	{17, nil, "OUT-OF-SCOPE :: mining (yield only; reach back if mining UI lands)"},
	{67, nil, "OUT-OF-SCOPE :: miningLaser (yield only)"},
	{2726, nil, "OUT-OF-SCOPE :: miningClouds (yield only)"},
	{6687, nil, "OUT-OF-SCOPE :: npcEntityRemoteArmorRepairer"},
	{6688, nil, "OUT-OF-SCOPE :: npcEntityRemoteShieldBooster"},
	{6689, nil, "OUT-OF-SCOPE :: npcEntityRemoteHullRepairer"},
	{6690, nil, "OUT-OF-SCOPE :: remoteWebifierEntity"},
	{6691, nil, "OUT-OF-SCOPE :: entityEnergyNeutralizerFalloff"},
	{6692, nil, "OUT-OF-SCOPE :: remoteTargetPaintEntity"},
	{6693, nil, "OUT-OF-SCOPE :: remoteSensorDampEntity"},
	{6694, nil, "OUT-OF-SCOPE :: npcEntityWeaponDisruptor"},
	{6695, nil, "OUT-OF-SCOPE :: entityECMFalloff"},
	{6437, nil, "DEFERRED :: fighterAbilityECM (projection — needs jam-strength attr mapping)"},
	{6442, nil, "OUT-OF-SCOPE :: fighterAbilityMicroJumpDrive (self-mod movement, no projection)"},
	{6485, nil, "DEFERRED :: fighterAbilityLaunchBomb (projected damage — needs projected-DPS surface)"},
	{6554, nil, "DEFERRED :: fighterAbilityKamikaze (projected damage — needs projected-DPS surface)"},
	{6482, nil, "OUT-OF-SCOPE :: doomsdayAOEBubble (projection)"},
	{6513, nil, "OUT-OF-SCOPE :: doomsdayAOEECM (projection)"},
	{11691, nil, "OUT-OF-SCOPE :: debuffLance (Lancer dread projection)"},
	{6753, nil, "OUT-OF-SCOPE :: moduleTitanEffectGenerator (projection)"},
	{6443, nil, "OUT-OF-SCOPE :: pointDefense (Standup anti-missile, projection-only)"},
	{7120, nil, "OUT-OF-SCOPE :: structureCynoJammerOnline (system-wide cyno suppression state, no fit-panel stat)"},
	{12916, nil, "OUT-OF-SCOPE :: moduleBonusBreacherPodDamageControl (breacher-pod DoT resist, no fit-panel stat)"},
}

// BoosterSideEffectIDs e' BOOSTER_SIDE_EFFECT_IDS (insieme: solo Has).
func IsBoosterSideEffectID(id int) bool {
	switch id {
	case 2735, 2736, 2737, 2739, 2741, 2745, 2746, 2747, 2748, 2749, 2791, 4970:
		return true
	}
	return false
}

// VerifyComplaint e' un elemento di verifyLegacyEffectIds.
type VerifyComplaint struct {
	ID       int
	Kind     string // "missing" | "name-mismatch"
	Expected *string
	Actual   *string
}

// VerifyLegacyEffectIDs e' verifyLegacyEffectIds. `effectName` non e' un campo del bundle
// (il campo e' `name`), come nel TS dove `eff.effectName` e' sempre undefined: il controllo
// "name-mismatch" non scatta mai, resta solo "missing".
func VerifyLegacyEffectIDs(has func(id int) bool) []VerifyComplaint {
	var out []VerifyComplaint
	for _, e := range LegacyEffectIDs {
		if !has(e.ID) {
			out = append(out, VerifyComplaint{ID: e.ID, Kind: "missing", Expected: e.Name})
		}
	}
	return out
}

// NamedConst e' una voce (nome TS, valore) degli oggetti di costanti, per il dump di verifica.
type NamedConst struct {
	Name  string
	Value int
}

// ConstantTables restituisce gli oggetti numerici di constants.ts nell'ordine del TS, per nome.
func ConstantTables() map[string][]NamedConst {
	return map[string][]NamedConst{
		"CATEGORY": {
			{"SHIP", CategoryShip},
			{"MODULE", CategoryModule},
			{"CHARGE", CategoryCharge},
			{"SKILL", CategorySkill},
			{"DRONE", CategoryDrone},
			{"IMPLANT", CategoryImplant},
			{"SUBSYSTEM", CategorySubsystem},
			{"FIGHTER", CategoryFighter},
			{"STRUCTURE_MODULE", CategoryStructureModule},
		},
		"SLOT_EFFECT_ID": {
			{"LO_POWER", SlotEffectLoPower},
			{"HI_POWER", SlotEffectHiPower},
			{"MED_POWER", SlotEffectMedPower},
			{"RIG_SLOT", SlotEffectRigSlot},
			{"SUBSYSTEM", SlotEffectSubsystem},
			{"SERVICE_SLOT", SlotEffectServiceSlot},
		},
		"ACTIVATION_EFFECT_ID": {
			{"ONLINE_FOR_STRUCTURES", ActivationEffectOnlineForStructures},
			{"ONLINE", ActivationEffectOnline},
			{"LASER_TURRET", ActivationEffectLaserTurret},
			{"PROJECTILE_TURRET", ActivationEffectProjectileTurret},
			{"HYBRID_TURRET", ActivationEffectHybridTurret},
			{"MISSILE_LAUNCH", ActivationEffectMissileLaunch},
			{"MISSILE_LAUNCH_DUMB", ActivationEffectMissileLaunchDumb},
			{"DRONE_DAMAGE_AMP", ActivationEffectDroneDamageAmp},
			{"SHIELD_BOOSTING", ActivationEffectShieldBoosting},
			{"SHIELD_BOOSTING_FUELED", ActivationEffectShieldBoostingFueled},
			{"ARMOR_REPAIR", ActivationEffectArmorRepair},
			{"ARMOR_REPAIR_FUELED", ActivationEffectArmorRepairFueled},
			{"HULL_REPAIR", ActivationEffectHullRepair},
			{"AB_THRUST", ActivationEffectAbThrust},
			{"MWD_THRUST", ActivationEffectMwdThrust},
			{"SHIELD_TRANSFER", ActivationEffectShieldTransfer},
			{"ARMOR_TRANSFER", ActivationEffectArmorTransfer},
			{"REMOTE_CAP_TRANSFER", ActivationEffectRemoteCapTransfer},
			{"NOSFERATU", ActivationEffectNosferatu},
			{"WEB_TARGETED", ActivationEffectWebTargeted},
			{"WARP_SCRAMBLER", ActivationEffectWarpScrambler},
			{"SENSOR_DAMP", ActivationEffectSensorDamp},
			{"TRACKING_DISRUPTOR", ActivationEffectTrackingDisruptor},
			{"ECM", ActivationEffectEcm},
			{"ENERGY_NEUTRALIZE", ActivationEffectEnergyNeutralize},
		},
		"ATTR": {
			{"MASS", AttrMass},
			{"HP", AttrHp},
			{"AGILITY", AttrAgility},
			{"VOLUME", AttrVolume},
			{"CAPACITY", AttrCapacity},
			{"POWER_OUTPUT", AttrPowerOutput},
			{"POWER_USED", AttrPowerUsed},
			{"CPU_OUTPUT", AttrCpuOutput},
			{"CPU_USED", AttrCpuUsed},
			{"UPGRADE_CAPACITY", AttrUpgradeCapacity},
			{"UPGRADE_COST", AttrUpgradeCost},
			{"DRONE_BANDWIDTH", AttrDroneBandwidth},
			{"DRONE_BANDWIDTH_USED", AttrDroneBandwidthUsed},
			{"DRONE_CAPACITY", AttrDroneCapacity},
			{"HI_SLOTS", AttrHiSlots},
			{"MED_SLOTS", AttrMedSlots},
			{"LOW_SLOTS", AttrLowSlots},
			{"RIG_SLOTS", AttrRigSlots},
			{"SUBSYSTEM_SLOTS", AttrSubsystemSlots},
			{"SERVICE_SLOTS", AttrServiceSlots},
			{"LAUNCHER_HARDPOINTS", AttrLauncherHardpoints},
			{"TURRET_HARDPOINTS", AttrTurretHardpoints},
			{"RIG_SIZE", AttrRigSize},
			{"MAX_GROUP_FITTED", AttrMaxGroupFitted},
			{"MAX_TYPE_FITTED", AttrMaxTypeFitted},
			{"CAPACITOR_CAPACITY", AttrCapacitorCapacity},
			{"CAPACITOR_RECHARGE_RATE", AttrCapacitorRechargeRate},
			{"SHIELD_CAPACITY", AttrShieldCapacity},
			{"SHIELD_RECHARGE_RATE", AttrShieldRechargeRate},
			{"SHIELD_EM_RES", AttrShieldEmRes},
			{"SHIELD_THERMAL_RES", AttrShieldThermalRes},
			{"SHIELD_KINETIC_RES", AttrShieldKineticRes},
			{"SHIELD_EXPLOSIVE_RES", AttrShieldExplosiveRes},
			{"ARMOR_HP", AttrArmorHp},
			{"ARMOR_EM_RES", AttrArmorEmRes},
			{"ARMOR_THERMAL_RES", AttrArmorThermalRes},
			{"ARMOR_KINETIC_RES", AttrArmorKineticRes},
			{"ARMOR_EXPLOSIVE_RES", AttrArmorExplosiveRes},
			{"STRUCTURE_EM_RES", AttrStructureEmRes},
			{"STRUCTURE_THERMAL_RES", AttrStructureThermalRes},
			{"STRUCTURE_KINETIC_RES", AttrStructureKineticRes},
			{"STRUCTURE_EXPLOSIVE_RES", AttrStructureExplosiveRes},
			{"MAX_TARGET_RANGE", AttrMaxTargetRange},
			{"MAX_TARGET_RANGE_CAP", AttrMaxTargetRangeCap},
			{"MAX_LOCKED_TARGETS", AttrMaxLockedTargets},
			{"SIGNATURE_RADIUS", AttrSignatureRadius},
			{"SCAN_RESOLUTION", AttrScanResolution},
			{"SCAN_RADAR_STRENGTH", AttrScanRadarStrength},
			{"SCAN_LADAR_STRENGTH", AttrScanLadarStrength},
			{"SCAN_MAGNETOMETRIC_STRENGTH", AttrScanMagnetometricStrength},
			{"SCAN_GRAVIMETRIC_STRENGTH", AttrScanGravimetricStrength},
			{"DRONE_CONTROL_RANGE", AttrDroneControlRange},
			{"MAX_VELOCITY", AttrMaxVelocity},
			{"WARP_SPEED_MULTIPLIER", AttrWarpSpeedMultiplier},
			{"DAMAGE_MULTIPLIER", AttrDamageMultiplier},
			{"EM_DAMAGE", AttrEmDamage},
			{"THERMAL_DAMAGE", AttrThermalDamage},
			{"KINETIC_DAMAGE", AttrKineticDamage},
			{"EXPLOSIVE_DAMAGE", AttrExplosiveDamage},
			{"OPTIMAL_RANGE", AttrOptimalRange},
			{"FALLOFF_RANGE", AttrFalloffRange},
			{"TRACKING_SPEED", AttrTrackingSpeed},
			{"RATE_OF_FIRE", AttrRateOfFire},
			{"DAMAGE_DURATION", AttrDamageDuration},
			{"MISSILE_DAMAGE_MULTIPLIER", AttrMissileDamageMultiplier},
			{"EXPLOSION_VELOCITY", AttrExplosionVelocity},
			{"EXPLOSION_RADIUS", AttrExplosionRadius},
			{"DRF", AttrDrf},
			{"EXPLOSION_DELAY", AttrExplosionDelay},
			{"DOT_DURATION", AttrDotDuration},
			{"DOT_MAX_DAMAGE_PER_TICK", AttrDotMaxDamagePerTick},
			{"DOT_MAX_HP_PERCENTAGE_PER_TICK", AttrDotMaxHpPercentagePerTick},
			{"CHARGE_GROUP_1", AttrChargeGroup1},
			{"CHARGE_GROUP_2", AttrChargeGroup2},
			{"CHARGE_GROUP_3", AttrChargeGroup3},
			{"CHARGE_GROUP_4", AttrChargeGroup4},
			{"CHARGE_GROUP_5", AttrChargeGroup5},
			{"CHARGE_SIZE", AttrChargeSize},
			{"REQUIRED_SKILL_1", AttrRequiredSkill1},
			{"REQUIRED_SKILL_2", AttrRequiredSkill2},
			{"REQUIRED_SKILL_3", AttrRequiredSkill3},
			{"REQUIRED_SKILL_4", AttrRequiredSkill4},
			{"REQUIRED_SKILL_5", AttrRequiredSkill5},
			{"REQUIRED_SKILL_6", AttrRequiredSkill6},
			{"REQUIRED_SKILL_1_LEVEL", AttrRequiredSkill1Level},
			{"REQUIRED_SKILL_2_LEVEL", AttrRequiredSkill2Level},
			{"REQUIRED_SKILL_3_LEVEL", AttrRequiredSkill3Level},
			{"REQUIRED_SKILL_4_LEVEL", AttrRequiredSkill4Level},
			{"REQUIRED_SKILL_5_LEVEL", AttrRequiredSkill5Level},
			{"REQUIRED_SKILL_6_LEVEL", AttrRequiredSkill6Level},
			{"MAX_ACTIVE_DRONES", AttrMaxActiveDrones},
			{"MAX_VELOCITY_LIMIT", AttrMaxVelocityLimit},
			{"MAX_RANGE_LIMIT", AttrMaxRangeLimit},
			{"CAN_FIT_SHIP_GROUP_1", AttrCanFitShipGroup1},
			{"CAN_FIT_SHIP_GROUP_2", AttrCanFitShipGroup2},
			{"CAN_FIT_SHIP_GROUP_3", AttrCanFitShipGroup3},
			{"CAN_FIT_SHIP_GROUP_4", AttrCanFitShipGroup4},
			{"CAN_FIT_SHIP_GROUP_5", AttrCanFitShipGroup5},
			{"CAN_FIT_SHIP_GROUP_6", AttrCanFitShipGroup6},
			{"CAN_FIT_SHIP_GROUP_7", AttrCanFitShipGroup7},
			{"CAN_FIT_SHIP_GROUP_8", AttrCanFitShipGroup8},
			{"CAN_FIT_SHIP_GROUP_9", AttrCanFitShipGroup9},
			{"CAN_FIT_SHIP_TYPE_1", AttrCanFitShipType1},
			{"CAN_FIT_SHIP_TYPE_2", AttrCanFitShipType2},
			{"CAN_FIT_SHIP_TYPE_3", AttrCanFitShipType3},
			{"CAN_FIT_SHIP_TYPE_4", AttrCanFitShipType4},
		},
	}
}
