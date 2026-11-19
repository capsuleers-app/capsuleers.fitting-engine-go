// Porting letterale di src/derived/structure.ts: metadati degli Upwell structure (slot di servizio, fuel).
package derived

import (
	"fmt"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

const (
	structureCategoryID    = 65
	attrServiceSlots       = 2056
	attrServiceFuelPerHour = 2109
)

// ComputeStructureMeta: nil se la nave ospite non e' uno structure.
func ComputeStructureMeta(ctx *engine.FitContext) *types.StructureMeta {
	host, ok := ctx.Dataset.GetType(ctx.Ship.TypeID)
	if !ok {
		return nil
	}
	hostGroup, ok := ctx.Dataset.Groups.Get(host.GroupID)
	if !ok || hostGroup.CategoryID != structureCategoryID {
		return nil
	}
	serviceSlotsMax := ctx.Ship.GetFinal(attrServiceSlots, 0)
	services := []types.StructureServiceModule{}
	fuelBlocksPerHour := 0.0
	serviceSlotsUsed := 0.0
	for i := 0; i < len(ctx.Modules); i++ {
		m := ctx.Modules[i]
		if st, ok := m.SlotType(); !ok || st != types.SlotSERVICE {
			continue
		}
		serviceSlotsUsed += 1
		fuelPerHourBase := m.GetFinal(attrServiceFuelPerHour, 0)
		consuming := m.State == types.StateOnline || m.State == types.StateActive || m.State == types.StateOverload
		fuelPerHour := 0.0
		if consuming {
			fuelPerHour = fuelPerHourBase
		}
		fuelBlocksPerHour += fuelPerHour
		name := fmt.Sprintf("typeID %d", m.TypeID)
		if t, ok := ctx.Dataset.GetType(m.TypeID); ok {
			name = t.Name
		}
		services = append(services, types.StructureServiceModule{
			ModuleIndex: i, TypeID: m.TypeID, Name: name, State: m.State, FuelBlocksPerHour: fuelPerHour,
		})
	}
	return &types.StructureMeta{ServiceSlotsMax: serviceSlotsMax, ServiceSlotsUsed: serviceSlotsUsed, FuelBlocksPerHour: fuelBlocksPerHour, Services: services}
}
