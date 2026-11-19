// Tipi condivisi (dataset SDE, modello del fit). Definiti qui, in un package foglia,
// perche' internal/engine e fitengine (che espone ComputeFit) non formino un ciclo; fitengine li
// riesporta con alias. Equivalente di src/types.ts: stessi nomi di campo (in PascalCase), campi
// opzionali con puntatori.
package types

import "github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"

// Table e' una tabella del bundle indicizzata per id numerico. In JS gli
// oggetti a chiavi numeriche si iterano per intero ascendente: la slice interna
// e' gia' in quell'ordine ed e' l'unica vista da usare per iterare (mai range
// su map nei package dell'engine).
type Table[V any] struct {
	keys   []int
	values []V
	byKey  map[int]int
}

// NewTable costruisce la tabella da chiavi e valori GIA' ordinati per chiave
// ascendente (lo garantisce il loader del bundle).
func NewTable[V any](keys []int, values []V) Table[V] {
	idx := make(map[int]int, len(keys))
	for i, k := range keys {
		idx[k] = i
	}
	return Table[V]{keys: keys, values: values, byKey: idx}
}

// Len e' il numero di voci.
func (t Table[V]) Len() int { return len(t.keys) }

// At restituisce la i-esima voce nell'ordine ascendente per chiave.
func (t Table[V]) At(i int) (key int, v *V) { return t.keys[i], &t.values[i] }

// Get cerca per chiave.
func (t Table[V]) Get(key int) (*V, bool) {
	i, ok := t.byKey[key]
	if !ok {
		return nil, false
	}
	return &t.values[i], true
}

// Has dice se la chiave e' presente.
func (t Table[V]) Has(key int) bool { _, ok := t.byKey[key]; return ok }

// Keys restituisce le chiavi in ordine ascendente (la slice e' condivisa: non modificarla).
func (t Table[V]) Keys() []int { return t.keys }

type SdeAttribute struct {
	ID                  int     `json:"id"`
	Name                string  `json:"name"`
	DisplayName         *string `json:"displayName,omitempty"`
	UnitID              *int    `json:"unitID,omitempty"`
	IconID              *int    `json:"iconID,omitempty"`
	DefaultValue        float64 `json:"defaultValue"`
	HighIsGood          bool    `json:"highIsGood"`
	Stackable           bool    `json:"stackable"`
	AttributeCategoryID *int    `json:"attributeCategoryID,omitempty"`
	DataType            *int    `json:"dataType,omitempty"`
	// MaxAttributeID: attributo il cui valore pone un tetto a questo (letto sullo stesso item).
	MaxAttributeID *int `json:"maxAttributeID,omitempty"`
	// MinAttributeID: attributo il cui valore pone un minimo a questo.
	MinAttributeID *int `json:"minAttributeID,omitempty"`
}

type SdeUnit struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	DisplayName *string `json:"displayName,omitempty"`
}

// SdeModifierInfo: Domain e Func sono le stringhe dell'SDE (vedi types.ts).
type SdeModifierInfo struct {
	Domain               string `json:"domain"`
	Func                 string `json:"func"`
	ModifiedAttributeID  *int   `json:"modifiedAttributeID,omitempty"`
	ModifyingAttributeID *int   `json:"modifyingAttributeID,omitempty"`
	Operation            *int   `json:"operation,omitempty"`
	GroupID              *int   `json:"groupID,omitempty"`
	SkillTypeID          *int   `json:"skillTypeID,omitempty"`
	// EffectID: solo per func EffectStopper.
	EffectID *int `json:"effectID,omitempty"`
}

