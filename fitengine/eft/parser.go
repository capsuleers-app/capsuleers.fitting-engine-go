// Package eft e' il porting letterale di src/eft/parser.ts e format.ts: parsing e formattazione EFT.
package eft

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

// Warning e' un avviso di parsing (riga 1-based, testo, motivo).
type Warning struct {
	Line   int
	Text   string
	Reason string
}

// ParseResult e' EftParseResult.
type ParseResult struct {
	Fit      *types.Fit
	Warnings []Warning
}

// NameIndex e' l'indice nome (minuscolo) -> typeID e typeID -> type. Solo lookup per chiave.
type NameIndex struct {
	Map   map[string]int
	Types map[int]*types.SdeType
}

// BuildNameIndex indicizza ogni type di ogni bucket (nell'ordine dei bucket e per id): in caso di
// collisione di nome vince l'ultimo.
func BuildNameIndex(dataset *types.FittingDataset) *NameIndex {
	idx := &NameIndex{Map: map[string]int{}, Types: map[int]*types.SdeType{}}
	for b := types.TypeBucket(0); b < types.NumBuckets; b++ {
		bucket := &dataset.TypesByBucket[b]
		for i := 0; i < bucket.Len(); i++ {
			_, t := bucket.At(i)
			idx.Types[t.ID] = t
			if t.Name != "" {
				idx.Map[strings.ToLower(t.Name)] = t.ID
			}
		}
	}
	return idx
}

// \s e . di JS (non quelli di RE2): spazi Unicode e terminatori di riga.
const (
	jsSpace = `[\t\n\v\f\r \x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}\x{feff}]`
	jsDot   = `[^\n\r\x{2028}\x{2029}]`
)

var (
	headerRe   = regexp.MustCompile(`^\[(` + jsDot + `+?),` + jsSpace + `*(` + jsDot + `+?)\]` + jsSpace + `*$`)
	quantityRe = regexp.MustCompile(`(?i)` + jsSpace + `x([0-9]+)` + jsSpace + `*$`)
)

func isEmptyPlaceholder(lower string) bool {
	switch lower {
	case "[empty low slot]", "[empty med slot]", "[empty high slot]", "[empty rig slot]", "[empty subsystem slot]":
		return true
	}
	return false
}

// subsystemGroupToSlot e' SUBSYSTEM_GROUP_TO_SLOT.
func subsystemGroupToSlot(groupID int) (int, bool) {
	switch groupID {
	case 954:
		return 1, true
	case 956:
		return 2, true
	case 957:
		return 3, true
	case 958:
		return 4, true
	}
	return 0, false
}

type entry struct {
	line int
	text string
}

type section struct {
	lineNo  int
	entries []entry
}

// splitLines e' text.split(/\r?\n/).
func splitLines(text string) []string {
	parts := strings.Split(text, "\n")
	for i, p := range parts {
		parts[i] = strings.TrimSuffix(p, "\r")
	}
	return parts
}

