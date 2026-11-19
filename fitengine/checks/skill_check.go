// Package checks: porting letterale di src/skillCheck.ts e src/fitChecks.ts.
package checks

import (
	"fmt"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

type SkillRequirement struct {
	SkillID       int
	SkillName     string
	RequiredLevel float64
	CurrentLevel  float64
}

type UnmetSource struct {
	SourceTypeID int
	SourceName   string
	SourceKind   string // module | charge | drone | fighter | implant | booster | subsystem
	Requirements []SkillRequirement
}

type AggregatedSkill struct {
	SkillName     string
	RequiredLevel float64
	CurrentLevel  float64
}

type SkillCheckResult struct {
	Unmet      []UnmetSource
	Aggregated *omap.Map[int, AggregatedSkill]
}

type requirement struct {
	skillID int
	level   float64
}

func readRequirements(t *types.SdeType) []requirement {
	var out []requirement
	for _, pair := range engine.RequiredSkillPairs {
		skillID, hasS := attrOf(t, pair[0])
		level, hasL := attrOf(t, pair[1])
		if hasS && hasL && skillID != 0 && level != 0 && skillID > 0 && level > 0 {
			out = append(out, requirement{int(jsmath.Round(skillID)), jsmath.Round(level)})
		}
	}
	return out
}

func attrOf(t *types.SdeType, id int) (float64, bool) {
	for _, a := range t.Attributes {
		if a.ID == id {
			return a.V, true
		}
	}
	return 0, false
}

// CheckSkills verifica i prerequisiti di skill di tutto il fit contro un profilo.
func CheckSkills(fit *types.Fit, dataset *types.FittingDataset, profile *types.SkillProfile) SkillCheckResult {
	var unmet []UnmetSource
	aggregated := omap.New[int, AggregatedSkill]()

	skillLevel := func(id int) float64 {
		if profile != nil && profile.Skills != nil {
			if v, ok := profile.Skills.Get(id); ok {
				return float64(v)
			}
		}
		return 0
	}

	inspect := func(typeID int, kind string) {
		t, ok := dataset.GetType(typeID)
		if !ok {
			return
		}
		reqs := readRequirements(t)
		if len(reqs) == 0 {
			return
		}
		var failures []SkillRequirement
		for _, r := range reqs {
			skillName := fmt.Sprintf("Skill %d", r.skillID)
			if st, ok := dataset.GetType(r.skillID); ok {
				skillName = st.Name
			}
			current := skillLevel(r.skillID)
			agg, has := aggregated.Get(r.skillID)
			if !has || agg.RequiredLevel < r.level {
				aggregated.Set(r.skillID, AggregatedSkill{SkillName: skillName, RequiredLevel: r.level, CurrentLevel: current})
			}
			if current < r.level {
				failures = append(failures, SkillRequirement{SkillID: r.skillID, SkillName: skillName, RequiredLevel: r.level, CurrentLevel: current})
			}
		}
		if len(failures) > 0 {
			unmet = append(unmet, UnmetSource{SourceTypeID: typeID, SourceName: t.Name, SourceKind: kind, Requirements: failures})
		}
	}

	inspect(fit.ShipTypeID, "module") // la nave conta come check dello "scafo"
	for _, m := range fit.Modules {
		inspect(m.TypeID, "module")
		if m.ChargeTypeID != nil && *m.ChargeTypeID != 0 {
			inspect(*m.ChargeTypeID, "charge")
		}
	}
	for _, d := range fit.Drones {
		inspect(d.TypeID, "drone")
	}
	for _, f := range fit.Fighters {
		inspect(f.TypeID, "fighter")
	}
	for _, i := range fit.Implants {
		inspect(i.TypeID, "implant")
	}
	for _, b := range fit.Boosters {
		inspect(b.TypeID, "booster")
	}
	for _, s := range fit.Subsystems {
		inspect(s.TypeID, "subsystem")
	}
	return SkillCheckResult{Unmet: unmet, Aggregated: aggregated}
}
