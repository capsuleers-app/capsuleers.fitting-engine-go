package jsmath

import (
	"math"
	"testing"
)

func same(a, b float64) bool {
	if math.IsNaN(a) || math.IsNaN(b) {
		return math.IsNaN(a) && math.IsNaN(b)
	}
	return math.Float64bits(a) == math.Float64bits(b)
}

// Casi limite noti, sempre eseguiti (i golden completi su Node sono stati verificati all'estrazione del porting).
func TestRound(t *testing.T) {
	negZero := math.Copysign(0, -1)
	cases := []struct{ in, want float64 }{
		{0.5, 1}, {-0.5, negZero}, {1.5, 2}, {-1.5, -1}, {2.5, 3}, {-2.5, -2},
		{0.49999999999999994, 0}, {-0.49999999999999994, negZero}, {0.5000000000000001, 1},
		{negZero, negZero}, {0, 0}, {-0.1, negZero},
		{4503599627370495.5, 4503599627370496}, {4503599627370497, 4503599627370497},
		{9007199254740993, 9007199254740992}, {1e21, 1e21}, {math.Inf(-1), math.Inf(-1)}, {math.NaN(), math.NaN()},
	}
	for _, c := range cases {
		if got := Round(c.in); !same(got, c.want) {
			t.Errorf("Round(%v) = %v, atteso %v", c.in, got, c.want)
		}
	}
}

func TestToNumber(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"", 0}, {"  \t\n", 0}, {" 12 ", 12}, {"\u00a012\ufeff", 12}, {"0x1F", 31}, {"0b101", 5}, {"0o17", 15},
		{"-0x1", math.NaN()}, {"0x", math.NaN()}, {"1_0", math.NaN()}, {"1e", math.NaN()}, {".5", 0.5}, {"5.", 5},
		{".", math.NaN()}, {"-Infinity", math.Inf(-1)}, {"infinity", math.NaN()}, {"1e1000", math.Inf(1)},
		{"-0", math.Copysign(0, -1)}, {"\u200b1", math.NaN()}, {"12abc", math.NaN()},
	}
	for _, c := range cases {
		if got := ToNumber(c.in); !same(got, c.want) {
			t.Errorf("ToNumber(%q) = %v, atteso %v", c.in, got, c.want)
		}
	}
}

func TestParseFloatParseInt(t *testing.T) {
	if got := ParseFloat("  3.5e2xyz"); got != 350 {
		t.Errorf("ParseFloat prefisso: %v", got)
	}
	if got := ParseFloat("0x10"); got != 0 {
		t.Errorf("ParseFloat(0x10) = %v, atteso 0", got)
	}
	if got := ParseFloat("1e+"); got != 1 {
		t.Errorf("ParseFloat(1e+) = %v", got)
	}
	if !math.IsNaN(ParseFloat("-.e1")) {
		t.Error("ParseFloat(-.e1) deve essere NaN")
	}
	if got := ParseInt("0x1f", 0); got != 31 {
		t.Errorf("ParseInt(0x1f) = %v", got)
	}
	if got := ParseInt("12.9px", 0); got != 12 {
		t.Errorf("ParseInt(12.9px) = %v", got)
	}
	if got := ParseInt("-0", 10); !same(got, math.Copysign(0, -1)) {
		t.Errorf("ParseInt(-0) = %v", got)
	}
	if !math.IsNaN(ParseInt("z", 10)) || !math.IsNaN(ParseInt("1", 1)) || !math.IsNaN(ParseInt("1", 37)) {
		t.Error("ParseInt: NaN atteso")
	}
}

func TestCompare(t *testing.T) {
	if CompareUTF16("\uffff", "\U00010000") != 1 {
		t.Error("U+FFFF deve seguire la coppia surrogata di U+10000 (unita' UTF-16)")
	}
	if CompareUTF16("a", "ab") != -1 || CompareUTF16("ab", "ab") != 0 {
		t.Error("CompareUTF16 prefisso/uguaglianza")
	}
	if LocaleCompare("a", "B") != -1 || LocaleCompare("A", "a") != 1 || LocaleCompare("a-b", "ab") != -1 || LocaleCompare("9", "a") != -1 {
		t.Error("LocaleCompare ordine ICU")
	}
}

func TestCmpResult(t *testing.T) {
	if CmpResult(math.NaN()) != 0 || CmpResult(-3) != -1 || CmpResult(2) != 1 || CmpResult(0) != 0 {
		t.Error("CmpResult")
	}
}

func TestMinMax(t *testing.T) {
	nz := math.Copysign(0, -1)
	if !math.IsNaN(Max(math.Inf(1), math.NaN())) || !math.IsNaN(Min(math.Inf(-1), math.NaN())) {
		t.Error("NaN deve propagarsi anche con Inf (math.Max/Min no)")
	}
	if !same(Max(nz, 0), 0) || !same(Min(0, nz), nz) || !same(Max(nz, nz), nz) {
		t.Error("±0 in Min/Max")
	}
}
