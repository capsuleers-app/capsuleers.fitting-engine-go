// GENERATO UNA VOLTA da src/modifierEngine.ts (tabelle numeriche di primo livello) e poi mantenuto a mano.
// Tabella doppia con eve-fit-engine (src/modifierEngine.ts): ogni modifica va fatta in entrambi i repo; la
// deriva la scoprono solo la parita' con pyfa (test/parity, tools/diff), nessun confronto automatico.
package engine

var percentUnitIdsEntries = []int{105, 109, 111, 121, 124, 127}
var percentUnitIds = newIntSet(percentUnitIdsEntries)

var RuntimeEarliestEffectIdsEntries = []int{12963, 12964, 12965, 12966, 12989, 12990, 12991, 12992, 12993}
var RuntimeEarliestEffectIds = newIntSet(RuntimeEarliestEffectIdsEntries)

const attrFighterWebSpeedPen = 2184

const attrFighterNeutDuration = 2208

const attrFighterNeutAmount = 2211

const attrMaxVel = 37

var shipBonusScalingSkillEntries = [][2]int{{2359, 24311}, {2360, 24311}, {2361, 24311}, {2362, 24311}, {5981, 24311}, {2363, 24312}, {2364, 24312}, {2365, 24312}, {2366, 24312}, {5982, 24312}, {2367, 24313}, {2368, 24313}, {2369, 24313}, {2370, 24313}, {5983, 24313}, {2371, 24314}, {2372, 24314}, {2373, 24314}, {2374, 24314}, {5984, 24314}, {2308, 24311}, {2309, 24311}, {2310, 24311}, {2320, 24311}, {6114, 24311}, {2311, 24312}, {2312, 24312}, {2313, 24312}, {2321, 24312}, {6113, 24312}, {2314, 24313}, {2315, 24313}, {2316, 24313}, {2322, 24313}, {6112, 24313}, {2317, 24314}, {2318, 24314}, {2319, 24314}, {2323, 24314}, {6116, 24314}, {2375, 24311}, {2376, 24311}, {2377, 24311}, {2378, 24311}, {2379, 24311}, {2380, 24312}, {2381, 24312}, {2382, 24312}, {2383, 24312}, {2384, 24312}, {2385, 24313}, {2386, 24313}, {2387, 24313}, {2388, 24313}, {2389, 24313}, {2390, 24314}, {2391, 24314}, {2392, 24314}, {2393, 24314}, {2394, 24314}, {2283, 20525}, {2284, 20525}, {2285, 20525}, {5214, 20525}, {2286, 20530}, {2287, 20530}, {2288, 20530}, {2289, 20531}, {2290, 20531}, {2291, 20531}, {5215, 20531}, {5216, 20531}, {2292, 20532}, {2293, 20532}, {2294, 20532}, {5248, 20532}, {5417, 77738}, {5418, 77738}, {2829, 52997}, {2830, 52997}, {2831, 52997}, {795, 89609}, {1889, 89609}, {5205, 89609}, {2406, 3347}, {2407, 3347}, {2408, 3347}, {2409, 3347}, {2410, 3346}, {2411, 3346}, {2412, 3346}, {2413, 3346}, {2423, 3346}, {2414, 3344}, {2415, 3344}, {2416, 3344}, {2417, 3344}, {2418, 3345}, {2419, 3345}, {2420, 3345}, {2421, 3345}, {462, 3328}, {586, 3328}, {1625, 3328}, {460, 3329}, {587, 3329}, {1626, 3329}, {463, 3330}, {588, 3330}, {1624, 3330}, {464, 3331}, {485, 3331}, {1623, 3331}, {486, 3332}, {658, 3332}, {2014, 3332}, {489, 3333}, {659, 3333}, {5747, 3333}, {487, 3334}, {657, 3334}, {1535, 3334}, {478, 3335}, {656, 3335}, {2070, 3335}, {6476, 3335}, {500, 3336}, {561, 3336}, {5240, 3336}, {490, 3337}, {518, 3337}, {491, 3338}, {501, 3338}, {598, 3338}, {585, 3339}, {5960, 3339}, {678, 12096}, {679, 12096}, {2092, 40328}, {2093, 40328}, {493, 3341}, {814, 3341}, {3157, 3340}, {3241, 3340}, {494, 3343}, {809, 3343}, {495, 3342}, {811, 3342}, {496, 3340}, {813, 3340}, {1887, 33091}, {1888, 33091}, {5218, 33091}, {5219, 33091}, {5220, 33091}, {5221, 33091}, {5222, 33091}, {734, 33092}, {735, 33092}, {5225, 33092}, {5228, 33092}, {5229, 33092}, {738, 33093}, {739, 33093}, {5230, 33093}, {5232, 33093}, {5233, 33093}, {729, 33094}, {740, 33094}, {5235, 33094}, {5236, 33094}, {5237, 33094}, {743, 33096}, {745, 33096}, {5044, 33096}, {746, 33097}, {747, 33097}, {5046, 33097}, {748, 33098}, {749, 33098}, {5207, 33098}, {568, 12092}, {804, 12092}, {569, 12093}, {839, 12093}, {1578, 12093}, {2731, 12093}, {673, 12095}, {675, 12095}, {692, 16591}, {693, 16591}, {807, 19719}, {808, 19719}, {1360, 19719}, {1361, 19719}, {962, 22761}, {963, 22761}, {1537, 22761}, {999, 23950}, {1000, 23950}, {1924, 23950}, {5772, 23950}, {1012, 12098}, {1013, 12098}, {1246, 28609}, {1247, 28609}, {1249, 28615}, {1250, 28615}, {2069, 28615}, {1257, 28656}, {1258, 28656}, {2627, 28656}, {2628, 28656}, {1265, 28667}, {1266, 28667}, {5792, 12096}, {2460, 12096}, {886, 20524}, {887, 20524}, {888, 20526}, {889, 20526}, {890, 20527}, {891, 20527}, {892, 20528}, {893, 20528}, {1239, 28374}, {1240, 28374}, {1243, 28374}, {1244, 28374}, {2582, 28374}, {3233, 28374}, {1311, 29029}, {1312, 29029}, {1356, 29637}, {1358, 29637}, {2474, 29637}, {2475, 29637}, {2577, 29637}, {3211, 29637}, {3212, 29637}, {3235, 29637}, {1983, 34327}, {1984, 34327}, {5647, 81032}, {5648, 81032}, {5649, 81044}, {5650, 81044}, {5654, 81044}, {1986, 34390}, {1987, 34390}, {1988, 34390}, {2004, 34533}, {2005, 34533}, {2006, 34533}, {2015, 35680}, {2016, 35680}, {2017, 35680}, {2027, 35685}, {2028, 35685}, {2029, 35685}, {2059, 37615}, {2060, 37615}, {2061, 37615}, {1503, 30650}, {2677, 30650}, {1504, 30651}, {2676, 30651}, {1505, 30652}, {2678, 30652}, {1506, 30653}, {2679, 30653}, {1517, 30540}, {1519, 30546}, {1520, 30553}, {1521, 30550}, {1522, 30551}, {1523, 30554}, {1525, 30547}, {1526, 30545}, {1531, 30537}, {1532, 30550}, {1533, 30549}, {1534, 30551}, {2680, 30532}, {2681, 30539}, {2682, 30544}, {2683, 30548}, {2684, 30540}, {2685, 30546}, {2686, 30545}, {2687, 30547}, {2762, 47867}, {2763, 47867}, {2764, 47868}, {2765, 47868}, {2766, 47869}, {2767, 47869}, {2799, 49742}, {2800, 49742}, {2801, 49743}, {2802, 49743}, {3041, 55031}, {3042, 55031}, {3043, 55032}, {3044, 55032}, {3045, 54794}, {3046, 54794}, {3201, 22551}, {3202, 22551}, {3226, 22551}, {3213, 33856}, {3214, 33856}, {5820, 89241}, {2752, 47445}, {5681, 83094}, {5821, 89241}, {5822, 89241}, {5953, 89241}, {5955, 32918}, {5956, 33096}, {5939, 89609}, {5940, 89609}, {5941, 89609}, {5942, 89609}, {5943, 89609}, {5960, 3339}, {5981, 24311}, {5982, 24312}, {5983, 24313}, {5984, 24314}, {6112, 24313}, {6113, 24312}, {6114, 24311}, {6116, 24314}, {66, 89611}, {349, 19760}, {492, 3339}, {1296, 21610}, {1669, 3184}, {1670, 3184}, {1842, 32918}, {3167, 33856}, {3181, 17940}, {3182, 17940}, {3183, 17940}, {3184, 17940}, {3185, 17940}, {3187, 17940}, {3188, 17940}, {3190, 33856}, {3191, 33856}, {3192, 33856}, {3193, 22551}, {3194, 22551}, {3197, 22551}, {3198, 22551}, {3199, 22551}, {3203, 29637}, {3204, 29637}, {3205, 29637}, {3210, 3341}, {3221, 29637}, {3222, 29637}, {3223, 28374}, {3224, 28374}, {3237, 32918}, {3240, 32918}, {3326, 28374}, {6088, 33092}, {6089, 33094}}
var shipBonusScalingSkill = newIntMap(shipBonusScalingSkillEntries)

