// Package export e' il porting letterale di src/export.ts: DNA, multibuy e elenco di typeID.
package export

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

type counter struct{ m *omap.Map[int, float64] }

func newCounter() *counter { return &counter{m: omap.New[int, float64]()} }

func (c *counter) bump(id int, n float64) {
	cur, _ := c.m.Get(id)
	c.m.Set(id, cur+n)
}

func hasCharge(id *int) bool { return id != nil && *id != 0 }

// FormatDna: formato DNA di EVE (`ship:type;count:...::`).
func FormatDna(fit *types.Fit) string {
	var sb strings.Builder
	sb.WriteString(strconv.Itoa(fit.ShipTypeID) + ":")
	counts := newCounter()
	for _, m := range fit.Modules {
		counts.bump(m.TypeID, 1)
		if hasCharge(m.ChargeTypeID) {
			counts.bump(*m.ChargeTypeID, 1)
		}
	}
	for _, d := range fit.Drones {
		counts.bump(d.TypeID, d.CountTotal)
	}
	for _, f := range fit.Fighters {
		counts.bump(f.TypeID, f.Count)
	}
	for _, c := range fit.Cargo {
		counts.bump(c.TypeID, c.Count)
	}
	for _, i := range fit.Implants {
		counts.bump(i.TypeID, 1)
	}
	for _, b := range fit.Boosters {
		counts.bump(b.TypeID, 1)
	}
	for _, s := range fit.Subsystems {
		counts.bump(s.TypeID, 1)
	}
	for i := 0; i < counts.m.Len(); i++ {
		id, count := counts.m.At(i)
		sb.WriteString(fmt.Sprintf("%d;%s:", id, jsmath.NumberToString(count)))
	}
	return sb.String() + ":"
}

// FormatMultibuy: elenco "Nome xN" ordinato per unita' UTF-16 (sort() di JS).
func FormatMultibuy(fit *types.Fit, dataset *types.FittingDataset) string {
	counts := newCounter()
	counts.bump(fit.ShipTypeID, 1)
	for _, m := range fit.Modules {
		counts.bump(m.TypeID, 1)
		if hasCharge(m.ChargeTypeID) {
			counts.bump(*m.ChargeTypeID, 1)
		}
	}
	for _, d := range fit.Drones {
		counts.bump(d.TypeID, d.CountTotal)
	}
	for _, f := range fit.Fighters {
		counts.bump(f.TypeID, f.Count)
	}
	for _, c := range fit.Cargo {
		counts.bump(c.TypeID, c.Count)
	}
	for _, i := range fit.Implants {
		counts.bump(i.TypeID, 1)
	}
	for _, b := range fit.Boosters {
		counts.bump(b.TypeID, 1)
	}
	for _, s := range fit.Subsystems {
		counts.bump(s.TypeID, 1)
	}
	var lines []string
	for i := 0; i < counts.m.Len(); i++ {
		id, count := counts.m.At(i)
		name := fmt.Sprintf("Type %d", id)
		if t, ok := dataset.GetType(id); ok {
			name = t.Name
		}
		lines = append(lines, name+" x"+jsmath.NumberToString(count))
	}
	sort.SliceStable(lines, func(i, j int) bool { return jsmath.CompareUTF16(lines[i], lines[j]) < 0 })
	return strings.Join(lines, "\n")
}

// FormatTypeIds: typeID distinti, ascendenti, uno per riga.
func FormatTypeIds(fit *types.Fit) string {
	ids := omap.NewSet[int]()
	ids.Add(fit.ShipTypeID)
	for _, m := range fit.Modules {
		ids.Add(m.TypeID)
		if hasCharge(m.ChargeTypeID) {
			ids.Add(*m.ChargeTypeID)
		}
	}
	for _, d := range fit.Drones {
		ids.Add(d.TypeID)
	}
	for _, f := range fit.Fighters {
		ids.Add(f.TypeID)
	}
	for _, c := range fit.Cargo {
		ids.Add(c.TypeID)
	}
	for _, i := range fit.Implants {
		ids.Add(i.TypeID)
	}
	for _, b := range fit.Boosters {
		ids.Add(b.TypeID)
	}
	for _, s := range fit.Subsystems {
		ids.Add(s.TypeID)
	}
	vals := ids.Values()
	sort.Ints(vals)
	out := make([]string, len(vals))
	for i, v := range vals {
		out[i] = strconv.Itoa(v)
	}
	return strings.Join(out, "\n")
}
