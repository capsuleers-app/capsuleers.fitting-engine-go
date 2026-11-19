// Porting letterale di src/derived/application.ts: applicazione del danno contro un bersaglio.
package derived

import (
	"math"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

// EffectiveDps: DPS applicato dal bersaglio a una data distanza (nil target = nessun bersaglio).
func EffectiveDps(weapon *types.WeaponContribution, target *types.TargetProfile, rangeMeters float64) float64 {
	if target == nil || target.SignatureRadius <= 0 {
		return weapon.Dps
	}
	if weapon.Kind == "TURRET" || weapon.Kind == "DRONE" {
		return weapon.Dps * turretApplication(weapon, target, rangeMeters)
	}
	if weapon.Kind == "MISSILE" {
		return weapon.Dps * missileApplication(weapon, target)
	}
	if weapon.Kind == "SMARTBOMB" {
		if rangeMeters <= weapon.Range.BurstRange {
			return weapon.Dps
		}
		return 0
	}
	return weapon.Dps
}

// orElse e' `x || d` di JS su numeri.
func orElse(x, d float64) float64 {
	if x == 0 || math.IsNaN(x) {
		return d
	}
	return x
}

func turretApplication(weapon *types.WeaponContribution, target *types.TargetProfile, rng float64) float64 {
	optimal := weapon.Range.Optimal
	falloff := weapon.Range.Falloff
	tracking := orElse(weapon.Range.Tracking, 1e-9)
	const sigRes = 40000.0
	sigTarget := orElse(target.SignatureRadius, 1)
	angular := 0.0
	if rng > 0 {
		angular = target.MaxVelocity / rng
	}
	trackingFactor := (angular / tracking) * (sigRes / sigTarget)
	rangeFactor := jsmath.Max(0, (rng-optimal)/jsmath.Max(falloff, 1e-9))
	exponent := float64(trackingFactor*trackingFactor) + float64(rangeFactor*rangeFactor)
	return math.Pow(0.5, exponent)
}

func missileApplication(weapon *types.WeaponContribution, target *types.TargetProfile) float64 {
	explosionRadius := weapon.Range.ExplosionRadius
	explosionVelocity := weapon.Range.ExplosionVelocity
	drf := orElse(weapon.Range.Drf, 1)
	if explosionRadius <= 0 || explosionVelocity <= 0 {
		return 1
	}
	sigFactor := jsmath.Min(1, target.SignatureRadius/explosionRadius)
	velFactor := 1.0
	if target.MaxVelocity > 0 {
		velFactor = jsmath.Min(1, (target.SignatureRadius*explosionVelocity)/(target.MaxVelocity*explosionRadius))
	}
	return sigFactor * math.Pow(velFactor, drf)
}
