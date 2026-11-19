// Porting letterale di src/itemState.ts: stato runtime di un item adattabile
// (nave, modulo, drone, fighter, impianto, booster, carica, subsystem, mode, personaggio).
package engine

import (
	"fmt"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

type ItemKind string

const (
	KindShip      ItemKind = "ship"
	KindModule    ItemKind = "module"
	KindDrone     ItemKind = "drone"
	KindFighter   ItemKind = "fighter"
	KindImplant   ItemKind = "implant"
	KindBooster   ItemKind = "booster"
	KindCharge    ItemKind = "charge"
	KindSubsystem ItemKind = "subsystem"
	KindMode      ItemKind = "mode"
	KindCharacter ItemKind = "character"
)

// ItemStateInit e' ItemStateInit. State "" = default ONLINE.
type ItemStateInit struct {
	Kind ItemKind
	ID   string
	Type *types.SdeType
	// State: stato runtime; "" vale ONLINE.
	State types.ModuleState
	// Charge: carica caricata (solo per i moduli).
	Charge *ItemState
	// AttributeOverrides: override per istanza (mutaplasmide), solo letture per chiave.
	AttributeOverrides map[int]float64
	// ActiveSideEffects: effetti collaterali del booster attivati dal pilota.
	ActiveSideEffects []int
}

type ItemState struct {
	Kind       ItemKind
	ID         string
	Type       *types.SdeType
	TypeID     int
	GroupID    int
	CategoryID int
	State      types.ModuleState
	Charge     *ItemState
	// EffectIDs: gli effetti dell'item, in ordine di apparizione nel type.
	EffectIDs *omap.Set[int]
	// ActiveSideEffects: effetti collaterali del booster accesi per questa istanza.
	ActiveSideEffects *omap.Set[int]
	// AttrSchema serve a rispettare i tetti dell'SDE (min/maxAttributeID); nil = nessun clamp.
	AttrSchema *types.Table[types.SdeAttribute]

	Attrs *omap.Map[int, *ModifiedAttribute]
}

// NewItemState e' il costruttore di ItemState.
func NewItemState(init ItemStateInit) *ItemState {
	s := &ItemState{
		Kind:       init.Kind,
		ID:         init.ID,
		Type:       init.Type,
		TypeID:     init.Type.ID,
		GroupID:    init.Type.GroupID,
		CategoryID: init.Type.CategoryID,
		State:      init.State,
		Charge:     init.Charge,
		Attrs:      omap.New[int, *ModifiedAttribute](),
	}
	if s.State == "" {
		s.State = types.StateOnline
	}
	s.ActiveSideEffects = omap.NewSet(init.ActiveSideEffects...)

	// Base degli attributi dal typeDogma; gli override (mutaplasmide) prevalgono.
	for _, a := range init.Type.Attributes {
		base := a.V
		if ov, ok := init.AttributeOverrides[a.ID]; ok {
			base = ov
		}
		s.Attrs.Set(a.ID, NewModifiedAttribute(a.ID, base))
	}

	// Massa / volume / capacita' stanno sul type, non nel typeDogma: back-fill.
	if !s.Attrs.Has(AttrMass) && init.Type.Mass != nil {
		s.Attrs.Set(AttrMass, NewModifiedAttribute(AttrMass, *init.Type.Mass))
	}
	if !s.Attrs.Has(AttrVolume) && init.Type.Volume != nil {
		s.Attrs.Set(AttrVolume, NewModifiedAttribute(AttrVolume, *init.Type.Volume))
	}
	if !s.Attrs.Has(AttrCapacity) && init.Type.Capacity != nil {
		s.Attrs.Set(AttrCapacity, NewModifiedAttribute(AttrCapacity, *init.Type.Capacity))
	}

	s.EffectIDs = omap.NewSet[int]()
	for _, e := range init.Type.Effects {
		s.EffectIDs.Add(e.ID)
	}
	return s
}

// Attr restituisce (creandolo) il ModifiedAttribute: permette ai modificatori di
// puntare ad attributi assenti dal typeDogma.
func (s *ItemState) Attr(id int, defaultBase float64) *ModifiedAttribute {
	ma, ok := s.Attrs.Get(id)
	if !ok {
		ma = NewModifiedAttribute(id, defaultBase)
		s.Attrs.Set(id, ma)
	}
	return ma
}

// GetBase legge il valore base; false se il type non porta l'attributo.
func (s *ItemState) GetBase(id int) (float64, bool) {
	if ma, ok := s.Attrs.Get(id); ok {
		return ma.Base, true
	}
	return 0, false
}

// HasAttr dice se l'attributo e' presente (distingue "assente" da "finale 0").
func (s *ItemState) HasAttr(id int) bool { return s.Attrs.Has(id) }

// GetFinal calcola il valore corrente (con i tetti dell'SDE); fallback se assente.
func (s *ItemState) GetFinal(id int, fallback float64) float64 {
	ma, ok := s.Attrs.Get(id)
	if !ok {
		return fallback
	}
	return s.clampToCaps(id, ma.Compute(nil))
}

// clampToCaps replica `__calculateValue` di pyfa: il tetto/minimo e' un altro
// attributo letto su QUESTO item.
func (s *ItemState) clampToCaps(id int, value float64) float64 {
	if s.AttrSchema == nil {
		return value
	}
	schema, ok := s.AttrSchema.Get(id)
	if !ok {
		return value
	}
	out := value
	if schema.MinAttributeID != nil {
		if floor, ok := s.capValue(*schema.MinAttributeID); ok {
			out = jsmath.Max(out, floor)
		}
	}
	if schema.MaxAttributeID != nil {
		if ceil, ok := s.capValue(*schema.MaxAttributeID); ok {
			out = jsmath.Min(out, ceil)
		}
	}
	return out
}

// capValue: valore dell'attributo che fa da limite, con ripiego sul default SDE.
func (s *ItemState) capValue(capAttrID int) (float64, bool) {
	if ma, ok := s.Attrs.Get(capAttrID); ok {
		return ma.Compute(nil), true
	}
	if s.AttrSchema != nil {
		if a, ok := s.AttrSchema.Get(capAttrID); ok {
			return a.DefaultValue, true
		}
	}
	return 0, false
}

// AddAffliction aggiunge una modifica; defaultBase e' la base se l'attributo non e' nel typeDogma.
func (s *ItemState) AddAffliction(attributeID int, a types.ModifierAffliction, defaultBase float64) {
	s.Attr(attributeID, defaultBase).AddAffliction(a)
}

// ResetAttributes azzera tutti gli attributi (ricalcolo incrementale).
func (s *ItemState) ResetAttributes() {
	for i := 0; i < s.Attrs.Len(); i++ {
		_, ma := s.Attrs.At(i)
		ma.Reset()
	}
}

// SnapshotAll: snapshot di tutti gli attributi in ordine di inserimento.
func (s *ItemState) SnapshotAll() *omap.Map[int, ComputedAttribute] {
	out := omap.New[int, ComputedAttribute]()
	for i := 0; i < s.Attrs.Len(); i++ {
		id, ma := s.Attrs.At(i)
		out.Set(id, ma.Snapshot(nil))
	}
	return out
}

// AppliesAtState: gli effetti dell'item sono attivi allo stato corrente?
func (s *ItemState) AppliesAtState(effect *types.SdeEffect) bool {
	switch effect.EffectCategoryID {
	case 0, 6:
		return true // sempre attivo
	case 4:
		return s.State != types.StateOffline // richiede alimentazione
	case 1, 2, 3:
		return s.State == types.StateActive || s.State == types.StateOverload
	case 5:
		return s.State == types.StateOverload
	default:
		// Categoria sconosciuta: prudenza, come ONLINE.
		return s.State != types.StateOffline
	}
}

// SlotType: slot di un modulo dedotto dagli effetti; false per gli item non-modulo.
func (s *ItemState) SlotType() (types.SlotType, bool) {
	for i := 0; i < s.EffectIDs.Len(); i++ {
		if slot, ok := SlotEffectToSlotType(s.EffectIDs.At(i)); ok {
			return slot, true
		}
	}
	return "", false
}

// RequiredSkill e' un elemento di RequiredSkills.
type RequiredSkill struct {
	SkillID float64
	Level   float64
}

// RequiredSkills: skill richieste, dagli attributi riservati REQUIRED_SKILL_N / _LEVEL.
func (s *ItemState) RequiredSkills() []RequiredSkill {
	var out []RequiredSkill
	for _, pair := range RequiredSkillPairs {
		skillID, ok := s.GetBase(pair[0])
		if ok && skillID > 0 {
			level, ok := s.GetBase(pair[1])
			if !ok {
				level = 1
			}
			out = append(out, RequiredSkill{SkillID: skillID, Level: level})
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// Factory: dai modelli Fit* persistenti agli ItemState runtime.
// ---------------------------------------------------------------------------

func MakeShipState(t *types.SdeType) *ItemState {
	return NewItemState(ItemStateInit{Kind: KindShip, ID: "ship", Type: t})
}

func MakeCharacterState(charType *types.SdeType) *ItemState {
	return NewItemState(ItemStateInit{Kind: KindCharacter, ID: "char", Type: charType})
}

func MakeModuleState(fm *types.FitModule, t *types.SdeType, chargeType *types.SdeType) *ItemState {
	var charge *ItemState
	if chargeType != nil {
		charge = NewItemState(ItemStateInit{Kind: KindCharge, ID: fm.ID + ":charge", Type: chargeType})
	}
	var overrides map[int]float64
	if fm.Mutator != nil {
		overrides = fm.Mutator.Attributes
	}
	return NewItemState(ItemStateInit{
		Kind: KindModule, ID: fm.ID, Type: t, State: fm.State, Charge: charge, AttributeOverrides: overrides,
	})
}

func MakeDroneState(fd *types.FitDrone, t *types.SdeType) *ItemState {
	// Tutti i droni restano nell'engine; il filtro avviene in fase di apply con countActive.
	state := types.StateOffline
	if fd.CountActive > 0 {
		state = types.StateActive
	}
	return NewItemState(ItemStateInit{Kind: KindDrone, ID: fd.ID, Type: t, State: state})
}

func MakeFighterState(ff *types.FitFighter, t *types.SdeType) *ItemState {
	state := types.StateOffline
	if ff.Count > 0 {
		state = types.StateActive
	}
	return NewItemState(ItemStateInit{Kind: KindFighter, ID: ff.ID, Type: t, State: state})
}

func MakeImplantState(fi *types.FitImplant, t *types.SdeType) *ItemState {
	return NewItemState(ItemStateInit{Kind: KindImplant, ID: fi.ID, Type: t, State: types.StateOnline})
}

func MakeBoosterState(fb *types.FitBooster, t *types.SdeType) *ItemState {
	// activeSideEffects e' l'elenco opt-in del pilota: i drawback sono OFF finche' non li accende.
	return NewItemState(ItemStateInit{
		Kind: KindBooster, ID: fb.ID, Type: t, State: types.StateOnline, ActiveSideEffects: fb.ActiveSideEffects,
	})
}

func MakeSubsystemState(fs *types.FitSubsystem, t *types.SdeType) *ItemState {
	return NewItemState(ItemStateInit{Kind: KindSubsystem, ID: fs.ID, Type: t, State: types.StateOnline})
}

func MakeModeState(modeTypeID int, t *types.SdeType) *ItemState {
	return NewItemState(ItemStateInit{
		Kind: KindMode, ID: fmt.Sprintf("mode:%d", modeTypeID), Type: t, State: types.StateActive,
	})
}

// ShipHardpoints: torrette e lanciatori della nave.
func ShipHardpoints(ship *ItemState) (turret, launcher float64) {
	return ship.GetFinal(AttrTurretHardpoints, 0), ship.GetFinal(AttrLauncherHardpoints, 0)
}
