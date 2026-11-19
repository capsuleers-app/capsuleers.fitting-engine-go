// Porting letterale di src/modifiedAttribute.ts: pipeline di modifica per
// attributo (PreAssign, PreMul/PreDiv, ModAdd/ModSub, PostMul/PostDiv/PostPercent,
// PostAssign, tetto opzionale).
package engine

import (
	"math"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

// ModifiedAttribute e' una coppia (item x attributeID): valore base e lista piatta di afflictions.
type ModifiedAttribute struct {
	AttributeID int
	Base        float64
	Afflictions []types.ModifierAffliction

	cache    float64
	hasCache bool
}

// NewModifiedAttribute crea l'attributo con il valore base.
func NewModifiedAttribute(attributeID int, base float64) *ModifiedAttribute {
	return &ModifiedAttribute{AttributeID: attributeID, Base: base}
}

// AddAffliction aggiunge una modifica e invalida la cache.
func (m *ModifiedAttribute) AddAffliction(a types.ModifierAffliction) {
	m.Afflictions = append(m.Afflictions, a)
	m.hasCache = false
}

// Reset azzera le modifiche mantenendo la base.
func (m *ModifiedAttribute) Reset() {
	m.Afflictions = m.Afflictions[:0]
	m.hasCache = false
}

// Compute calcola il valore finale (in cache fino al prossimo AddAffliction/Reset).
// maxAttribute (nil = assente) e' un tetto: min(calcolato, maxAttribute) se finito.
func (m *ModifiedAttribute) Compute(maxAttribute *float64) float64 {
	if m.hasCache {
		return m.cache
	}

	// Una lista per operazione, in ordine di inserimento (equivale al Map<op, list> del TS:
	// il TS legge sempre una singola operazione per volta).
	var byOp [9][]*types.ModifierAffliction
	for i := range m.Afflictions {
		a := &m.Afflictions[i]
		if k := opIndex(a.Operation); k >= 0 {
			byOp[k] = append(byOp[k], a)
		}
	}

	value := m.Base

	// 1. PreAssign: vince l'ultimo.
	if preAssign := byOp[opPreAssign]; len(preAssign) > 0 {
		value = preAssign[len(preAssign)-1].Value
	}

	// 2. Fase pre-moltiplicativa (penalizzata). PreDiv diventa moltiplicazione inversa.
	var preMulEntries []StackEntry
	for _, a := range byOp[opPreMul] {
		preMulEntries = append(preMulEntries, StackEntry{Value: a.Value, StackingGroup: a.StackingGroup})
	}
	for _, a := range byOp[opPreDiv] {
		preMulEntries = append(preMulEntries, StackEntry{Value: 1 / a.Value, StackingGroup: a.StackingGroup})
	}
	if len(preMulEntries) > 0 {
		value *= CombineMultiplicative(preMulEntries)
	}

	// 3. Fase additiva: ADDITIVE_OPS = ModAdd, ModSub (ModSub sottrae).
	for _, a := range byOp[opModAdd] {
		value += a.Value
	}
	for _, a := range byOp[opModSub] {
		value += -a.Value
	}

	// 4. Fase post-moltiplicativa: PostMul, PostDiv (invertito) e PostPercent
	// condividono un unico pool per gruppo di penalita' (vedi modifiedAttribute.ts).
	var postMulEntries []StackEntry
	for _, a := range byOp[opPostMul] {
		postMulEntries = append(postMulEntries, StackEntry{Value: a.Value, StackingGroup: a.StackingGroup})
	}
	for _, a := range byOp[opPostDiv] {
		postMulEntries = append(postMulEntries, StackEntry{Value: 1 / a.Value, StackingGroup: a.StackingGroup})
	}
	for _, a := range byOp[opPostPercent] {
		postMulEntries = append(postMulEntries, StackEntry{Value: 1 + a.Value, StackingGroup: a.StackingGroup})
	}
	if len(postMulEntries) > 0 {
		value *= CombineMultiplicative(postMulEntries)
	}

	// 5. PostAssign: vince l'ultimo e sovrascrive tutto.
	if postAssign := byOp[opPostAssign]; len(postAssign) > 0 {
		value = postAssign[len(postAssign)-1].Value
	}

	// 6. Tetto opzionale (Number.isFinite).
	if maxAttribute != nil && !math.IsInf(*maxAttribute, 0) && !math.IsNaN(*maxAttribute) {
		value = jsmath.Min(value, *maxAttribute)
	}

	m.cache = value
	m.hasCache = true
	return value
}

// ComputedAttribute e' il risultato di Snapshot (ComputedAttribute di types.ts).
type ComputedAttribute struct {
	ID          int
	Base        float64
	Final       float64
	Afflictions []types.ModifierAffliction
}

// Snapshot calcola e impacchetta; la slice delle afflictions e' condivisa.
func (m *ModifiedAttribute) Snapshot(maxAttribute *float64) ComputedAttribute {
	return ComputedAttribute{
		ID:          m.AttributeID,
		Base:        m.Base,
		Final:       m.Compute(maxAttribute),
		Afflictions: m.Afflictions,
	}
}

const (
	opPreAssign = iota
	opPreMul
	opPreDiv
	opModAdd
	opModSub
	opPostMul
	opPostDiv
	opPostPercent
	opPostAssign
)

func opIndex(op types.ModifierOperation) int {
	switch op {
	case types.OpPreAssign:
		return opPreAssign
	case types.OpPreMul:
		return opPreMul
	case types.OpPreDiv:
		return opPreDiv
	case types.OpModAdd:
		return opModAdd
	case types.OpModSub:
		return opModSub
	case types.OpPostMul:
		return opPostMul
	case types.OpPostDiv:
		return opPostDiv
	case types.OpPostPercent:
		return opPostPercent
	case types.OpPostAssign:
		return opPostAssign
	}
	return -1
}
