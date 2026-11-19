// Package parity is the fixture suite and the release gate: hand-curated fits (fixtures.json: fit +
// the values Pyfa shows, screenshots in pyfa/) checked with approx() and per-fixture tolerance
// overrides. It must report 662 passed, 0 failed. eve-fit-engine runs the same fixtures and
// assertions (test/parity in that repo): a fixture added or changed here goes there too.
package parity

import (
	"encoding/json"
	"math"
	"os"
	"testing"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine/bundle"
)

type resist struct{ Em, Therm, Kin, Exp float64 }

type expected struct {
	EhpTotal, EhpShield, EhpArmor, EhpHull float64
	ShieldResist, ArmorResist, HullResist  resist
	CpuUsed, CpuMax, PowerUsed, PowerMax   float64
	CalibrationUsed, CalibrationMax        float64
	WeaponDps                              *float64
	CapCapacity                            float64
	CapStablePercent, CapSecondsToEmpty    *float64
	MaxVelocity, AlignTime                 float64
	MaxTargetingRangeKm, ScanResolution    float64
	SensorStrength, DroneRangeKm           float64
	SignatureRadius                        float64
}

type fixture struct {
	Name               string
	Screenshot         string
	Fit                fitengine.Fit
	Expected           expected
	ToleranceOverrides map[string]float64
}

// approx: ratio |diff/expected| (expected 0: 0 if |actual| < 0.01, else 1), same as eve-fit-engine.
func approx(actual, exp, tol float64) bool {
	diff := actual - exp
	var ratio float64
	if exp == 0 {
		if math.Abs(actual) < 0.01 {
			ratio = 0
		} else {
			ratio = 1
		}
	} else {
		ratio = math.Abs(diff / exp)
	}
	return ratio <= tol
}