type SdeEffect struct {
	ID                            int               `json:"id"`
	Name                          string            `json:"name"`
	DisplayName                   *string           `json:"displayName,omitempty"`
	EffectCategoryID              int               `json:"effectCategoryID"`
	IsOffensive                   bool              `json:"isOffensive"`
	IsAssistance                  bool              `json:"isAssistance"`
	IsWarpSafe                    bool              `json:"isWarpSafe"`
	DurationAttributeID           *int              `json:"durationAttributeID,omitempty"`
	DischargeAttributeID          *int              `json:"dischargeAttributeID,omitempty"`
	RangeAttributeID              *int              `json:"rangeAttributeID,omitempty"`
	FalloffAttributeID            *int              `json:"falloffAttributeID,omitempty"`
	TrackingSpeedAttributeID      *int              `json:"trackingSpeedAttributeID,omitempty"`
	FittingUsageChanceAttributeID *int              `json:"fittingUsageChanceAttributeID,omitempty"`
	ResistanceAttributeID         *int              `json:"resistanceAttributeID,omitempty"`
	Distribution                  *int              `json:"distribution,omitempty"`
	PropulsionChance              bool              `json:"propulsionChance"`
	ElectronicChance              bool              `json:"electronicChance"`
	RangeChance                   bool              `json:"rangeChance"`
	DisallowAutoRepeat            bool              `json:"disallowAutoRepeat"`
	GUID                          *string           `json:"guid,omitempty"`
	ModifierInfo                  []SdeModifierInfo `json:"modifierInfo"`
}

type SdeColor struct {
	R float64 `json:"r"`
	G float64 `json:"g"`
	B float64 `json:"b"`
}

type SdeMetaGroup struct {
	ID     int       `json:"id"`
	Name   *string   `json:"name,omitempty"`
	Color  *SdeColor `json:"color,omitempty"`
	IconID *int      `json:"iconID,omitempty"`
}

type SdeCategory struct {
	ID   int     `json:"id"`
	Name *string `json:"name,omitempty"`
}

type SdeGroup struct {
	ID         int     `json:"id"`
	CategoryID int     `json:"categoryID"`
	Name       *string `json:"name,omitempty"`
}

// SdeMarketGroup: nodo della tassonomia Market (albero via ParentGroupID).
type SdeMarketGroup struct {
	ID            int     `json:"id"`
	Name          *string `json:"name,omitempty"`
	ParentGroupID *int    `json:"parentGroupID,omitempty"`
	IconID        *int    `json:"iconID,omitempty"`
	HasTypes      *bool   `json:"hasTypes,omitempty"`
}

type SdeCloneGradeSkill struct {
	TypeID int `json:"typeID"`
	Level  int `json:"level"`
}

type SdeCloneGrade struct {
	ID     int                  `json:"id"`
	Name   string               `json:"name"`
	Skills []SdeCloneGradeSkill `json:"skills"`
}

type SdeDbuffAttr struct {
	DogmaAttributeID int `json:"dogmaAttributeID"`
}

type SdeDbuffGroupAttr struct {
	DogmaAttributeID int `json:"dogmaAttributeID"`
	GroupID          int `json:"groupID"`
}

type SdeDbuffSkillAttr struct {
	DogmaAttributeID int `json:"dogmaAttributeID"`
	SkillID          int `json:"skillID"`
}

type SdeDbuffCollection struct {
	ID                             int                 `json:"id"`
	AggregateMode                  string              `json:"aggregateMode"`
	OperationName                  string              `json:"operationName"`
	DisplayName                    *string             `json:"displayName,omitempty"`
	ShowOutputValueInUI            *string             `json:"showOutputValueInUI,omitempty"`
	ItemModifiers                  []SdeDbuffAttr      `json:"itemModifiers"`
	LocationModifiers              []SdeDbuffAttr      `json:"locationModifiers"`
	LocationGroupModifiers         []SdeDbuffGroupAttr `json:"locationGroupModifiers"`
	LocationRequiredSkillModifiers []SdeDbuffSkillAttr `json:"locationRequiredSkillModifiers"`
}

