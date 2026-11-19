package engine_test

// Effect coverage audit (port of `npm run audit:coverage` of eve-fit-engine): every SDE effect referenced
// by a fittable type must either carry data-driven modifierInfo (the generic dispatcher) OR be registered
// in LegacyEffectIDs / OutOfScopeEffectIDs, OR be an SDE marker read outside the dispatcher (slot routing,
// weapon dispatch, online gating). Headline: "truly silent (real gaps)" must be 0; it is part of the
// release gate, because a new CCP effect with empty modifierInfo would otherwise compute as nothing.
//
//	go test ./internal/engine -run TestEffectCoverage -v   # full report

import (
	"regexp"
	"sort"
	"testing"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine/bundle"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

var fittableBuckets = []types.TypeBucket{
	types.BucketShips, types.BucketModules, types.BucketCharges, types.BucketDrones, types.BucketFighters,
	types.BucketSubsystems, types.BucketSkills, types.BucketImplants, types.BucketSystemEffects,
	types.BucketStructures, types.BucketStructureModules,
}

func isSdeMarker(id int) bool {
	if _, ok := engine.SlotEffectToSlotType(id); ok {
		return true
	}
	if _, ok := engine.WeaponKindOf(id); ok {
		return true
	}
	return id == engine.ActivationEffectOnline
}

func TestEffectCoverage(t *testing.T) {
	ds, err := bundle.LoadDataset(bundle.FS())
	if err != nil {
		t.Fatal(err)
	}
	usage := map[int]map[int]bool{} // effectID -> typeIDs
	for _, b := range fittableBuckets {
		tbl := ds.TypesByBucket[b]
		for i := 0; i < tbl.Len(); i++ {
			_, ty := tbl.At(i)
			for _, e := range ty.Effects {
				if usage[e.ID] == nil {
					usage[e.ID] = map[int]bool{}
				}
				usage[e.ID][ty.ID] = true
			}
		}
	}
	legacy, oos := map[int]bool{}, map[int]bool{}
	for _, e := range engine.LegacyEffectIDs {
		legacy[e.ID] = true
	}
	for _, e := range engine.OutOfScopeEffectIDs {
		oos[e.ID] = true
	}

	var dataDriven, empty, unused, byMarker, byLegacy, byOos int
	var silent []*types.SdeEffect
	for i := 0; i < ds.Effects.Len(); i++ {
		_, eff := ds.Effects.At(i)
		if len(eff.ModifierInfo) > 0 {
			dataDriven++
			continue
		}
		empty++
		switch {
		case len(usage[eff.ID]) == 0:
			unused++
		case isSdeMarker(eff.ID):
			byMarker++
		case legacy[eff.ID]:
			byLegacy++
		case oos[eff.ID]:
			byOos++
		default:
			silent = append(silent, eff)
		}
	}
	sort.SliceStable(silent, func(i, j int) bool { return len(usage[silent[i].ID]) > len(usage[silent[j].ID]) })

	var legacyMissing, oosMissing []int
	for _, e := range engine.LegacyEffectIDs {
		if !ds.Effects.Has(e.ID) {
			legacyMissing = append(legacyMissing, e.ID)
		}
	}
	for _, e := range engine.OutOfScopeEffectIDs {
		if !ds.Effects.Has(e.ID) {
			oosMissing = append(oosMissing, e.ID)
		}
	}
	t.Logf("SDE bundle %s: %d effects, %d data-driven, %d empty modifierInfo", ds.Version, ds.Effects.Len(), dataDriven, empty)
	t.Logf("  unused %d | SDE markers %d | LegacyEffectIDs %d | OutOfScopeEffectIDs %d | truly silent (real gaps) %d",
		unused, byMarker, byLegacy, byOos, len(silent))
	t.Logf("  legacy IDs missing from SDE: %v | out-of-scope IDs missing from SDE: %v", legacyMissing, oosMissing)

	// Drift cross-check 1: legacy ids with NON-empty modifierInfo would be double-handled unless the
	// dispatcher skips them (LegacyHandledEffectIds).
	for _, e := range engine.LegacyEffectIDs {
		eff, ok := ds.Effects.Get(e.ID)
		if !ok || len(eff.ModifierInfo) == 0 {
			continue
		}
		if !engine.LegacyHandledEffectIds.Has(e.ID) {
			t.Logf("WARN legacy id %d `%s` has %d modifierInfo entries and is NOT in LegacyHandledEffectIds — add it to the skip set or remove it from LegacyEffectIDs",
				e.ID, eff.Name, len(eff.ModifierInfo))
		}
	}
	// Drift cross-check 2: legacy ids no fittable type uses (Upwell structure modules are expected).
	structureLike := regexp.MustCompile(`(?i)structure|upwell|standup`)
	for _, e := range engine.LegacyEffectIDs {
		eff, ok := ds.Effects.Get(e.ID)
		if !ok || len(usage[e.ID]) > 0 {
			continue
		}
		if !structureLike.MatchString(eff.Name) {
			t.Logf("WARN legacy id %d `%s` is not referenced by any fittable type (potential dead handler)", e.ID, eff.Name)
		}
	}

	for i, eff := range silent {
		if i == 60 {
			t.Logf("... %d more", len(silent)-60)
			break
		}
		var sample []int
		for id := range usage[eff.ID] {
			sample = append(sample, id)
		}
		sort.Ints(sample)
		if len(sample) > 3 {
			sample = sample[:3]
		}
		t.Errorf("GAP effect %d `%s` (category %d) used by %d type(s), e.g. %v: empty modifierInfo and no LegacyEffectIDs / OutOfScopeEffectIDs entry",
			eff.ID, eff.Name, eff.EffectCategoryID, len(usage[eff.ID]), sample)
	}
}
