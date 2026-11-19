package jsmath

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// ToFixed e' Number.prototype.toFixed(f), 0 <= f <= 100: arrotonda il valore ESATTO del double a f
// decimali scegliendo il n piu' grande nei pareggi (half-up sul valore esatto, a differenza di
// strconv che arrotonda al pari). NaN -> "NaN"; |x| >= 1e21 -> Number::toString (formato 'g' di JS).
func ToFixed(x float64, f int) string {
	if math.IsNaN(x) {
		return "NaN"
	}
	if math.IsInf(x, 0) || math.Abs(x) >= 1e21 {
		return NumberToString(x)
	}
	neg := x < 0
	if neg {
		x = -x
	}
	if x == 0 {
		x = 0 // -0 -> +0
	}
	exact := new(big.Float).SetFloat64(x).Text('f', 1100) // espansione decimale esatta
	dot := strings.IndexByte(exact, '.')
	intPart, frac := exact[:dot], exact[dot+1:]
	digits := []byte(intPart + frac[:f])
	if frac[f] >= '5' {
		i := len(digits) - 1
		for ; i >= 0; i-- {
			if digits[i] == '9' {
				digits[i] = '0'
				continue
			}
			digits[i]++
			break
		}
		if i < 0 {
			digits = append([]byte{'1'}, digits...)
		}
	}
	n := len(digits) - f
	s := string(digits[:n])
	if f > 0 {
		s += "." + string(digits[n:])
	}
	if neg {
		s = "-" + s
	}
	return s
}

// NumberToString e' String(number) di JS (Number::toString, radice 10): forma piu' corta che fa
// il round trip; esponenziale se l'esponente decimale k < -6 o >= 21, "e+"/"e-" senza zeri di riempimento.
func NumberToString(x float64) string {
	switch {
	case math.IsNaN(x):
		return "NaN"
	case x == 0:
		return "0"
	case math.IsInf(x, 1):
		return "Infinity"
	case math.IsInf(x, -1):
		return "-Infinity"
	}
	sign := ""
	if x < 0 {
		sign, x = "-", -x
	}
	// cifre significative e esponente dalla rappresentazione piu' corta
	e := strconv.FormatFloat(x, 'e', -1, 64) // d.ddddde+XX
	mant, expStr, _ := strings.Cut(e, "e")
	exp, _ := strconv.Atoi(expStr)
	digits := strings.Replace(mant, ".", "", 1)
	k := len(digits)
	n := exp + 1 // posizione della virgola rispetto alle cifre
	switch {
	case k <= n && n <= 21:
		return sign + digits + strings.Repeat("0", n-k)
	case 0 < n && n <= 21:
		return sign + digits[:n] + "." + digits[n:]
	case -6 < n && n <= 0:
		return sign + "0." + strings.Repeat("0", -n) + digits
	}
	es := "e+"
	if n-1 < 0 {
		es = "e-"
	}
	ex := n - 1
	if ex < 0 {
		ex = -ex
	}
	if k == 1 {
		return sign + digits + es + strconv.Itoa(ex)
	}
	return sign + digits[:1] + "." + digits[1:] + es + strconv.Itoa(ex)
}
