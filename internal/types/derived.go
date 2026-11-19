package types

// Tipi del risultato di computeFit (sezione "Calc engine types" di types.ts). I campi opzionali sono
// puntatori; le collezioni a chiave (Map in TS) restano in internal/engine come omap.

type Resistances struct{ Em, Thermal, Kinetic, Explosive float64 }

// LayerEhp (derived/ehp.ts).
type LayerEhp struct {
	Hp                float64
	EhpUniform        float64
	EhpAgainstProfile float64
	Resistances       Resistances
}

type Defense struct {
	Shield, Armor, Hull    LayerEhp
	EhpTotalAgainstProfile float64
}

type SlotUse struct{ Used, Max float64 }

type Hardpoints struct{ Turret, Launcher SlotUse }

type Fitting struct {
	CpuUsed, CpuMax                       float64
	PowerUsed, PowerMax                   float64
	CalibrationUsed, CalibrationMax       float64
	DroneBandwidthUsed, DroneBandwidthMax float64
	DroneBayUsed, DroneBayMax             float64
	Slots                                 SlotTable
	Hardpoints                            Hardpoints
}

// SlotTable e' Record<SlotType, {used,max}> nell'ordine HI, MED, LO, RIG, SUBSYSTEM, SERVICE.
type SlotTable struct{ HI, MED, LO, RIG, SUBSYSTEM, SERVICE SlotUse }

type WeaponKind string

type DamageSplit struct{ Em, Thermal, Kinetic, Explosive, Total float64 }

type WeaponRange struct {
	Optimal, Falloff, Tracking, BurstRange               float64
	ExplosionRadius, ExplosionVelocity, Drf, FlightRange float64
}

type DisintegratorInfo struct{ MaxBonus, BonusPerCycle, BaseDps float64 }

// WeaponContribution (types.ts).
type WeaponContribution struct {
	SourceID         string
	TypeID           int
	Name             *string
	Kind             WeaponKind
	Alpha            float64
	Dps              float64
	SustainedDps     *float64
	ReloadSeconds    *float64
	ChargesPerLoad   *float64
	ChainDpsMax      *float64
	ChainTargetCount *float64
	CycleSeconds     float64
	Damages          DamageSplit
	Range            WeaponRange
	ChargeTypeID     *int
	Count            float64
	Disintegrator    *DisintegratorInfo
}

type Offense struct {
	WeaponDps, WeaponSustainedDps, DroneDps, FighterDps              float64
	TotalDps, TotalSustainedDps, AlphaStrike                         float64
	WeaponOptimal, WeaponFalloff                                     float64
	WeaponTracking, ExplosionVelocity, ExplosionRadius, MissileRange *float64
	Breakdown                                                        []WeaponContribution
}

type CapacitorStats struct {
	Capacity, RechargeMs, PeakRechargeRate, UsagePerSecond float64
	Stable                                                 bool
	StablePercent                                          float64
	SecondsToEmpty                                         *float64
}

type TankRates struct {
	ShieldRepairAmount, ShieldRepairDuration, ShieldRepairPerSecond, ShieldRepairPerSecondSustained float64
	ArmorRepairAmount, ArmorRepairDuration, ArmorRepairPerSecond, ArmorRepairPerSecondSustained     float64
	HullRepairAmount, HullRepairDuration, HullRepairPerSecond, HullRepairPerSecondSustained         float64
	PassiveShieldRegenPeak                                                                          float64
}

type Navigation struct{ MaxVelocity, Mass, Agility, AlignTimeSeconds, WarpSpeed float64 }

type Targeting struct {
	MaxTargetingRange, MaxLockedTargets, SignatureRadius, ScanResolution, SensorStrength float64
	SensorType                                                                           string
}

type DroneStats struct {
	BayUsed, BayMax, BandwidthUsed, BandwidthMax, Active, ControlRange float64
}

type StructureServiceModule struct {
	ModuleIndex       int
	TypeID            int
	Name              string
	State             ModuleState
	FuelBlocksPerHour float64
}

type StructureMeta struct {
	ServiceSlotsMax, ServiceSlotsUsed, FuelBlocksPerHour float64
	Services                                             []StructureServiceModule
}

// ProjectedEffectReport (types.ts).
type ProjectedEffectReport struct {
	TypeID    int
	Kind      string
	JamChance *float64
	PerSecond *float64
	Summary   string
}