var subsystemBonusScalingSkillEntries = [][2]int{{1431, 30539}, {1509, 30539}, {1432, 30536}, {1508, 30536}, {1433, 30532}, {1507, 30532}, {1434, 30537}, {1511, 30537}, {1435, 30538}, {1512, 30538}, {1441, 30548}, {1515, 30548}, {1442, 30542}, {1514, 30542}, {1443, 30544}, {1516, 30544}, {1444, 30549}, {1510, 30549}, {1445, 30552}, {1513, 30552}, {1436, 30546}, {1437, 30541}, {1438, 30540}, {1439, 30550}, {1440, 30553}, {1446, 30547}, {1447, 30543}, {1448, 30545}, {1449, 30551}, {1450, 30554}, {1517, 30540}, {1519, 30546}, {1520, 30553}, {1521, 30550}, {1522, 30551}, {1523, 30554}, {1525, 30547}, {1526, 30545}, {1531, 30537}, {1532, 30550}, {1533, 30549}, {1534, 30551}, {2680, 30532}, {2681, 30539}, {2682, 30544}, {2683, 30548}, {2684, 30540}, {2685, 30546}, {2686, 30545}, {2687, 30547}}
var subsystemBonusScalingSkill = newIntMap(subsystemBonusScalingSkillEntries)

