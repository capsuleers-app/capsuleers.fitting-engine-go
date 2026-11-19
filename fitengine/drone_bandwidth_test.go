package fitengine_test

// Drone bandwidth regression (port of `npm run test:drones` of eve-fit-engine). Two bugs it pins down:
//  1. bandwidthUsed read attribute 1271 (droneBandwidth, the SHIP's cap) off the drone: always 0.
//     Drones carry 1272 (droneBandwidthUsed).
//  2. `active` counted drone STACKS in space, so five active Ogre IIs reported 1.
// Bandwidth decides how many drones actually fly, so both numbers are load-bearing.

import (
	"testing"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine/bundle"
)

const (
	dominix     = 645
	ogreII      = 2446
	hobgoblinII = 2456
)

func TestDroneBandwidth(t *testing.T) {
	ds, err := bundle.LoadDataset(bundle.FS())
	if err != nil {
		t.Fatal(err)
	}
	skills := bundle.BuildAllVSkillProfile(ds)
	attr := func(typeID, id int) (float64, bool) {
		ty, ok := ds.GetType(typeID)
		if !ok {
			return 0, false
		}
		for _, a := range ty.Attributes {
			if a.ID == id {
				return a.V, true
			}
		}
		return 0, false
	}
	// The attributes the engine depends on, asserted against the bundle itself.
	if v, ok := attr(ogreII, 1272); !ok || v != 25 {
		t.Errorf("SDE: Ogre II 1272 = %v (%v), want 25", v, ok)
	}
	if _, ok := attr(ogreII, 1271); ok {
		t.Error("SDE: Ogre II carries 1271")
	}
	if v, ok := attr(dominix, 1271); !ok || v != 125 {
		t.Errorf("SDE: Dominix 1271 = %v (%v), want 125", v, ok)
	}
	if _, ok := attr(dominix, 1272); ok {
		t.Error("SDE: Dominix carries 1272")
	}

	compute := func(drones []fitengine.FitDrone) fitengine.DerivedStats {
		fit := &fitengine.Fit{Name: "BW regression", ShipTypeID: dominix, Visibility: "PRIVATE", Tags: []string{},
			Modules: []fitengine.FitModule{}, Subsystems: []fitengine.FitSubsystem{}, Fighters: []fitengine.FitFighter{},
			Cargo: []fitengine.FitCargo{}, Implants: []fitengine.FitImplant{}, Boosters: []fitengine.FitBooster{}, Drones: drones}
		c, err := fitengine.ComputeFit(fit, ds, &fitengine.ComputeFitOptions{SkillProfile: skills})
		if err != nil {
			t.Fatal(err)
		}
		return c.Derived
	}
	check := func(label string, ok bool, val any) {
		t.Helper()
		if !ok {
			t.Errorf("%s: got %v", label, val)
		}
	}

	d := compute([]fitengine.FitDrone{{ID: "d1", TypeID: ogreII, CountTotal: 5, CountActive: 5}})
	check("5 active Ogre II -> 125 Mbit/s used", d.Drones.BandwidthUsed == 125, d.Drones.BandwidthUsed)
	check("cap reported as 125", d.Drones.BandwidthMax == 125, d.Drones.BandwidthMax)
	check("active counts DRONES not stacks", d.Drones.Active == 5, d.Drones.Active)
	check("fitting mirror agrees", d.Fitting.DroneBandwidthUsed == 125, d.Fitting.DroneBandwidthUsed)

	d = compute([]fitengine.FitDrone{
		{ID: "d1", TypeID: ogreII, CountTotal: 5, CountActive: 2},
		{ID: "d2", TypeID: hobgoblinII, CountTotal: 5, CountActive: 0},
	})
	check("idle drones burn no bandwidth", d.Drones.BandwidthUsed == 50, d.Drones.BandwidthUsed)
	check("bay counts the full stacks", d.Drones.BayUsed == 150, d.Drones.BayUsed)
	check("active = 2", d.Drones.Active == 2, d.Drones.Active)

	d = compute([]fitengine.FitDrone{
		{ID: "d1", TypeID: ogreII, CountTotal: 5, CountActive: 5},
		{ID: "d2", TypeID: hobgoblinII, CountTotal: 5, CountActive: 5},
	})
	check("5 Ogre + 5 Hobgoblin -> 150 Mbit/s", d.Drones.BandwidthUsed == 150, d.Drones.BandwidthUsed)
	check("over the 125 cap", d.Drones.BandwidthUsed > d.Drones.BandwidthMax, d.Drones.BandwidthUsed)
	check("active = 10", d.Drones.Active == 10, d.Drones.Active)

	d = compute([]fitengine.FitDrone{})
	check("no drones -> 0 used", d.Drones.BandwidthUsed == 0, d.Drones.BandwidthUsed)
	check("no drones -> 0 active", d.Drones.Active == 0, d.Drones.Active)
}