// ParseEft interpreta un fit in formato EFT. Errore: input vuoto, header mancante, nave sconosciuta.
func ParseEft(text string, dataset *types.FittingDataset) (*ParseResult, error) {
	lines := splitLines(text)
	var warnings []Warning

	i := 0
	for i < len(lines) && jsmath.TrimJS(lines[i]) == "" {
		i++
	}
	if i >= len(lines) {
		return nil, errors.New("EFT: empty input")
	}
	headerLine := lines[i]
	m := headerRe.FindStringSubmatch(headerLine)
	if m == nil {
		return nil, fmt.Errorf("EFT: missing or malformed header on line %d: %s", i+1, jsmath.TrimJS(headerLine))
	}
	i++
	shipName := jsmath.TrimJS(m[1])
	fitName := jsmath.TrimJS(m[2])

	idx := BuildNameIndex(dataset)
	shipID, ok := idx.Map[strings.ToLower(shipName)]
	if !ok {
		return nil, fmt.Errorf("EFT: unknown ship \"%s\"", shipName)
	}
	shipType := idx.Types[shipID]
	if shipType.CategoryID != engine.CategoryShip {
		return nil, fmt.Errorf("EFT: header ship \"%s\" is not a Ship category type", shipName)
	}

	// Sezioni separate da righe vuote.
	var sections []section
	var current *section
	for ; i < len(lines); i++ {
		trimmed := jsmath.TrimJS(lines[i])
		if trimmed == "" {
			if current != nil && len(current.entries) > 0 {
				sections = append(sections, *current)
				current = nil
			}
			continue
		}
		if current == nil {
			current = &section{lineNo: i + 1}
		}
		current.entries = append(current.entries, entry{i + 1, trimmed})
	}
	if current != nil && len(current.entries) > 0 {
		sections = append(sections, *current)
	}

	fit := &types.Fit{
		ShipTypeID: shipType.ID, Name: fitName, Visibility: "PRIVATE", Tags: []string{},
		Modules: []types.FitModule{}, Drones: []types.FitDrone{}, Fighters: []types.FitFighter{}, Cargo: []types.FitCargo{},
		Implants: []types.FitImplant{}, Boosters: []types.FitBooster{}, Subsystems: []types.FitSubsystem{},
	}
	slotCursor := map[types.SlotType]float64{} // solo lookup/incremento per chiave

	for _, sec := range sections {
		kind := classifySection(sec.entries, idx)
		for _, e := range sec.entries {
			consumeEntry(e, kind, fit, slotCursor, idx, dataset, &warnings)
		}
	}
	return &ParseResult{Fit: fit, Warnings: warnings}, nil
}

type sectionKind int

const (
	kindModules sectionKind = iota
	kindDrones
	kindCargo
	kindImplants
)

func classifySection(entries []entry, idx *NameIndex) sectionKind {
	sawImplant, allImplant := false, true
	for _, e := range entries {
		name, _, quantity := parseEntryName(e.text)
		typeID, ok := idx.Map[strings.ToLower(name)]
		if !ok {
			allImplant = false
			continue
		}
		t := idx.Types[typeID]
		if quantity == nil && isFittableImplant(t) {
			sawImplant = true
		} else {
			allImplant = false
		}
	}
	if sawImplant && allImplant {
		return kindImplants
	}
	for _, e := range entries {
		name, _, quantity := parseEntryName(e.text)
		typeID, ok := idx.Map[strings.ToLower(name)]
		if !ok {
			continue
		}
		t := idx.Types[typeID]
		if t.CategoryID == engine.CategoryDrone {
			return kindDrones
		}
		if quantity != nil && t.CategoryID == engine.CategoryModule {
			return kindCargo
		}
		if quantity != nil {
			return kindCargo
		}
		if t.CategoryID == engine.CategoryModule || t.CategoryID == engine.CategorySubsystem || t.CategoryID == engine.CategoryImplant {
			return kindModules
		}
	}
	return kindCargo // sezioni sconosciute
}

const (
	attrImplantness = 331
	attrBoosterness = 1087
)

func attrValue(t *types.SdeType, id int) (float64, bool) {
	for _, a := range t.Attributes {
		if a.ID == id {
			return a.V, true
		}
	}
	return 0, false
}

func isFittableImplant(t *types.SdeType) bool {
	if t.CategoryID != engine.CategoryImplant {
		return false
	}
	_, a := attrValue(t, attrImplantness)
	_, b := attrValue(t, attrBoosterness)
	return a || b
}

// parseEntryName: "<nome> xN" (quantita'), "<modulo>, <carica>" o solo nome. charge "" = assente;
// quantity nil = assente.
func parseEntryName(text string) (name, charge string, quantity *float64) {
	if loc := quantityRe.FindStringSubmatchIndex(text); loc != nil {
		q := jsmath.ToNumber(text[loc[2]:loc[3]])
		return jsmath.TrimJS(text[:loc[0]]), "", &q
	}
	if comma := strings.IndexByte(text, ','); comma >= 0 {
		return jsmath.TrimJS(text[:comma]), jsmath.TrimJS(text[comma+1:]), nil
	}
	return jsmath.TrimJS(text), "", nil
}