var shipRoleBonusAttrsEntries = []int{793, 1688, 1803, 2059, 2060, 2064, 2298, 2299, 2300, 2301, 2302, 2303, 5952, 1823, 1827, 1830, 1836, 1831, 1268, 1576, 2580, 3179, 5746, 5748, 1989}
var shipRoleBonusAttrs = newIntSet(shipRoleBonusAttrsEntries)

var legacyDroneDmgEffectIdsEntries = []int{1730}
var legacyDroneDmgEffectIds = newIntSet(legacyDroneDmgEffectIdsEntries)

var legacyPropmodEffectIdsEntries = []int{6730, 6731}
var legacyPropmodEffectIds = newIntSet(legacyPropmodEffectIdsEntries)

const attrDamageMultiplierBonus = 292

const attrDamageMultiplier = 64

const attrMaxVelocity = 37

const attrMass = 4

const attrSignatureRadius = 552

const attrSpeedFactor = 20

const attrSpeedBoostFactor = 567

const attrMassAddition = 796

const attrSignatureRadiusBonus = 554

const attrHullEmRes = 113

const attrHullThermRes = 110

const attrHullKinRes = 109

const attrHullExpRes = 111

const attrHullEheEmRes = 974

const attrHullEheThermRes = 977

const attrHullEheKinRes = 976

const attrHullEheExpRes = 975

const attrArmorEmRes = 267

const attrArmorExpRes = 268

const attrArmorKinRes = 269

const attrArmorThermRes = 270

const attrRahResistanceShift = 1849

const attrScanGravimetric = 208

const attrScanLadar = 209

const attrScanMagnetometric = 210

const attrScanRadar = 211

const attrMassBonusPercentage = 1131

const attrSignatureRadiusBonusLegacy = 554

const attrSignatureRadiusBonusPercent = 973

const attrSpeedFactorBonus = 1164

const attrSpeedBoostFactorBonus = 1270

const attrDisallowAssistance = 854

const attrRangeOptimal = 54

const attrRangeFalloff = 2044

const attrMaxTargetRange = 76

const attrMaxTargetRangeBonus = 309

const attrScanResolution = 564

const attrScanResolutionBonus = 566

const attrTrackingSpeed = 160

const attrTrackingSpeedBonus = 767

const attrFalloff = 158

const attrFalloffBonus = 349

const attrMaxRange = 54

const attrMaxRangeBonus = 351

const attrAoeCloudSize = 654

const attrAoeCloudSizeBonus = 848

const attrAoeVelocity = 653

const attrAoeVelocityBonus = 847

const attrExplosionDelay = 281

const attrExplosionDelayBonus = 596

const attrMissileVelocityBonus = 547

var attrWarfareBuffIds = [][2]int{{2468, 2469}, {2470, 2471}, {2472, 2473}, {2536, 2537}}

const propulsionModuleGroupId = 46

const skillGunnery = 3300

const skillMissileLauncherOp = 3319

var legacyMissileChargeDmgEffectsEntries = [][2]int{{660, 114}, {661, 116}, {668, 117}, {662, 118}}
var legacyMissileChargeDmgEffects = newIntMap(legacyMissileChargeDmgEffectsEntries)

const attrRofBonus = 293

const entosisLinkEffectId = 6063

const capitalEheEffectId = 6484

const hicBubbleEffectId = 3380

var legacyMjdEffectIdsEntries = []int{4921, 6208, 12126}
var legacyMjdEffectIds = newIntSet(legacyMjdEffectIdsEntries)

const attrDurationMs = 73

const attrShieldRepAmount = 68

const attrArmorRepAmount = 84

const attrHullRepAmount = 83

const attrPowerTransferAmount = 90

const attrNeutAmount = 97

const attrDisallowOffensive = 872

const attrScanRadarStr = 208

const attrScanLadarStr = 209

const attrScanMagnetometricStr = 210

const attrScanGravimetricStr = 211

const attrScanRadarBonus = 241

const attrScanLadarBonus = 239

const attrScanMagnetometricBonus = 240

const attrScanGravimetricBonus = 238

const attrSiegeWarpStatus = 852

const attrWarpScrambleStatus = 104

var projectionEcmEffectIdsEntries = []int{6470, 6685, 6714}
var projectionEcmEffectIds = newIntSet(projectionEcmEffectIdsEntries)

var overheatPairs = [][2]int{{1205, 51}, {1210, 64}, {1223, 20}, {1181, 73}, {1230, 84}, {1153, 68}}

const bastionEffectId = 6658

const skillLargeEnergyTurret = 3306

const skillLargeHybridTurret = 3308

const skillLargeProjectileTurret = 3304

const skillLargePrecursorWeapon = 47872

const skillTorpedoes = 3325

const skillCruiseMissiles = 3326

const skillHeavyMissiles = 3324

const skillTorpedoSpec = 20213

const skillCruiseMissileSpec = 20212

const skillRepairSystems = 3393

