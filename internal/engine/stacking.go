// Porting letterale di src/stacking.ts: matematica della penalita' di stacking.
package engine

import (
	"math"
	"sort"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
)

// StackableValue e' StackableValue: moltiplicatore grezzo (1.05 = +5%, 0.9 = -10%).
type StackableValue struct {
	Value float64
}

// StackEntry e' l'elemento `{ value, stackingGroup }` di combineMultiplicative;
// StackingGroup nil = nessuna penalita'.
type StackEntry struct {
	Value         float64
	StackingGroup *string
}

// CombinePenalized riduce i moltiplicatori di uno stesso gruppo a un fattore; 1 se vuoto.
func CombinePenalized(values []StackableValue) float64 {
	if len(values) == 0 {
		return 1
	}

	var bonuses, penalties, neutrals []StackableValue
	for _, v := range values {
		if v.Value > 1 {
			bonuses = append(bonuses, v)
		}
	}
	for _, v := range values {
		if v.Value < 1 {
			penalties = append(penalties, v)
		}
	}
	for _, v := range values {
		if v.Value == 1 {
			neutrals = append(neutrals, v)
		}
	}

	// Ordine per deviazione assoluta da 1, decrescente. Array.prototype.sort e'
	// stabile; un NaN del comparatore vale 0.
	byDeviation := func(s []StackableValue) {
		sort.SliceStable(s, func(i, j int) bool {
			a, b := s[i], s[j]
			return jsmath.CmpResult(math.Abs(b.Value-1)-math.Abs(a.Value-1)) < 0
		})
	}
	byDeviation(bonuses)
	byDeviation(penalties)

	factor := 1.0

	for i := 0; i < len(bonuses); i++ {
		raw := bonuses[i].Value
		fi := float64(i)
		attenuation := math.Exp(-(fi * fi) / StackingPenaltyK)
		factor *= 1 + float64((raw-1)*attenuation)
	}

	for i := 0; i < len(penalties); i++ {
		raw := penalties[i].Value
		fi := float64(i)
		attenuation := math.Exp(-(fi * fi) / StackingPenaltyK)
		factor *= 1 + float64((raw-1)*attenuation)
	}

	for range neutrals { // no-op esplicito, come nel TS
		factor *= 1
	}

	return factor
}

// CombineUnstacked: prodotto semplice, nessuna penalita'.
func CombineUnstacked(values []StackableValue) float64 {
	factor := 1.0
	for _, v := range values {
		factor *= v.Value
	}
	return factor
}

// CombineMultiplicative raggruppa per chiave di stacking (ordine di inserimento)
// e combina ogni gruppo; StackingGroup nil = moltiplicazione diretta.
func CombineMultiplicative(entries []StackEntry) float64 {
	if len(entries) == 0 {
		return 1
	}
	groups := omap.New[string, []StackableValue]()
	var unstacked []StackableValue
	for _, e := range entries {
		if e.StackingGroup == nil {
			unstacked = append(unstacked, StackableValue{Value: e.Value})
		} else {
			list, _ := groups.Get(*e.StackingGroup)
			list = append(list, StackableValue{Value: e.Value})
			groups.Set(*e.StackingGroup, list)
		}
	}
	factor := CombineUnstacked(unstacked)
	for i := 0; i < groups.Len(); i++ {
		_, list := groups.At(i)
		factor *= CombinePenalized(list)
	}
	return factor
}
