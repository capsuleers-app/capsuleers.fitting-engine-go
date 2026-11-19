package jsmath

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

// isJSSpace: WhiteSpace + LineTerminator della specifica ECMAScript
// (TAB VT FF SP NBSP ZWNBSP, categoria Zs, LF CR LS PS).
func isJSSpace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0xA0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return r >= 0x2000 && r <= 0x200A
}

func trimLeftJS(s string) string { return strings.TrimLeftFunc(s, isJSSpace) }
func trimJS(s string) string     { return strings.TrimFunc(s, isJSSpace) }
func isDigit(c byte) bool        { return c >= '0' && c <= '9' }
func nan() float64               { return math.NaN() }
func negZero() float64           { return math.Copysign(0, -1) }

// decimalPrefix restituisce la lunghezza del piu' lungo prefisso di s che e' uno
// StrDecimalLiteral JS: [+-]? ( Infinity | cifre[.cifre?] | .cifre ) ([eE][+-]?cifre)?
// (0 se nessuno). L'esponente incompleto non fa parte del prefisso.
func decimalPrefix(s string) int {
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	if strings.HasPrefix(s[i:], "Infinity") {
		return i + len("Infinity")
	}
	start := i
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	intDigits := i - start
	fracDigits := 0
	if i < len(s) && s[i] == '.' {
		j := i + 1
		for j < len(s) && isDigit(s[j]) {
			j++
		}
		fracDigits = j - (i + 1)
		if intDigits > 0 || fracDigits > 0 {
			i = j
		}
	}
	if intDigits == 0 && fracDigits == 0 {
		return 0
	}
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		j := i + 1
		if j < len(s) && (s[j] == '+' || s[j] == '-') {
			j++
		}
		k := j
		for k < len(s) && isDigit(s[k]) {
			k++
		}
		if k > j {
			i = k
		}
	}
	return i
}

// parseDecimal converte un letterale decimale gia' validato (decimalPrefix).
func parseDecimal(lit string) float64 {
	neg := false
	body := lit
	if body[0] == '+' || body[0] == '-' {
		neg = body[0] == '-'
		body = body[1:]
	}
	if body == "Infinity" {
		if neg {
			return math.Inf(-1)
		}
		return math.Inf(1)
	}
	v, _ := strconv.ParseFloat(body, 64) // arrotondamento corretto; overflow -> +Inf (err ignorato)
	if neg {
		return -v
	}
	return v
}

// ParseFloat e' parseFloat(string): spazi iniziali, piu' lungo prefisso
// StrDecimalLiteral, altrimenti NaN. Niente hex.
func ParseFloat(s string) float64 {
	s = trimLeftJS(s)
	n := decimalPrefix(s)
	if n == 0 {
		return nan()
	}
	return parseDecimal(s[:n])
}

func radixDigit(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'z':
		return int(c-'a') + 10
	case c >= 'A' && c <= 'Z':
		return int(c-'A') + 10
	}
	return 99
}

// bigToFloat arrotonda l'intero esatto al double piu' vicino (pari sui pareggi).
func bigToFloat(digits string, radix int) float64 {
	i, ok := new(big.Int).SetString(digits, radix)
	if !ok {
		return nan()
	}
	f, _ := new(big.Float).SetInt(i).Float64()
	return f
}

// ToNumber e' Number(string): trim JS; "" -> 0; 0x/0o/0b senza segno; Infinity
// con segno; letterale decimale completo; altrimenti NaN.
func ToNumber(s string) float64 {
	s = trimJS(s)
	if s == "" {
		return 0
	}
	if len(s) > 2 && s[0] == '0' {
		radix := 0
		switch s[1] {
		case 'x', 'X':
			radix = 16
		case 'o', 'O':
			radix = 8
		case 'b', 'B':
			radix = 2
		}
		if radix != 0 {
			for j := 2; j < len(s); j++ {
				if radixDigit(s[j]) >= radix {
					return nan()
				}
			}
			return bigToFloat(s[2:], radix)
		}
	}
	if decimalPrefix(s) != len(s) {
		return nan()
	}
	return parseDecimal(s)
}

// ParseInt e' parseInt(string, radix) con radix 0 = "non passato / 0". Esatto per
// radix 10 e potenze di 2 (arrotondamento corretto) e per i risultati < 2^53 in
// ogni radice; per le altre radici oltre 2^53 V8 approssima e qui no (non usato
// dall'engine).
func ParseInt(s string, radix int) float64 {
	s = trimLeftJS(s)
	neg := false
	if s != "" && (s[0] == '+' || s[0] == '-') {
		neg = s[0] == '-'
		s = s[1:]
	}
	strip := true
	if radix != 0 {
		if radix < 2 || radix > 36 {
			return nan()
		}
		if radix != 16 {
			strip = false
		}
	} else {
		radix = 10
	}
	if strip && len(s) >= 2 && s[0] == '0' && (s[1] == 'x' || s[1] == 'X') {
		s = s[2:]
		radix = 16
	}
	end := 0
	for end < len(s) && radixDigit(s[end]) < radix {
		end++
	}
	if end == 0 {
		return nan()
	}
	var v float64
	if radix == 10 {
		v, _ = strconv.ParseFloat(s[:end], 64)
	} else {
		v = bigToFloat(s[:end], radix)
	}
	if neg {
		return -v // "-0" -> -0, come JS
	}
	return v
}

// TrimJS e' String.prototype.trim (WhiteSpace e LineTerminator di ECMAScript).
func TrimJS(s string) string { return trimJS(s) }
