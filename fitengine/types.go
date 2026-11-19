// Package fitengine e' l'API pubblica (equivalente di src/index.ts). I tipi sono
// definiti in internal/types (package foglia, per evitare cicli con internal/engine)
// e riesposti qui con alias.
package fitengine

import "github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"

type (
	SdeAttribute        = types.SdeAttribute
	SdeUnit             = types.SdeUnit
	SdeModifierInfo     = types.SdeModifierInfo
	SdeEffect           = types.SdeEffect
	SdeColor            = types.SdeColor
	SdeMetaGroup        = types.SdeMetaGroup
	SdeCategory         = types.SdeCategory
	SdeGroup            = types.SdeGroup
	SdeMarketGroup      = types.SdeMarketGroup
	SdeCloneGradeSkill  = types.SdeCloneGradeSkill
	SdeCloneGrade       = types.SdeCloneGrade
	SdeDbuffAttr        = types.SdeDbuffAttr
	SdeDbuffGroupAttr   = types.SdeDbuffGroupAttr
	SdeDbuffSkillAttr   = types.SdeDbuffSkillAttr
	SdeDbuffCollection  = types.SdeDbuffCollection
	SdeDynamicAttrRange = types.SdeDynamicAttrRange
	SdeDynamicMapping   = types.SdeDynamicMapping
	SdeDynamicAttribute = types.SdeDynamicAttribute
	SdeTypeAttr         = types.SdeTypeAttr
	SdeTypeEffect       = types.SdeTypeEffect
	SdeType             = types.SdeType
	TypeBucket          = types.TypeBucket
	FittingDataset      = types.FittingDataset
	Table[V any]        = types.Table[V]

	SlotType           = types.SlotType
	ModuleState        = types.ModuleState
	MutatorData        = types.MutatorData
	FitModule          = types.FitModule
	FitDrone           = types.FitDrone
	FitFighter         = types.FitFighter
	FitImplant         = types.FitImplant
	FitBooster         = types.FitBooster
	FitSubsystem       = types.FitSubsystem
	FitCargo           = types.FitCargo
	ModifierOperation  = types.ModifierOperation
	ModifierAffliction = types.ModifierAffliction
)

const (
	BucketShips            = types.BucketShips
	BucketModules          = types.BucketModules
	BucketCharges          = types.BucketCharges
	BucketDrones           = types.BucketDrones
	BucketFighters         = types.BucketFighters
	BucketImplants         = types.BucketImplants
	BucketSubsystems       = types.BucketSubsystems
	BucketSkills           = types.BucketSkills
	BucketSystemEffects    = types.BucketSystemEffects
	BucketStructures       = types.BucketStructures
	BucketStructureModules = types.BucketStructureModules
	BucketMutaplasmids     = types.BucketMutaplasmids
	NumBuckets             = types.NumBuckets
)

var (
	BucketNames    = types.BucketNames
	BucketFiles    = types.BucketFiles
	BucketOptional = types.BucketOptional
)

// NewTable costruisce una Table da chiavi gia' ordinate (vedi types.NewTable).
func NewTable[V any](keys []int, values []V) Table[V] { return types.NewTable(keys, values) }