const skillShieldOperation = 3416

const skillMjdOperation = 4385

const attrShieldEmRes = 271

const attrShieldExpRes = 272

const attrShieldKinRes = 273

const attrShieldThermRes = 274

const attrShieldBoostMultiplier = 548

const attrSpeed = 51

const attrCapacitorNeed = 6

const attrActivationBlocked = 1349

const attrActivationBlockedStr = 1350

const attrRemoteRepairImpedanceBonus = 2342

const attrRemoteAssistImpedanceBonus = 2352

const attrSensorDampResistBonus = 2351

const attrWeaponDisruptResistBonus = 2353

const attrTargetPaintResistBonus = 2424

const attrRemoteRepairImpedance = 2116

const attrRemoteAssistImpedance = 2135

const attrSensorDampResist = 2112

const attrWeaponDisruptResist = 2113

const attrTargetPaintResist = 2114

const attrBastionRepairDurationBonus = 5964

const attrArmorDamageAmountBonus = 895

const mutadaptiveRarEffectId = 7166

const attrMutadaptiveBonusMax = 2767

const attrMutadaptiveBonusPerCycle = 2768

var legacyDoomsdayEffectIdsEntries = []int{6201, 6472, 6473, 6474, 4489, 4490, 4491, 4492}
var legacyDoomsdayEffectIds = newIntSet(legacyDoomsdayEffectIdsEntries)

var commandBurstEffectIdsEntries = []int{6732, 6733, 6734, 6735, 6736}
var commandBurstEffectIds = newIntSet(commandBurstEffectIdsEntries)

const smartBombGroupId = 72

var smallTurretSkillIdsEntries = []int{3300, 3303, 3315}
var smallTurretSkillIds = newIntSet(smallTurretSkillIdsEntries)

var smallMissileSkillIdsEntries = []int{3319, 3321}
var smallMissileSkillIds = newIntSet(smallMissileSkillIdsEntries)

var shieldBoosterEffectIdsEntries = []int{4, 4936}
var shieldBoosterEffectIds = newIntSet(shieldBoosterEffectIdsEntries)

var armorRepairerEffectIdsEntries = []int{27, 5275}
var armorRepairerEffectIds = newIntSet(armorRepairerEffectIdsEntries)

var remoteShieldXferEffectIdsEntries = []int{18, 6186, 6652}
var remoteShieldXferEffectIds = newIntSet(remoteShieldXferEffectIdsEntries)

var remoteArmorRepEffectIdsEntries = []int{592, 6188, 6651}
var remoteArmorRepEffectIds = newIntSet(remoteArmorRepEffectIdsEntries)

var energyNeutralizerEffectIdsEntries = []int{28, 6187, 6477, 6691}
var energyNeutralizerEffectIds = newIntSet(energyNeutralizerEffectIdsEntries)

var energyNosferatuEffectIdsEntries = []int{1, 6197}
var energyNosferatuEffectIds = newIntSet(energyNosferatuEffectIdsEntries)

var stasisWebEffectIdsEntries = []int{14, 6426, 6476, 6682}
var stasisWebEffectIds = newIntSet(stasisWebEffectIdsEntries)

var targetPainterEffectIdsEntries = []int{1019, 6425, 6478, 6683}
var targetPainterEffectIds = newIntSet(targetPainterEffectIdsEntries)

const fighterAbEffectId = 6440

const fighterMwdEffectId = 6441

const fighterEvasiveEffectId = 6439

const attrFighterAbSpeedBonus = 2151

const attrFighterMwdSpeedBonus = 2152

const attrFighterMwdSigBonus = 2153

const attrFighterEvasiveSpeedBonus = 2224

const attrFighterEvasiveSigBonus = 2225

const attrFighterEvasiveEmRes = 2118

const attrFighterEvasiveThmRes = 2119

const attrFighterEvasiveKinRes = 2120

const attrFighterEvasiveExpRes = 2121

const attrShieldEmReso = 271

const attrShieldExpReso = 272

const attrShieldKinReso = 273

const attrShieldThmReso = 274

const disintegratorEffectId = 6995

const attrDmgMultBonusPerCycle = 2733

const attrDmgMultBonusMax = 2734

const rahEffectId = 4928

var subsystemSlotBonusMap = [][2]int{{1368, 102}, {1369, 101}, {1374, 14}, {1375, 13}, {1376, 12}}

var legacyHandledHardcodedEffectsEntries = []int{6658}
var legacyHandledHardcodedEffects = newIntSet(legacyHandledHardcodedEffectsEntries)

var SecStatusScaledEffectIdsEntries = []int{6871, 12165, 12181, 12185, 12202}
var SecStatusScaledEffectIds = newIntSet(SecStatusScaledEffectIdsEntries)

// ModTable e' una tabella numerica di modifierEngine.ts, per il dump di verifica.
type ModTable struct {
	Name  string
	Kind  string // const | set | map | pairs
	Const int
	Ints  []int
	Pairs [][2]int
}