func TestPyfaParity(t *testing.T) {
	raw, err := os.ReadFile("fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []fixture
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	ds, err := bundle.LoadDataset(bundle.FS())
	if err != nil {
		t.Fatal(err)
	}
	skills := bundle.BuildAllVSkillProfile(ds)
	uniform := &fitengine.DamageProfile{Name: "Uniform", Em: 0.25, Thermal: 0.25, Kinetic: 0.25, Explosive: 0.25}
	one, spool := 1.0, 1.0

	passed, failed := 0, 0
	for _, f := range fixtures {
		fit := f.Fit
		computed, err := fitengine.ComputeFit(&fit, ds, &fitengine.ComputeFitOptions{
			SkillProfile: skills, DamageProfile: uniform, SpoolPercent: &spool, DisintegratorSpoolPercent: &one,
		})
		if err != nil {
			t.Fatalf("%s: %v", f.Name, err)
		}
		d := computed.Derived
		exp := f.Expected
		T := func(label string, fallback float64) float64 {
			if v, ok := f.ToleranceOverrides[label]; ok {
				return v
			}
			return fallback
		}
		type check struct {
			label       string
			actual, exp float64
			tol         float64
		}
		checks := []check{
			{"Total EHP", d.Defense.EhpTotalAgainstProfile, exp.EhpTotal, T("Total EHP", 0.01)},
			{"Shield EHP", d.Defense.Shield.EhpAgainstProfile, exp.EhpShield, T("Shield EHP", 0.02)},
			{"Armor EHP", d.Defense.Armor.EhpAgainstProfile, exp.EhpArmor, T("Armor EHP", 0.01)},
			{"Hull EHP", d.Defense.Hull.EhpAgainstProfile, exp.EhpHull, T("Hull EHP", 0.02)},

			{"Shield resist EM", d.Defense.Shield.Resistances.Em * 100, exp.ShieldResist.Em, T("Shield resist EM", 0.05)},
			{"Shield resist Th", d.Defense.Shield.Resistances.Thermal * 100, exp.ShieldResist.Therm, T("Shield resist Th", 0.05)},
			{"Shield resist Kin", d.Defense.Shield.Resistances.Kinetic * 100, exp.ShieldResist.Kin, T("Shield resist Kin", 0.05)},
			{"Shield resist Exp", d.Defense.Shield.Resistances.Explosive * 100, exp.ShieldResist.Exp, T("Shield resist Exp", 0.05)},

			{"Armor resist EM", d.Defense.Armor.Resistances.Em * 100, exp.ArmorResist.Em, T("Armor resist EM", 0.05)},
			{"Armor resist Th", d.Defense.Armor.Resistances.Thermal * 100, exp.ArmorResist.Therm, T("Armor resist Th", 0.05)},
			{"Armor resist Kin", d.Defense.Armor.Resistances.Kinetic * 100, exp.ArmorResist.Kin, T("Armor resist Kin", 0.05)},
			{"Armor resist Exp", d.Defense.Armor.Resistances.Explosive * 100, exp.ArmorResist.Exp, T("Armor resist Exp", 0.05)},

			{"Hull resist EM", d.Defense.Hull.Resistances.Em * 100, exp.HullResist.Em, T("Hull resist EM", 0.05)},
			{"Hull resist Th", d.Defense.Hull.Resistances.Thermal * 100, exp.HullResist.Therm, T("Hull resist Th", 0.05)},
			{"Hull resist Kin", d.Defense.Hull.Resistances.Kinetic * 100, exp.HullResist.Kin, T("Hull resist Kin", 0.05)},
			{"Hull resist Exp", d.Defense.Hull.Resistances.Explosive * 100, exp.HullResist.Exp, T("Hull resist Exp", 0.05)},

			{"CPU used", d.Fitting.CpuUsed, exp.CpuUsed, 0.01},
			{"CPU max", d.Fitting.CpuMax, exp.CpuMax, 0.01},
			{"Power used", d.Fitting.PowerUsed, exp.PowerUsed, 0.01},
			{"Power max", d.Fitting.PowerMax, exp.PowerMax, 0.01},
			{"Calibration used", d.Fitting.CalibrationUsed, exp.CalibrationUsed, 0.01},
			{"Calibration max", d.Fitting.CalibrationMax, exp.CalibrationMax, 0.01},
		}
		if exp.WeaponDps != nil {
			checks = append(checks, check{"Weapon DPS", d.Offense.WeaponDps, *exp.WeaponDps, T("Weapon DPS", 0.01)})
		}
		checks = append(checks, check{"Cap capacity", d.Capacitor.Capacity, exp.CapCapacity, 0.01})
		if exp.CapStablePercent != nil {
			checks = append(checks, check{"Cap stable %", d.Capacitor.StablePercent * 100, *exp.CapStablePercent * 100, T("Cap stable %", 0.10)})
		}
		if exp.CapSecondsToEmpty != nil {
			secs := 0.0
			if d.Capacitor.SecondsToEmpty != nil {
				secs = *d.Capacitor.SecondsToEmpty
			}
			checks = append(checks, check{"Cap to empty (s)", secs, *exp.CapSecondsToEmpty, T("Cap to empty (s)", 0.10)})
		}
		checks = append(checks,
			check{"Max velocity", d.Navigation.MaxVelocity, exp.MaxVelocity, 0.05},
			check{"Align time", d.Navigation.AlignTimeSeconds, exp.AlignTime, 0.02},
			check{"Max targeting (km)", d.Targeting.MaxTargetingRange / 1000, exp.MaxTargetingRangeKm, T("Max targeting (km)", 0.02)},
			check{"Scan resolution", d.Targeting.ScanResolution, exp.ScanResolution, T("Scan resolution", 0.02)},
			check{"Sensor strength", d.Targeting.SensorStrength, exp.SensorStrength, T("Sensor strength", 0.02)},
			check{"Drone range (km)", d.Drones.ControlRange / 1000, exp.DroneRangeKm, T("Drone range (km)", 0.02)},
			check{"Signature radius", d.Targeting.SignatureRadius, exp.SignatureRadius, T("Signature radius", 0.02)},
		)
		for _, c := range checks {
			if approx(c.actual, c.exp, c.tol) {
				passed++
			} else {
				failed++
				t.Errorf("%s: %s got %.2f, expected %.2f", f.Name, c.label, c.actual, c.exp)
			}
		}
	}
	t.Logf("Summary: %d passed, %d failed.", passed, failed)
	if passed != 662 || failed != 0 {
		t.Fatalf("atteso 662 passed / 0 failed, ottenuto %d / %d", passed, failed)
	}
}

// BenchmarkComputeFit: informative, over the parity fixtures (All-V, uniform profile).
//
//	go test ./test/parity -run xxx -bench ComputeFit
func BenchmarkComputeFit(b *testing.B) {
	raw, err := os.ReadFile("fixtures.json")
	if err != nil {
		b.Fatal(err)
	}
	var fixtures []fixture
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		b.Fatal(err)
	}
	ds, err := bundle.LoadDataset(bundle.FS())
	if err != nil {
		b.Fatal(err)
	}
	opts := &fitengine.ComputeFitOptions{SkillProfile: bundle.BuildAllVSkillProfile(ds),
		DamageProfile: &fitengine.DamageProfile{Name: "Uniform", Em: 0.25, Thermal: 0.25, Kinetic: 0.25, Explosive: 0.25}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, f := range fixtures {
			fit := f.Fit
			if _, err := fitengine.ComputeFit(&fit, ds, opts); err != nil {
				b.Fatal(err)
			}
		}
	}
}
