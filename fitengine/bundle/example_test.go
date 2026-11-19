package bundle_test

import (
	"fmt"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine/bundle"
)

// ExampleComputeFromEft is the README's zero-config example: EFT in, full stats out, with the embedded
// SDE bundle and All-V skills.
func ExampleComputeFromEft() {
	res, err := bundle.ComputeFromEft(`[Rifter, My Fit]
200mm AutoCannon II, Republic Fleet EMP S
200mm AutoCannon II, Republic Fleet EMP S
1MN Afterburner II
Gyrostabilizer II
`, nil)
	if err != nil {
		panic(err)
	}
	d := res.Computed.Derived
	fmt.Printf("DPS %.1f\n", d.Offense.TotalDps)
	fmt.Printf("max velocity %.0f m/s\n", d.Navigation.MaxVelocity)
	// Output:
	// DPS 110.1
	// max velocity 1193 m/s
}