type SdeDynamicAttrRange struct {
	ID  int     `json:"id"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type SdeDynamicMapping struct {
	ApplicableTypes []int `json:"applicableTypes"`
	ResultingType   int   `json:"resultingType"`
}

type SdeDynamicAttribute struct {
	ID                 int                   `json:"id"`
	AttributeIDs       []SdeDynamicAttrRange `json:"attributeIDs"`
	InputOutputMapping []SdeDynamicMapping   `json:"inputOutputMapping"`
}

type SdeTypeAttr struct {
	ID int     `json:"id"`
	V  float64 `json:"v"`
}

type SdeTypeEffect struct {
	ID  int `json:"id"`
	Def int `json:"def"` // 0 | 1
}

type SdeType struct {
	ID                    int             `json:"id"`
	Name                  string          `json:"name"`
	GroupID               int             `json:"groupID"`
	CategoryID            int             `json:"categoryID"`
	MarketGroupID         *int            `json:"marketGroupID,omitempty"`
	IconID                *int            `json:"iconID,omitempty"`
	MetaGroupID           *int            `json:"metaGroupID,omitempty"`
	MetaLevel             *int            `json:"metaLevel,omitempty"`
	VariationParentTypeID *int            `json:"variationParentTypeID,omitempty"`
	Mass                  *float64        `json:"mass,omitempty"`
	Volume                *float64        `json:"volume,omitempty"`
	Capacity              *float64        `json:"capacity,omitempty"`
	PortionSize           *float64        `json:"portionSize,omitempty"`
	BasePrice             *float64        `json:"basePrice,omitempty"`
	Attributes            []SdeTypeAttr   `json:"attributes"`
	Effects               []SdeTypeEffect `json:"effects"`
}

// TypeBucket e' un file di types/. L'ordine delle costanti e' l'ordine di
// BUCKET_FILES in node.ts, usato da GetType.
type TypeBucket int

const (
	BucketShips TypeBucket = iota
	BucketModules
	BucketCharges
	BucketDrones
	BucketFighters
	BucketImplants
	BucketSubsystems
	BucketSkills
	BucketSystemEffects
	BucketStructures
	BucketStructureModules
	BucketMutaplasmids
	NumBuckets
)

// BucketNames: nome TS (chiave di typesByBucket) per bucket.
var BucketNames = [NumBuckets]string{
	"ships", "modules", "charges", "drones", "fighters", "implants", "subsystems",
	"skills", "systemEffects", "structures", "structureModules", "mutaplasmids",
}

// BucketFiles: file (relativo a v<hash>/) per bucket.
var BucketFiles = [NumBuckets]string{
	"types/ships.json", "types/modules.json", "types/charges.json", "types/drones.json",
	"types/fighters.json", "types/implants.json", "types/subsystems.json", "types/skills.json",
	"types/system-effects.json", "types/structures.json", "types/structure-modules.json",
	"types/mutaplasmids.json",
}

// BucketOptional: bucket il cui file puo' mancare (diventa vuoto), come OPTIONAL in node.ts.
var BucketOptional = [NumBuckets]bool{
	BucketSystemEffects: true, BucketStructures: true, BucketStructureModules: true, BucketMutaplasmids: true,
}

// FittingDataset e' l'equivalente di FittingDataset in types.ts.
type FittingDataset struct {
	Version           string
	Attributes        Table[SdeAttribute]
	Units             Table[SdeUnit]
	Effects           Table[SdeEffect]
	MetaGroups        Table[SdeMetaGroup]
	Categories        Table[SdeCategory]
	Groups            Table[SdeGroup]
	MarketGroups      Table[SdeMarketGroup]
	CloneGrades       Table[SdeCloneGrade]
	DbuffCollections  Table[SdeDbuffCollection]
	DynamicAttributes Table[SdeDynamicAttribute]
	TypesByBucket     [NumBuckets]Table[SdeType]
}

// GetType risolve un type attraverso i bucket nell'ordine di TypeBucket.
func (d *FittingDataset) GetType(id int) (*SdeType, bool) {
	for b := TypeBucket(0); b < NumBuckets; b++ {
		if t, ok := d.TypesByBucket[b].Get(id); ok {
			return t, true
		}
	}
	return nil, false
}

// ---------------------------------------------------------------------------
// Modello del fit (sezione "Fit model" di types.ts) — solo quanto serve a ItemState.
// ---------------------------------------------------------------------------

type SlotType string

const (
	SlotHI        SlotType = "HI"
	SlotMED       SlotType = "MED"
	SlotLO        SlotType = "LO"
	SlotRIG       SlotType = "RIG"
	SlotSUBSYSTEM SlotType = "SUBSYSTEM"
	SlotSERVICE   SlotType = "SERVICE"
)

type ModuleState string

const (
	StateOffline  ModuleState = "OFFLINE"
	StateOnline   ModuleState = "ONLINE"
	StateActive   ModuleState = "ACTIVE"
	StateOverload ModuleState = "OVERLOAD"
)

// MutatorData: Attributes e' una mappa attributeID -> valore (solo letture per chiave).
type MutatorData struct {
	DynamicTypeID int
	Attributes    map[int]float64
	SourceTypeID  *int
}

type FitModule struct {
	ID           string
	SlotType     SlotType
	Position     float64
	TypeID       int
	State        ModuleState
	ChargeTypeID *int
	Mutator      *MutatorData
}

type FitDrone struct {
	ID          string
	TypeID      int
	CountTotal  float64
	CountActive float64
}

type FitFighter struct {
	ID           string
	TypeID       int
	Count        float64
	AbilityState map[int]bool
}

type FitImplant struct {
	ID     string
	TypeID int
	Slot   float64
}

type FitBooster struct {
	ID                string
	TypeID            int
	Slot              float64
	ActiveSideEffects []int
}

type FitSubsystem struct {
	ID     string
	Slot   int
	TypeID int
}

// ModifierOperation: ordine significativo (la pipeline le applica in sequenza).
type ModifierOperation string

const (
	OpPreAssign   ModifierOperation = "PreAssign"
	OpPreMul      ModifierOperation = "PreMul"
	OpPreDiv      ModifierOperation = "PreDiv"
	OpModAdd      ModifierOperation = "ModAdd"
	OpModSub      ModifierOperation = "ModSub"
	OpPostMul     ModifierOperation = "PostMul"
	OpPostDiv     ModifierOperation = "PostDiv"
	OpPostPercent ModifierOperation = "PostPercent"
	OpPostAssign  ModifierOperation = "PostAssign"
)

// ModifierAffliction: StackingGroup nil = senza penalita'. ResistanceAttributeID opzionale.
type ModifierAffliction struct {
	SourceKind            string
	SourceID              string
	Operation             ModifierOperation
	Value                 float64
	StackingGroup         *string `encode:"null"`
	ResistanceAttributeID *int
}

// SkillProfile (types.ts): Skills e' typeID -> 0..5, in ordine di id ascendente (come l'iterazione di un
// oggetto JS con chiavi intere); chi lo costruisce lo riempie in quell'ordine.
type SkillProfile struct {
	ID                *string
	Name              string
	IsDefault         bool
	Source            string
	SourceCharacterID *string
	Skills            *omap.Map[int, int]
	SyncedAt          *string
}

// DamageProfile (types.ts).
type DamageProfile struct {
	ID        *string
	Name      string
	Em        float64
	Thermal   float64
	Kinetic   float64
	Explosive float64
	IsPreset  *bool
}

// FitCargo (types.ts).
type FitCargo struct {
	ID     string
	TypeID int
	Count  float64
}

// Fit e' il modello persistente di un fit (types.ts). I campi puramente anagrafici sono omessi.
type Fit struct {
	ShipTypeID int
	Name       string
	Visibility string
	Tags       []string
	Modules    []FitModule
	Drones     []FitDrone
	Fighters   []FitFighter
	Cargo      []FitCargo
	Implants   []FitImplant
	Boosters   []FitBooster
	Subsystems []FitSubsystem
	// ModeTypeID: modo T3D/T3C esclusivo (nil = non specificato).
	ModeTypeID *int
}

// ProjectedSource: sorgente ostile proiettata sul fit (types.ts). ProjectionRange nil = non impostato.
type ProjectedSource struct {
	ID              string
	TypeID          int
	State           ModuleState
	ChargeTypeID    *int
	Mutator         *MutatorData
	ProjectionRange *float64
}

// TargetProfile (types.ts).
type TargetProfile struct {
	ID              *string
	Name            string
	SignatureRadius float64
	MaxVelocity     float64
	EmResist        float64
	ThermalResist   float64
	KineticResist   float64
	ExplosiveResist float64
	IsPreset        *bool
}
