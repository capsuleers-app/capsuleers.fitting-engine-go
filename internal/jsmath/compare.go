package jsmath

import "unicode/utf16"

// CompareUTF16 confronta due stringhe per unita' di codice UTF-16, come gli
// operatori < e > di JS (non per byte UTF-8: differiscono sopra U+FFFF contro
// U+E000..U+FFFF). Restituisce -1, 0 o 1.
func CompareUTF16(a, b string) int {
	ua := utf16.Encode([]rune(a))
	ub := utf16.Encode([]rune(b))
	n := len(ua)
	if len(ub) < n {
		n = len(ub)
	}
	for i := 0; i < n; i++ {
		if ua[i] != ub[i] {
			if ua[i] < ub[i] {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(ua) < len(ub):
		return -1
	case len(ua) > len(ub):
		return 1
	}
	return 0
}

// localeASCIIOrder e' l'ordine ICU (collazione root, Node 26 / ICU 79.1,
// locale it-IT) dei 95 caratteri ASCII stampabili, con la coppia minuscola/
// maiuscola vicina (primario uguale, terziario minuscola < maiuscola). E' un
// dato misurato su Node (golden `localeCompare` del porting, verificato all'estrazione).
const localeASCIIOrder = " _-,;:!?.'\"()[]{}@*/\\&#%`^+<=>|~$0123456789aAbBcCdDeEfFgGhHiIjJkKlLmMnNoOpPqQrRsStTuUvVwWxXyYzZ"

var localePrimary, localeTertiary [128]int8

func init() {
	for i := range localePrimary {
		localePrimary[i] = -1
	}
	rank := int8(-1)
	var prev byte
	for i := 0; i < len(localeASCIIOrder); i++ {
		c := localeASCIIOrder[i]
		lower := c
		if c >= 'A' && c <= 'Z' {
			lower = c + 32
		}
		if i == 0 || lower != prev {
			rank++
			prev = lower
		}
		localePrimary[c] = rank
		if c >= 'A' && c <= 'Z' {
			localeTertiary[c] = 1
		}
	}
}

func localeASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] > 0x7E {
			return false
		}
	}
	return true
}

// LocaleCompare replica a.localeCompare(b) di Node per stringhe ASCII
// stampabili (0x20..0x7E): confronto primario sull'intera stringa
// (punteggiatura < cifre < lettere senza distinguere il caso), poi terziario
// (minuscola prima di maiuscola). Fuori da quell'alfabeto ricade su
// CompareUTF16: usato solo per gli id dei moduli di un fit (uuid, cuid, interi
// in stringa), che stanno nell'alfabeto coperto.
func LocaleCompare(a, b string) int {
	if !localeASCII(a) || !localeASCII(b) {
		return CompareUTF16(a, b)
	}
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if pa, pb := localePrimary[a[i]], localePrimary[b[i]]; pa != pb {
			if pa < pb {
				return -1
			}
			return 1
		}
	}
	if len(a) != len(b) {
		if len(a) < len(b) {
			return -1
		}
		return 1
	}
	for i := 0; i < n; i++ {
		if ta, tb := localeTertiary[a[i]], localeTertiary[b[i]]; ta != tb {
			if ta < tb {
				return -1
			}
			return 1
		}
	}
	return 0
}
