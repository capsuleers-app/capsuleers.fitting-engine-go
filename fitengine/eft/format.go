package eft

import (
	"fmt"
	"sort"
	"strings"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

// typeName e' `t.name ?? "Type <id>"` (il nome e' sempre presente nel bundle).

func nameOrType(t *types.SdeType, id int) string {
	if t == nil {
		return fmt.Sprintf("Type %d", id)
	}
	return t.Name
}

// FormatEft esporta il fit in EFT: slot LO, MED, HI, RIG, poi subsystem, droni, impianti/booster/cargo.
func FormatEft(fit *types.Fit, dataset *types.FittingDataset) (string, error) {
	ship, ok := dataset.GetType(fit.ShipTypeID)
	if !ok {
		return "", fmt.Errorf("format-eft: unknown ship typeID %d", fit.ShipTypeID)
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("[%s, %s]", ship.Name, fit.Name))

	for _, slot := range [4]types.SlotType{types.SlotLO, types.SlotMED, types.SlotHI, types.SlotRIG} {
		var mods []types.FitModule
		for _, m := range fit.Modules {
			if m.SlotType == slot {
				mods = append(mods, m)
			}
		}
		sort.SliceStable(mods, func(i, j int) bool { return byPosition(&mods[i], &mods[j]) < 0 })
		if len(mods) == 0 {
			continue
		}
		lines = append(lines, "")
		for i := range mods {
			lines = append(lines, formatModule(&mods[i], dataset))
		}
	}

	if len(fit.Subsystems) > 0 {
		subs := append([]types.FitSubsystem(nil), fit.Subsystems...)
		sort.SliceStable(subs, func(i, j int) bool { return subs[i].Slot < subs[j].Slot })
		lines = append(lines, "")
		for _, s := range subs {
			if t, ok := dataset.GetType(s.TypeID); ok {
				lines = append(lines, t.Name)
			}
		}
	}

	if len(fit.Drones) > 0 {
		lines = append(lines, "")
		for _, d := range fit.Drones {
			t, ok := dataset.GetType(d.TypeID)
			if !ok {
				continue
			}
			lines = append(lines, t.Name+" x"+jsmath.NumberToString(d.CountTotal))
		}
	}

	var trailing []string
	for _, i := range fit.Implants {
		if t, ok := dataset.GetType(i.TypeID); ok {
			trailing = append(trailing, t.Name+" x1")
		}
	}
	for _, b := range fit.Boosters {
		if t, ok := dataset.GetType(b.TypeID); ok {
			trailing = append(trailing, t.Name+" x1")
		}
	}
	for _, c := range fit.Cargo {
		if t, ok := dataset.GetType(c.TypeID); ok {
			trailing = append(trailing, t.Name+" x"+jsmath.NumberToString(c.Count))
		}
	}
	if len(trailing) > 0 {
		lines = append(lines, "")
		lines = append(lines, trailing...)
	}

	return strings.Join(lines, "\n") + "\n", nil
}

// byPosition: a.position - b.position || a.id.localeCompare(b.id), come comparatore di sort JS.
func byPosition(a, b *types.FitModule) int {
	d := a.Position - b.Position
	if d != 0 && d == d { // non zero e non NaN
		return jsmath.CmpResult(d)
	}
	return jsmath.LocaleCompare(a.ID, b.ID)
}

func formatModule(m *types.FitModule, dataset *types.FittingDataset) string {
	t, _ := dataset.GetType(m.TypeID)
	name := nameOrType(t, m.TypeID)
	if m.ChargeTypeID != nil && *m.ChargeTypeID != 0 {
		if ct, ok := dataset.GetType(*m.ChargeTypeID); ok && ct.Name != "" {
			return name + ", " + ct.Name
		}
	}
	return name
}
