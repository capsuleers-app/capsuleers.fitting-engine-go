// Package jsmath replica la semantica JavaScript dove Go differisce per
// definizione (non per precisione): Math.round, Number(string), parseFloat,
// parseInt, confronto di stringhe per unita' UTF-16, localeCompare sull'alfabeto
// ASCII stampabile. Math.exp/log/pow NON sono qui: si usa `math` (tolleranza 1e-9).
package jsmath

import "math"

// Round e' Math.round: metà verso +Inf (math.Round va via da zero), -0 per x in
// [-0.5, -0), NaN/±Inf/±0 invariati, corretto su 0.49999999999999994 e oltre 2^52.
func Round(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) || x == 0 {
		return x
	}
	// A partire da 2^52 ogni double e' gia' un intero.
	if x >= 1<<52 || x <= -(1<<52) {
		return x
	}
	r := math.Floor(x)
	// x-r e' esatto per |x| < 2^52, a differenza di floor(x+0.5).
	if x-r >= 0.5 {
		r++
	}
	if r == 0 && x < 0 {
		return math.Copysign(0, -1)
	}
	return r
}

// CmpResult riduce il valore di un comparatore JS (a-b) a -1/0/1: NaN vale 0,
// come impone la specifica di Array.prototype.sort.
func CmpResult(v float64) int {
	switch {
	case v < 0:
		return -1
	case v > 0:
		return 1
	}
	return 0
}

// Max e' Math.max(a, b): NaN se uno dei due e' NaN (anche con +Inf, a differenza di
// math.Max), e +0 > -0.
func Max(a, b float64) float64 {
	switch {
	case math.IsNaN(a) || math.IsNaN(b):
		return math.NaN()
	case a == 0 && b == 0:
		if math.Signbit(a) {
			return b
		}
		return a
	case a > b:
		return a
	}
	return b
}

// Min e' Math.min(a, b): NaN se uno dei due e' NaN, e -0 < +0.
func Min(a, b float64) float64 {
	switch {
	case math.IsNaN(a) || math.IsNaN(b):
		return math.NaN()
	case a == 0 && b == 0:
		if math.Signbit(a) {
			return a
		}
		return b
	case a < b:
		return a
	}
	return b
}