// ModifierEngineTables elenca le tabelle nell'ordine del sorgente TS.
func ModifierEngineTables() []ModTable {
	return []ModTable{
		{Name: "PERCENT_UNIT_IDS", Kind: "set", Ints: percentUnitIdsEntries},
		{Name: "RUNTIME_EARLIEST_EFFECT_IDS", Kind: "set", Ints: RuntimeEarliestEffectIdsEntries},
		{Name: "ATTR_FIGHTER_WEB_SPEED_PEN", Kind: "const", Const: attrFighterWebSpeedPen},
		{Name: "ATTR_FIGHTER_NEUT_DURATION", Kind: "const", Const: attrFighterNeutDuration},
		{Name: "ATTR_FIGHTER_NEUT_AMOUNT", Kind: "const", Const: attrFighterNeutAmount},
		{Name: "ATTR_MAX_VEL", Kind: "const", Const: attrMaxVel},
		{Name: "SHIP_BONUS_SCALING_SKILL", Kind: "map", Pairs: shipBonusScalingSkillEntries},
		{Name: "SUBSYSTEM_BONUS_SCALING_SKILL", Kind: "map", Pairs: subsystemBonusScalingSkillEntries},
		{Name: "SHIP_ROLE_BONUS_ATTRS", Kind: "set", Ints: shipRoleBonusAttrsEntries},
		{Name: "LEGACY_DRONE_DMG_EFFECT_IDS", Kind: "set", Ints: legacyDroneDmgEffectIdsEntries},
		{Name: "LEGACY_PROPMOD_EFFECT_IDS", Kind: "set", Ints: legacyPropmodEffectIdsEntries},
		{Name: "ATTR_DAMAGE_MULTIPLIER_BONUS", Kind: "const", Const: attrDamageMultiplierBonus},
		{Name: "ATTR_DAMAGE_MULTIPLIER", Kind: "const", Const: attrDamageMultiplier},
		{Name: "ATTR_MAX_VELOCITY", Kind: "const", Const: attrMaxVelocity},
		{Name: "ATTR_MASS", Kind: "const", Const: attrMass},
		{Name: "ATTR_SIGNATURE_RADIUS", Kind: "const", Const: attrSignatureRadius},
		{Name: "ATTR_SPEED_FACTOR", Kind: "const", Const: attrSpeedFactor},
		{Name: "ATTR_SPEED_BOOST_FACTOR", Kind: "const", Const: attrSpeedBoostFactor},
		{Name: "ATTR_MASS_ADDITION", Kind: "const", Const: attrMassAddition},
		{Name: "ATTR_SIGNATURE_RADIUS_BONUS", Kind: "const", Const: attrSignatureRadiusBonus},
		{Name: "ATTR_HULL_EM_RES", Kind: "const", Const: attrHullEmRes},
		{Name: "ATTR_HULL_THERM_RES", Kind: "const", Const: attrHullThermRes},
		{Name: "ATTR_HULL_KIN_RES", Kind: "const", Const: attrHullKinRes},
		{Name: "ATTR_HULL_EXP_RES", Kind: "const", Const: attrHullExpRes},
		{Name: "ATTR_HULL_EHE_EM_RES", Kind: "const", Const: attrHullEheEmRes},
		{Name: "ATTR_HULL_EHE_THERM_RES", Kind: "const", Const: attrHullEheThermRes},
		{Name: "ATTR_HULL_EHE_KIN_RES", Kind: "const", Const: attrHullEheKinRes},
		{Name: "ATTR_HULL_EHE_EXP_RES", Kind: "const", Const: attrHullEheExpRes},
		{Name: "ATTR_ARMOR_EM_RES", Kind: "const", Const: attrArmorEmRes},
		{Name: "ATTR_ARMOR_EXP_RES", Kind: "const", Const: attrArmorExpRes},
		{Name: "ATTR_ARMOR_KIN_RES", Kind: "const", Const: attrArmorKinRes},
		{Name: "ATTR_ARMOR_THERM_RES", Kind: "const", Const: attrArmorThermRes},
		{Name: "ATTR_RAH_RESISTANCE_SHIFT", Kind: "const", Const: attrRahResistanceShift},
		{Name: "ATTR_SCAN_GRAVIMETRIC", Kind: "const", Const: attrScanGravimetric},
		{Name: "ATTR_SCAN_LADAR", Kind: "const", Const: attrScanLadar},
		{Name: "ATTR_SCAN_MAGNETOMETRIC", Kind: "const", Const: attrScanMagnetometric},
		{Name: "ATTR_SCAN_RADAR", Kind: "const", Const: attrScanRadar},
		{Name: "ATTR_MASS_BONUS_PERCENTAGE", Kind: "const", Const: attrMassBonusPercentage},
		{Name: "ATTR_SIGNATURE_RADIUS_BONUS_LEGACY", Kind: "const", Const: attrSignatureRadiusBonusLegacy},
		{Name: "ATTR_SIGNATURE_RADIUS_BONUS_PERCENT", Kind: "const", Const: attrSignatureRadiusBonusPercent},
		{Name: "ATTR_SPEED_FACTOR_BONUS", Kind: "const", Const: attrSpeedFactorBonus},
		{Name: "ATTR_SPEED_BOOST_FACTOR_BONUS", Kind: "const", Const: attrSpeedBoostFactorBonus},
		{Name: "ATTR_DISALLOW_ASSISTANCE", Kind: "const", Const: attrDisallowAssistance},
		{Name: "ATTR_RANGE_OPTIMAL", Kind: "const", Const: attrRangeOptimal},
		{Name: "ATTR_RANGE_FALLOFF", Kind: "const", Const: attrRangeFalloff},
		{Name: "ATTR_MAX_TARGET_RANGE", Kind: "const", Const: attrMaxTargetRange},
		{Name: "ATTR_MAX_TARGET_RANGE_BONUS", Kind: "const", Const: attrMaxTargetRangeBonus},
		{Name: "ATTR_SCAN_RESOLUTION", Kind: "const", Const: attrScanResolution},
		{Name: "ATTR_SCAN_RESOLUTION_BONUS", Kind: "const", Const: attrScanResolutionBonus},
		{Name: "ATTR_TRACKING_SPEED", Kind: "const", Const: attrTrackingSpeed},
		{Name: "ATTR_TRACKING_SPEED_BONUS", Kind: "const", Const: attrTrackingSpeedBonus},
		{Name: "ATTR_FALLOFF", Kind: "const", Const: attrFalloff},
		{Name: "ATTR_FALLOFF_BONUS", Kind: "const", Const: attrFalloffBonus},
		{Name: "ATTR_MAX_RANGE", Kind: "const", Const: attrMaxRange},
		{Name: "ATTR_MAX_RANGE_BONUS", Kind: "const", Const: attrMaxRangeBonus},
		{Name: "ATTR_AOE_CLOUD_SIZE", Kind: "const", Const: attrAoeCloudSize},
		{Name: "ATTR_AOE_CLOUD_SIZE_BONUS", Kind: "const", Const: attrAoeCloudSizeBonus},
		{Name: "ATTR_AOE_VELOCITY", Kind: "const", Const: attrAoeVelocity},
		{Name: "ATTR_AOE_VELOCITY_BONUS", Kind: "const", Const: attrAoeVelocityBonus},
		{Name: "ATTR_EXPLOSION_DELAY", Kind: "const", Const: attrExplosionDelay},
		{Name: "ATTR_EXPLOSION_DELAY_BONUS", Kind: "const", Const: attrExplosionDelayBonus},
		{Name: "ATTR_MISSILE_VELOCITY_BONUS", Kind: "const", Const: attrMissileVelocityBonus},
		{Name: "ATTR_WARFARE_BUFF_IDS", Kind: "pairs", Pairs: attrWarfareBuffIds},
		{Name: "PROPULSION_MODULE_GROUP_ID", Kind: "const", Const: propulsionModuleGroupId},
		{Name: "SKILL_GUNNERY", Kind: "const", Const: skillGunnery},
		{Name: "SKILL_MISSILE_LAUNCHER_OP", Kind: "const", Const: skillMissileLauncherOp},
		{Name: "LEGACY_MISSILE_CHARGE_DMG_EFFECTS", Kind: "map", Pairs: legacyMissileChargeDmgEffectsEntries},
		{Name: "ATTR_ROF_BONUS", Kind: "const", Const: attrRofBonus},
		{Name: "ENTOSIS_LINK_EFFECT_ID", Kind: "const", Const: entosisLinkEffectId},
		{Name: "CAPITAL_EHE_EFFECT_ID", Kind: "const", Const: capitalEheEffectId},
		{Name: "HIC_BUBBLE_EFFECT_ID", Kind: "const", Const: hicBubbleEffectId},
		{Name: "LEGACY_MJD_EFFECT_IDS", Kind: "set", Ints: legacyMjdEffectIdsEntries},
		{Name: "ATTR_DURATION_MS", Kind: "const", Const: attrDurationMs},
		{Name: "ATTR_SHIELD_REP_AMOUNT", Kind: "const", Const: attrShieldRepAmount},
		{Name: "ATTR_ARMOR_REP_AMOUNT", Kind: "const", Const: attrArmorRepAmount},
		{Name: "ATTR_HULL_REP_AMOUNT", Kind: "const", Const: attrHullRepAmount},
		{Name: "ATTR_POWER_TRANSFER_AMOUNT", Kind: "const", Const: attrPowerTransferAmount},
		{Name: "ATTR_NEUT_AMOUNT", Kind: "const", Const: attrNeutAmount},
		{Name: "ATTR_DISALLOW_OFFENSIVE", Kind: "const", Const: attrDisallowOffensive},
		{Name: "ATTR_SCAN_RADAR_STR", Kind: "const", Const: attrScanRadarStr},
		{Name: "ATTR_SCAN_LADAR_STR", Kind: "const", Const: attrScanLadarStr},
		{Name: "ATTR_SCAN_MAGNETOMETRIC_STR", Kind: "const", Const: attrScanMagnetometricStr},
		{Name: "ATTR_SCAN_GRAVIMETRIC_STR", Kind: "const", Const: attrScanGravimetricStr},
		{Name: "ATTR_SCAN_RADAR_BONUS", Kind: "const", Const: attrScanRadarBonus},
		{Name: "ATTR_SCAN_LADAR_BONUS", Kind: "const", Const: attrScanLadarBonus},
		{Name: "ATTR_SCAN_MAGNETOMETRIC_BONUS", Kind: "const", Const: attrScanMagnetometricBonus},
		{Name: "ATTR_SCAN_GRAVIMETRIC_BONUS", Kind: "const", Const: attrScanGravimetricBonus},
		{Name: "ATTR_SIEGE_WARP_STATUS", Kind: "const", Const: attrSiegeWarpStatus},
		{Name: "ATTR_WARP_SCRAMBLE_STATUS", Kind: "const", Const: attrWarpScrambleStatus},
		{Name: "PROJECTION_ECM_EFFECT_IDS", Kind: "set", Ints: projectionEcmEffectIdsEntries},
		{Name: "OVERHEAT_PAIRS", Kind: "pairs", Pairs: overheatPairs},
		{Name: "BASTION_EFFECT_ID", Kind: "const", Const: bastionEffectId},
		{Name: "SKILL_LARGE_ENERGY_TURRET", Kind: "const", Const: skillLargeEnergyTurret},
		{Name: "SKILL_LARGE_HYBRID_TURRET", Kind: "const", Const: skillLargeHybridTurret},
		{Name: "SKILL_LARGE_PROJECTILE_TURRET", Kind: "const", Const: skillLargeProjectileTurret},
		{Name: "SKILL_LARGE_PRECURSOR_WEAPON", Kind: "const", Const: skillLargePrecursorWeapon},
		{Name: "SKILL_TORPEDOES", Kind: "const", Const: skillTorpedoes},
		{Name: "SKILL_CRUISE_MISSILES", Kind: "const", Const: skillCruiseMissiles},
		{Name: "SKILL_HEAVY_MISSILES", Kind: "const", Const: skillHeavyMissiles},
		{Name: "SKILL_TORPEDO_SPEC", Kind: "const", Const: skillTorpedoSpec},
		{Name: "SKILL_CRUISE_MISSILE_SPEC", Kind: "const", Const: skillCruiseMissileSpec},
		{Name: "SKILL_REPAIR_SYSTEMS", Kind: "const", Const: skillRepairSystems},
		{Name: "SKILL_SHIELD_OPERATION", Kind: "const", Const: skillShieldOperation},
		{Name: "SKILL_MJD_OPERATION", Kind: "const", Const: skillMjdOperation},
		{Name: "ATTR_SHIELD_EM_RES", Kind: "const", Const: attrShieldEmRes},
		{Name: "ATTR_SHIELD_EXP_RES", Kind: "const", Const: attrShieldExpRes},
		{Name: "ATTR_SHIELD_KIN_RES", Kind: "const", Const: attrShieldKinRes},
		{Name: "ATTR_SHIELD_THERM_RES", Kind: "const", Const: attrShieldThermRes},
		{Name: "ATTR_SHIELD_BOOST_MULTIPLIER", Kind: "const", Const: attrShieldBoostMultiplier},
		{Name: "ATTR_SPEED", Kind: "const", Const: attrSpeed},
		{Name: "ATTR_CAPACITOR_NEED", Kind: "const", Const: attrCapacitorNeed},
		{Name: "ATTR_ACTIVATION_BLOCKED", Kind: "const", Const: attrActivationBlocked},
		{Name: "ATTR_ACTIVATION_BLOCKED_STR", Kind: "const", Const: attrActivationBlockedStr},
		{Name: "ATTR_REMOTE_REPAIR_IMPEDANCE_BONUS", Kind: "const", Const: attrRemoteRepairImpedanceBonus},
		{Name: "ATTR_REMOTE_ASSIST_IMPEDANCE_BONUS", Kind: "const", Const: attrRemoteAssistImpedanceBonus},
		{Name: "ATTR_SENSOR_DAMP_RESIST_BONUS", Kind: "const", Const: attrSensorDampResistBonus},
		{Name: "ATTR_WEAPON_DISRUPT_RESIST_BONUS", Kind: "const", Const: attrWeaponDisruptResistBonus},
		{Name: "ATTR_TARGET_PAINT_RESIST_BONUS", Kind: "const", Const: attrTargetPaintResistBonus},
		{Name: "ATTR_REMOTE_REPAIR_IMPEDANCE", Kind: "const", Const: attrRemoteRepairImpedance},
		{Name: "ATTR_REMOTE_ASSIST_IMPEDANCE", Kind: "const", Const: attrRemoteAssistImpedance},
		{Name: "ATTR_SENSOR_DAMP_RESIST", Kind: "const", Const: attrSensorDampResist},
		{Name: "ATTR_WEAPON_DISRUPT_RESIST", Kind: "const", Const: attrWeaponDisruptResist},
		{Name: "ATTR_TARGET_PAINT_RESIST", Kind: "const", Const: attrTargetPaintResist},
		{Name: "ATTR_BASTION_REPAIR_DURATION_BONUS", Kind: "const", Const: attrBastionRepairDurationBonus},
		{Name: "ATTR_ARMOR_DAMAGE_AMOUNT_BONUS", Kind: "const", Const: attrArmorDamageAmountBonus},
		{Name: "MUTADAPTIVE_RAR_EFFECT_ID", Kind: "const", Const: mutadaptiveRarEffectId},
		{Name: "ATTR_MUTADAPTIVE_BONUS_MAX", Kind: "const", Const: attrMutadaptiveBonusMax},
		{Name: "ATTR_MUTADAPTIVE_BONUS_PER_CYCLE", Kind: "const", Const: attrMutadaptiveBonusPerCycle},
		{Name: "LEGACY_DOOMSDAY_EFFECT_IDS", Kind: "set", Ints: legacyDoomsdayEffectIdsEntries},
		{Name: "COMMAND_BURST_EFFECT_IDS", Kind: "set", Ints: commandBurstEffectIdsEntries},
		{Name: "SMART_BOMB_GROUP_ID", Kind: "const", Const: smartBombGroupId},
		{Name: "SMALL_TURRET_SKILL_IDS", Kind: "set", Ints: smallTurretSkillIdsEntries},
		{Name: "SMALL_MISSILE_SKILL_IDS", Kind: "set", Ints: smallMissileSkillIdsEntries},
		{Name: "SHIELD_BOOSTER_EFFECT_IDS", Kind: "set", Ints: shieldBoosterEffectIdsEntries},
		{Name: "ARMOR_REPAIRER_EFFECT_IDS", Kind: "set", Ints: armorRepairerEffectIdsEntries},
		{Name: "REMOTE_SHIELD_XFER_EFFECT_IDS", Kind: "set", Ints: remoteShieldXferEffectIdsEntries},
		{Name: "REMOTE_ARMOR_REP_EFFECT_IDS", Kind: "set", Ints: remoteArmorRepEffectIdsEntries},
		{Name: "ENERGY_NEUTRALIZER_EFFECT_IDS", Kind: "set", Ints: energyNeutralizerEffectIdsEntries},
		{Name: "ENERGY_NOSFERATU_EFFECT_IDS", Kind: "set", Ints: energyNosferatuEffectIdsEntries},
		{Name: "STASIS_WEB_EFFECT_IDS", Kind: "set", Ints: stasisWebEffectIdsEntries},
		{Name: "TARGET_PAINTER_EFFECT_IDS", Kind: "set", Ints: targetPainterEffectIdsEntries},
		{Name: "FIGHTER_AB_EFFECT_ID", Kind: "const", Const: fighterAbEffectId},
		{Name: "FIGHTER_MWD_EFFECT_ID", Kind: "const", Const: fighterMwdEffectId},
		{Name: "FIGHTER_EVASIVE_EFFECT_ID", Kind: "const", Const: fighterEvasiveEffectId},
		{Name: "ATTR_FIGHTER_AB_SPEED_BONUS", Kind: "const", Const: attrFighterAbSpeedBonus},
		{Name: "ATTR_FIGHTER_MWD_SPEED_BONUS", Kind: "const", Const: attrFighterMwdSpeedBonus},
		{Name: "ATTR_FIGHTER_MWD_SIG_BONUS", Kind: "const", Const: attrFighterMwdSigBonus},
		{Name: "ATTR_FIGHTER_EVASIVE_SPEED_BONUS", Kind: "const", Const: attrFighterEvasiveSpeedBonus},
		{Name: "ATTR_FIGHTER_EVASIVE_SIG_BONUS", Kind: "const", Const: attrFighterEvasiveSigBonus},
		{Name: "ATTR_FIGHTER_EVASIVE_EM_RES", Kind: "const", Const: attrFighterEvasiveEmRes},
		{Name: "ATTR_FIGHTER_EVASIVE_THM_RES", Kind: "const", Const: attrFighterEvasiveThmRes},
		{Name: "ATTR_FIGHTER_EVASIVE_KIN_RES", Kind: "const", Const: attrFighterEvasiveKinRes},
		{Name: "ATTR_FIGHTER_EVASIVE_EXP_RES", Kind: "const", Const: attrFighterEvasiveExpRes},
		{Name: "ATTR_SHIELD_EM_RESO", Kind: "const", Const: attrShieldEmReso},
		{Name: "ATTR_SHIELD_EXP_RESO", Kind: "const", Const: attrShieldExpReso},
		{Name: "ATTR_SHIELD_KIN_RESO", Kind: "const", Const: attrShieldKinReso},
		{Name: "ATTR_SHIELD_THM_RESO", Kind: "const", Const: attrShieldThmReso},
		{Name: "DISINTEGRATOR_EFFECT_ID", Kind: "const", Const: disintegratorEffectId},
		{Name: "ATTR_DMG_MULT_BONUS_PER_CYCLE", Kind: "const", Const: attrDmgMultBonusPerCycle},
		{Name: "ATTR_DMG_MULT_BONUS_MAX", Kind: "const", Const: attrDmgMultBonusMax},
		{Name: "RAH_EFFECT_ID", Kind: "const", Const: rahEffectId},
		{Name: "SUBSYSTEM_SLOT_BONUS_MAP", Kind: "pairs", Pairs: subsystemSlotBonusMap},
		{Name: "LEGACY_HANDLED_HARDCODED_EFFECTS", Kind: "set", Ints: legacyHandledHardcodedEffectsEntries},
		{Name: "SEC_STATUS_SCALED_EFFECT_IDS", Kind: "set", Ints: SecStatusScaledEffectIdsEntries},
	}
}