var tempCounter atomic.Int64

// tempID: id temporaneo `tmp:<n>:<timestamp base36>` (come tempId() del TS).
func tempID() string {
	n := tempCounter.Add(1)
	return "tmp:" + strconv.FormatInt(n, 10) + ":" + strconv.FormatInt(time.Now().UnixMilli(), 36)
}

func detectSlot(t *types.SdeType, dataset *types.FittingDataset) (types.SlotType, bool) {
	for _, e := range t.Effects {
		effect, ok := dataset.Effects.Get(e.ID)
		if !ok {
			continue
		}
		if slot, ok := engine.SlotEffectToSlotType(effect.ID); ok {
			return slot, true
		}
	}
	return "", false
}

func consumeEntry(e entry, kind sectionKind, fit *types.Fit, slotCursor map[types.SlotType]float64, idx *NameIndex, dataset *types.FittingDataset, warnings *[]Warning) {
	lower := strings.ToLower(e.text)
	if isEmptyPlaceholder(lower) {
		return
	}
	name, charge, quantity := parseEntryName(e.text)
	typeID, ok := idx.Map[strings.ToLower(name)]
	if !ok {
		*warnings = append(*warnings, Warning{e.line, e.text, fmt.Sprintf("unknown type \"%s\"", name)})
		return
	}
	t := idx.Types[typeID]

	if kind == kindDrones || (quantity != nil && t.CategoryID == engine.CategoryDrone) {
		total := 1.0
		if quantity != nil {
			total = *quantity
		}
		fit.Drones = append(fit.Drones, types.FitDrone{ID: tempID(), TypeID: typeID, CountTotal: total, CountActive: 0})
		return
	}
	if kind == kindImplants && quantity == nil && isFittableImplant(t) {
		if slot, ok := attrValue(t, attrImplantness); ok {
			fit.Implants = append(fit.Implants, types.FitImplant{ID: tempID(), TypeID: typeID, Slot: slot})
		} else {
			slot, ok := attrValue(t, attrBoosterness)
			if !ok {
				slot = 1
			}
			fit.Boosters = append(fit.Boosters, types.FitBooster{ID: tempID(), TypeID: typeID, Slot: slot, ActiveSideEffects: []int{}})
		}
		return
	}
	if kind == kindCargo || quantity != nil {
		count := 1.0
		if quantity != nil {
			count = *quantity
		}
		fit.Cargo = append(fit.Cargo, types.FitCargo{ID: tempID(), TypeID: typeID, Count: count})
		return
	}

	slot, ok := detectSlot(t, dataset)
	if !ok {
		*warnings = append(*warnings, Warning{e.line, e.text, fmt.Sprintf("cannot determine slot for \"%s\"", name)})
		return
	}

	var chargeTypeID *int
	if charge != "" {
		cid, ok := idx.Map[strings.ToLower(charge)]
		if !ok {
			*warnings = append(*warnings, Warning{e.line, e.text, fmt.Sprintf("unknown charge \"%s\"", charge)})
		} else {
			chargeTypeID = &cid
		}
	}

	if slot == types.SlotSUBSYSTEM {
		subSlot, ok := subsystemGroupToSlot(t.GroupID)
		if !ok {
			subSlot = int(slotCursor[types.SlotSUBSYSTEM]) + 1
		}
		fit.Subsystems = append(fit.Subsystems, types.FitSubsystem{ID: tempID(), Slot: subSlot, TypeID: typeID})
		slotCursor[types.SlotSUBSYSTEM]++
		return
	}

	pos := slotCursor[slot]
	slotCursor[slot]++
	fit.Modules = append(fit.Modules, types.FitModule{
		ID: tempID(), SlotType: slot, Position: pos, TypeID: typeID, State: types.StateOnline, ChargeTypeID: chargeTypeID,
	})
}
