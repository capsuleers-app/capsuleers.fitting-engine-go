// Package profiles: porting di src/profiles.ts (preset di danno e di bersaglio). Tabelle scritte a mano, identiche
// a quelle di eve-fit-engine: una modifica va fatta in entrambi i repo.
package profiles

import "github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"

func bp() *bool { b := true; return &b }

// DamageProfilePresets e' DAMAGE_PROFILE_PRESETS.
var DamageProfilePresets = []types.DamageProfile{
	{Name: "Uniform", Em: 0.25, Thermal: 0.25, Kinetic: 0.25, Explosive: 0.25, IsPreset: bp()},
	{Name: "EM", Em: 1, Thermal: 0, Kinetic: 0, Explosive: 0, IsPreset: bp()},
	{Name: "Thermal", Em: 0, Thermal: 1, Kinetic: 0, Explosive: 0, IsPreset: bp()},
	{Name: "Kinetic", Em: 0, Thermal: 0, Kinetic: 1, Explosive: 0, IsPreset: bp()},
	{Name: "Explosive", Em: 0, Thermal: 0, Kinetic: 0, Explosive: 1, IsPreset: bp()},
	{Name: "Sansha", Em: 0.56, Thermal: 0.44, Kinetic: 0, Explosive: 0, IsPreset: bp()},
	{Name: "Blood", Em: 0.62, Thermal: 0.38, Kinetic: 0, Explosive: 0, IsPreset: bp()},
	{Name: "Amarr Navy", Em: 0.58, Thermal: 0.42, Kinetic: 0, Explosive: 0, IsPreset: bp()},
	{Name: "Guristas", Em: 0, Thermal: 0.36, Kinetic: 0.64, Explosive: 0, IsPreset: bp()},
	{Name: "Caldari Navy", Em: 0, Thermal: 0.32, Kinetic: 0.68, Explosive: 0, IsPreset: bp()},
	{Name: "Serpentis", Em: 0, Thermal: 0.62, Kinetic: 0.38, Explosive: 0, IsPreset: bp()},
	{Name: "Gallente", Em: 0, Thermal: 0.62, Kinetic: 0.38, Explosive: 0, IsPreset: bp()},
	{Name: "Angel", Em: 0.26, Thermal: 0.13, Kinetic: 0.36, Explosive: 0.25, IsPreset: bp()},
	{Name: "Minmatar", Em: 0.13, Thermal: 0.18, Kinetic: 0.31, Explosive: 0.38, IsPreset: bp()},
	{Name: "Mordu", Em: 0, Thermal: 0.27, Kinetic: 0.73, Explosive: 0, IsPreset: bp()},
	{Name: "Drone", Em: 0.05, Thermal: 0.14, Kinetic: 0.36, Explosive: 0.45, IsPreset: bp()},
	{Name: "Sleeper", Em: 0.25, Thermal: 0.25, Kinetic: 0.25, Explosive: 0.25, IsPreset: bp()},
	{Name: "Triglavian", Em: 0, Thermal: 0.5, Kinetic: 0, Explosive: 0, IsPreset: bp()},
	{Name: "EDENCOM", Em: 0.75, Thermal: 0.25, Kinetic: 0, Explosive: 0, IsPreset: bp()},
}

// TargetProfilePresets e' TARGET_PROFILE_PRESETS.
var TargetProfilePresets = []types.TargetProfile{
	{Name: "None", SignatureRadius: 125, MaxVelocity: 0, EmResist: 0, ThermalResist: 0, KineticResist: 0, ExplosiveResist: 0, IsPreset: bp()},
	{Name: "Frigate", SignatureRadius: 38, MaxVelocity: 380, EmResist: 0.35, ThermalResist: 0.35, KineticResist: 0.35, ExplosiveResist: 0.35, IsPreset: bp()},
	{Name: "AB Frigate", SignatureRadius: 60, MaxVelocity: 950, EmResist: 0.35, ThermalResist: 0.35, KineticResist: 0.35, ExplosiveResist: 0.35, IsPreset: bp()},
	{Name: "MWD Frigate", SignatureRadius: 200, MaxVelocity: 2000, EmResist: 0.35, ThermalResist: 0.35, KineticResist: 0.35, ExplosiveResist: 0.35, IsPreset: bp()},
	{Name: "Destroyer", SignatureRadius: 60, MaxVelocity: 250, EmResist: 0.4, ThermalResist: 0.4, KineticResist: 0.4, ExplosiveResist: 0.4, IsPreset: bp()},
	{Name: "Cruiser", SignatureRadius: 130, MaxVelocity: 250, EmResist: 0.5, ThermalResist: 0.5, KineticResist: 0.5, ExplosiveResist: 0.5, IsPreset: bp()},
	{Name: "AB Cruiser", SignatureRadius: 200, MaxVelocity: 540, EmResist: 0.5, ThermalResist: 0.5, KineticResist: 0.5, ExplosiveResist: 0.5, IsPreset: bp()},
	{Name: "MWD Cruiser", SignatureRadius: 500, MaxVelocity: 1100, EmResist: 0.5, ThermalResist: 0.5, KineticResist: 0.5, ExplosiveResist: 0.5, IsPreset: bp()},
	{Name: "Battlecruiser", SignatureRadius: 285, MaxVelocity: 180, EmResist: 0.55, ThermalResist: 0.55, KineticResist: 0.55, ExplosiveResist: 0.55, IsPreset: bp()},
	{Name: "Battleship", SignatureRadius: 410, MaxVelocity: 130, EmResist: 0.6, ThermalResist: 0.6, KineticResist: 0.6, ExplosiveResist: 0.6, IsPreset: bp()},
	{Name: "AB Battleship", SignatureRadius: 500, MaxVelocity: 280, EmResist: 0.6, ThermalResist: 0.6, KineticResist: 0.6, ExplosiveResist: 0.6, IsPreset: bp()},
	{Name: "Carrier", SignatureRadius: 3500, MaxVelocity: 90, EmResist: 0.65, ThermalResist: 0.65, KineticResist: 0.65, ExplosiveResist: 0.65, IsPreset: bp()},
	{Name: "Dreadnought", SignatureRadius: 12000, MaxVelocity: 75, EmResist: 0.65, ThermalResist: 0.65, KineticResist: 0.65, ExplosiveResist: 0.65, IsPreset: bp()},
	{Name: "Titan", SignatureRadius: 22000, MaxVelocity: 60, EmResist: 0.7, ThermalResist: 0.7, KineticResist: 0.7, ExplosiveResist: 0.7, IsPreset: bp()},
	{Name: "Astrahus", SignatureRadius: 8000, MaxVelocity: 0, EmResist: 0.5, ThermalResist: 0.5, KineticResist: 0.5, ExplosiveResist: 0.5, IsPreset: bp()},
	{Name: "Fortizar", SignatureRadius: 16000, MaxVelocity: 0, EmResist: 0.6, ThermalResist: 0.6, KineticResist: 0.6, ExplosiveResist: 0.6, IsPreset: bp()},
	{Name: "Keepstar", SignatureRadius: 30000, MaxVelocity: 0, EmResist: 0.7, ThermalResist: 0.7, KineticResist: 0.7, ExplosiveResist: 0.7, IsPreset: bp()},
}
