package fitengine

import (
	"fmt"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
)

// entriesAny e' implementato da omap.Map.
type entriesAny interface{ EntriesAny() [][2]any }

// EncodeJSON riproduce JSON.stringify del valore equivalente in TS (eve-fit-engine): numeri nel formato ES
// (-0 -> 0, NaN/+-Inf -> null), campi opzionali nil omessi (undefined) salvo quelli col tag `encode:"null"`
// (null esplicito), nomi in lowerCamelCase, campi nell'ordine della struct (= ordine dei letterali TS),
// omap.Map come oggetto con le chiavi nell'ordine JS (indici di array ascendenti, poi inserimento).
// La forma che un consumatore da' alle Map resta una sua scelta: questa e' Object.fromEntries(map).
func EncodeJSON(v any) ([]byte, error) {
	var sb strings.Builder
	if err := encodeValue(&sb, reflect.ValueOf(v)); err != nil {
		return nil, err
	}
	return []byte(sb.String()), nil
}

func jsonFieldName(name string) string {
	if name == "ID" {
		return "id"
	}
	if name == strings.ToUpper(name) {
		return name // HI, MED, ...
	}
	return strings.ToLower(name[:1]) + name[1:]
}

// writeJSString replica la serializzazione di stringa di JSON.stringify (ES2019: JSON ben formato).
func writeJSString(sb *strings.Builder, s string) {
	sb.WriteByte('"')
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == '"':
			sb.WriteString(`\"`)
		case r == '\\':
			sb.WriteString(`\\`)
		case r == '\b':
			sb.WriteString(`\b`)
		case r == '\f':
			sb.WriteString(`\f`)
		case r == '\n':
			sb.WriteString(`\n`)
		case r == '\r':
			sb.WriteString(`\r`)
		case r == '\t':
			sb.WriteString(`\t`)
		case r < 0x20:
			sb.WriteString(fmt.Sprintf(`\u%04x`, r))
		case r == utf8.RuneError && size == 1:
			sb.WriteString(`�`) // byte non valido ~ surrogate isolato: U+FFFD
		default:
			sb.WriteString(s[i : i+size])
		}
		i += size
	}
	sb.WriteByte('"')
}

func writeNumber(sb *strings.Builder, f float64) {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		sb.WriteString("null")
		return
	}
	sb.WriteString(jsmath.NumberToString(f))
}

// jsKeyOrder ordina le chiavi come un oggetto JS: indici di array canonici ascendenti, poi le altre in
// ordine di inserimento.
func jsKeyOrder(keys []string) []int {
	var ints, others []int
	for i, k := range keys {
		if n, err := strconv.ParseUint(k, 10, 32); err == nil && n <= 4294967294 && strconv.FormatUint(n, 10) == k {
			ints = append(ints, i)
		} else {
			others = append(others, i)
		}
	}
	sort.SliceStable(ints, func(a, b int) bool {
		x, _ := strconv.ParseUint(keys[ints[a]], 10, 32)
		y, _ := strconv.ParseUint(keys[ints[b]], 10, 32)
		return x < y
	})
	return append(ints, others...)
}

func encodeValue(sb *strings.Builder, rv reflect.Value) error {
	if !rv.IsValid() {
		sb.WriteString("null")
		return nil
	}
	if rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			sb.WriteString("null")
			return nil
		}
		if e, ok := rv.Interface().(entriesAny); ok {
			return encodeEntries(sb, e)
		}
		return encodeValue(sb, rv.Elem())
	}
	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		writeNumber(sb, rv.Float())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		sb.WriteString(strconv.FormatInt(rv.Int(), 10))
	case reflect.String:
		writeJSString(sb, rv.String())
	case reflect.Bool:
		sb.WriteString(strconv.FormatBool(rv.Bool()))
	case reflect.Slice, reflect.Array:
		sb.WriteByte('[')
		for i := 0; i < rv.Len(); i++ {
			if i > 0 {
				sb.WriteByte(',')
			}
			if err := encodeValue(sb, rv.Index(i)); err != nil {
				return err
			}
		}
		sb.WriteByte(']')
	case reflect.Struct:
		sb.WriteByte('{')
		first := true
		t := rv.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fv := rv.Field(i)
			if fv.Kind() == reflect.Ptr && fv.IsNil() && f.Tag.Get("encode") != "null" {
				continue // undefined
			}
			if !first {
				sb.WriteByte(',')
			}
			first = false
			writeJSString(sb, jsonFieldName(f.Name))
			sb.WriteByte(':')
			if err := encodeValue(sb, fv); err != nil {
				return err
			}
		}
		sb.WriteByte('}')
	default:
		return fmt.Errorf("EncodeJSON: tipo non supportato %s", rv.Kind())
	}
	return nil
}

func encodeEntries(sb *strings.Builder, e entriesAny) error {
	entries := e.EntriesAny()
	keys := make([]string, len(entries))
	for i, kv := range entries {
		keys[i] = fmt.Sprint(kv[0])
	}
	sb.WriteByte('{')
	for n, i := range jsKeyOrder(keys) {
		if n > 0 {
			sb.WriteByte(',')
		}
		writeJSString(sb, keys[i])
		sb.WriteByte(':')
		if err := encodeValue(sb, reflect.ValueOf(entries[i][1])); err != nil {
			return err
		}
	}
	sb.WriteByte('}')
	return nil
}
