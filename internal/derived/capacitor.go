// Porting letterale di src/derived/capacitor.ts: simulazione del capacitor (port di eos capSim) con la
// stessa min-heap, le stesse sift up/down e lo stesso confronto a tuple del TS (non container/heap).
package derived

import (
	"math"
	"strings"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

type CapacitorReport = types.CapacitorStats

func f64p(v float64) *float64 { return &v }

// ComputeCapacitor: capacita', recharge, stabilita' e tempo a vuoto.
func ComputeCapacitor(ctx *engine.FitContext, dataset *types.FittingDataset) CapacitorReport {
	ship := ctx.Ship
	capacity := ship.GetFinal(engine.AttrCapacitorCapacity, 0)
	rechargeMs := ship.GetFinal(engine.AttrCapacitorRechargeRate, 0)
	if capacity <= 0 || rechargeMs <= 0 {
		return CapacitorReport{Capacity: capacity, RechargeMs: rechargeMs}
	}
	peakRechargeRate := PeakRecharge(capacity, rechargeMs)
	drains, totalGrossDrain := collectDrains(ctx, dataset)
	if len(drains) == 0 {
		return CapacitorReport{Capacity: capacity, RechargeMs: rechargeMs, PeakRechargeRate: peakRechargeRate, UsagePerSecond: 0, Stable: true, StablePercent: 1}
	}
	sim := runCapSim(simInput{capacity: capacity, rechargeMs: rechargeMs, peakRechargeRate: peakRechargeRate, drains: drains, grossDrain: totalGrossDrain})
	if sim.Stable || sim.SecondsToEmpty != nil {
		return sim
	}
	return solveClosedForm(capacity, rechargeMs, peakRechargeRate, totalGrossDrain)
}

// PeakRecharge: picco di recharge (GJ/s), raggiunto al 25% di riempimento.
func PeakRecharge(capacity, rechargeMs float64) float64 {
	if capacity <= 0 || rechargeMs <= 0 {
		return 0
	}
	return (2.5 * capacity) / (rechargeMs / 1000)
}

// RechargeRateAt: velocita' di recharge a una data frazione di riempimento.
func RechargeRateAt(capacity, rechargeMs, fillFraction float64) float64 {
	if capacity <= 0 || rechargeMs <= 0 {
		return 0
	}
	x := jsmath.Max(0, jsmath.Min(1, fillFraction))
	return (10 * capacity / (rechargeMs / 1000)) * (math.Sqrt(x) - x)
}

func solveClosedForm(capacity, rechargeMs, peakRechargeRate, usagePerSecond float64) CapacitorReport {
	if usagePerSecond <= 0 {
		return CapacitorReport{Capacity: capacity, RechargeMs: rechargeMs, PeakRechargeRate: peakRechargeRate, UsagePerSecond: 0, Stable: true, StablePercent: 1}
	}
	u := (usagePerSecond * (rechargeMs / 1000)) / (10 * capacity)
	if u > 0.25 {
		return CapacitorReport{
			Capacity: capacity, RechargeMs: rechargeMs, PeakRechargeRate: peakRechargeRate, UsagePerSecond: usagePerSecond,
			Stable: false, StablePercent: 0, SecondsToEmpty: integrateTimeToEmpty(capacity, peakRechargeRate, usagePerSecond),
		}
	}
	y := (1 + math.Sqrt(1-float64(4*u))) / 2
	return CapacitorReport{Capacity: capacity, RechargeMs: rechargeMs, PeakRechargeRate: peakRechargeRate, UsagePerSecond: usagePerSecond, Stable: true, StablePercent: y * y}
}

// integrateTimeToEmpty: integrale numerico (trapezi, N=200); nil se non finito.
func integrateTimeToEmpty(capacity, peakRechargeRate, usage float64) *float64 {
	const N = 200
	integral := 0.0
	prevInv := 1 / jsmath.Max(usage, 1e-9)
	for i := 1; i <= N; i++ {
		x := float64(i) / N
		rate := 4 * peakRechargeRate * (math.Sqrt(x) - x)
		net := usage - rate
		inv := 1 / jsmath.Max(net, 1e-9)
		integral += (prevInv + inv) * 0.5 / N
		prevInv = inv
	}
	seconds := capacity * integral
	if math.IsInf(seconds, 0) || math.IsNaN(seconds) {
		return nil
	}
	return &seconds
}

type drainEntry struct {
	cycleMs, capNeed, clipSize, reloadMs float64
	isInjector, disableStagger           bool
}

const (
	capBoosterEffectID = 48
	attrCapacitorBonus = 67
	attrCapacitorNeed  = 6
	attrDurationMs     = 73
	attrReactivationMs = 669
	attrReloadTime     = 1795
	attrCapacity       = 38
	attrVolume         = 161
)

func b2s(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// drainKey e' la firma `${cycleMs}|${capNeed}|${clipSize}|${reloadMs}|${isInjector ? 1 : 0}|${disableStagger ? 1 : 0}`.
func (d *drainEntry) key() string {
	return jsmath.NumberToString(d.cycleMs) + "|" + jsmath.NumberToString(d.capNeed) + "|" + jsmath.NumberToString(d.clipSize) + "|" +
		jsmath.NumberToString(d.reloadMs) + "|" + b2s(d.isInjector) + "|" + b2s(d.disableStagger)
}

func collectDrains(ctx *engine.FitContext, dataset *types.FittingDataset) ([]drainEntry, float64) {
	var drains []drainEntry
	totalGrossDrain := 0.0
	for _, m := range ctx.Modules {
		if m.State != types.StateActive && m.State != types.StateOverload {
			continue
		}
		if m.EffectIDs.Has(capBoosterEffectID) && m.Charge != nil {
			if entry := boosterDrainEntry(m); entry != nil {
				drains = append(drains, *entry)
			}
			continue
		}
		seen := map[string]bool{} // solo lookup
		for _, eid := range m.EffectIDs.Values() {
			effect, ok := dataset.Effects.Get(eid)
			if !ok {
				continue
			}
			entry := drainEntryFromEffect(effect, m)
			if entry == nil {
				continue
			}
			sig := entry.key()
			if seen[sig] {
				continue
			}
			seen[sig] = true
			drains = append(drains, *entry)
			if entry.capNeed > 0 && entry.cycleMs > 0 {
				totalGrossDrain += entry.capNeed / (entry.cycleMs / 1000)
			}
		}
	}
	return drains, totalGrossDrain
}

func drainEntryFromEffect(effect *types.SdeEffect, mod *engine.ItemState) *drainEntry {
	if effect.DischargeAttributeID == nil {
		return nil
	}
	discharge := mod.GetFinal(*effect.DischargeAttributeID, 0)
	if discharge <= 0 {
		return nil
	}
	baseCycleMs := 1000.0 // ripiego per gli effetti di discharge senza durata
	if effect.DurationAttributeID != nil {
		baseCycleMs = mod.GetFinal(*effect.DurationAttributeID, 0)
	}
	if baseCycleMs <= 0 {
		return nil
	}
	cycleMs := jsmath.Round(baseCycleMs + mod.GetFinal(attrReactivationMs, 0))
	return &drainEntry{cycleMs: cycleMs, capNeed: discharge, clipSize: 0, reloadMs: 0, isInjector: false, disableStagger: hasTurretHardpoint(mod.Type)}
}

// effectTurretFitted: pyfa's TURRET hardpoint marker (eos Module.__calculateHardpoint).
const effectTurretFitted = 42

// hasTurretHardpoint: pyfa's `mod.hardpoint == TURRET` (disableStagger), i.e. the `turretFitted` effect,
// which every turret weapon carries but so do mining lasers and other turret-slot utilities (two Citizen
// Miners on a Deacon read 42.30 % cap stable against pyfa's 41.88 % while they were staggered).
func hasTurretHardpoint(t *types.SdeType) bool {
	for _, e := range t.Effects {
		if e.ID == effectTurretFitted {
			return true
		}
	}
	return false
}

func boosterDrainEntry(mod *engine.ItemState) *drainEntry {
	charge := mod.Charge
	if charge == nil {
		return nil
	}
	capNeed := mod.GetFinal(attrCapacitorNeed, 0)
	inject := charge.GetFinal(attrCapacitorBonus, 0)
	cycleMs := jsmath.Round(mod.GetFinal(attrDurationMs, 0))
	if cycleMs <= 0 {
		return nil
	}
	reloadMs := jsmath.Round(mod.GetFinal(attrReloadTime, 0))
	if reloadMs <= 0 {
		reloadMs = 10000 // default di Pyfa per i cap booster
	}
	capacity := mod.GetFinal(attrCapacity, 0)
	chargeVol := charge.GetFinal(attrVolume, 0)
	charges := 0.0
	if capacity > 0 && chargeVol > 0 {
		charges = math.Floor(capacity / chargeVol)
	}
	return &drainEntry{cycleMs: cycleMs, capNeed: capNeed - inject, clipSize: charges, reloadMs: reloadMs, isInjector: true, disableStagger: false}
}

type simInput struct {
	capacity, rechargeMs, peakRechargeRate float64
	drains                                 []drainEntry
	grossDrain                             float64
}

// activation e' la tupla [t_now, cycleMs, capNeed, shot, clipSize, reloadMs, isInjector].
type activation struct {
	tNow, cycleMs, capNeed, shot, clipSize, reloadMs float64
	isInjector                                       bool
}

// lcm: minimo comune multiplo (0 se uno e' 0), con l'algoritmo di Euclide su float come il TS.
func lcm(a, b float64) float64 {
	if a == 0 || b == 0 {
		return 0
	}
	product := a * b
	x, y := a, b
	for y != 0 && !math.IsNaN(y) {
		x, y = y, math.Mod(x, y)
	}
	return product / x
}

// minHeap e' la MinHeap del TS (stesse bubbleUp/bubbleDown e stesso confronto a tuple).
type minHeap struct{ a []activation }

func (h *minHeap) push(v activation) {
	h.a = append(h.a, v)
	h.bubbleUp(len(h.a) - 1)
}

func (h *minHeap) pop() (activation, bool) {
	if len(h.a) == 0 {
		return activation{}, false
	}
	top := h.a[0]
	last := h.a[len(h.a)-1]
	h.a = h.a[:len(h.a)-1]
	if len(h.a) > 0 {
		h.a[0] = last
		h.bubbleDown(0)
	}
	return top, true
}

func (h *minHeap) bubbleUp(i int) {
	a := h.a
	for i > 0 {
		p := (i - 1) >> 1
		if lt(&a[i], &a[p]) {
			a[i], a[p] = a[p], a[i]
			i = p
		} else {
			break
		}
	}
}

func (h *minHeap) bubbleDown(i int) {
	a := h.a
	n := len(a)
	for {
		l, r := 2*i+1, 2*i+2
		best := i
		if l < n && lt(&a[l], &a[best]) {
			best = l
		}
		if r < n && lt(&a[r], &a[best]) {
			best = r
		}
		if best == i {
			break
		}
		a[i], a[best] = a[best], a[i]
		i = best
	}
}

// lt: confronto lessicografico delle tuple (i booleani valgono 0/1).
func lt(x, y *activation) bool {
	xs := [7]float64{x.tNow, x.cycleMs, x.capNeed, x.shot, x.clipSize, x.reloadMs, boolNum(x.isInjector)}
	ys := [7]float64{y.tNow, y.cycleMs, y.capNeed, y.shot, y.clipSize, y.reloadMs, boolNum(y.isInjector)}
	for i := 0; i < 7; i++ {
		if xs[i] < ys[i] {
			return true
		}
		if xs[i] > ys[i] {
			return false
		}
	}
	return false
}

func boolNum(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// injector e' un AwaitingInjector: [cycleMs, capNeed, shot, clipSize, reloadMs, true].
type injector struct{ cycleMs, capNeed, shot, clipSize, reloadMs float64 }

func runCapSim(inp simInput) CapacitorReport {
	capacity, rechargeMs, peakRechargeRate, drains, grossDrain := inp.capacity, inp.rechargeMs, inp.peakRechargeRate, inp.drains, inp.grossDrain
	tau := rechargeMs / 5.0
	capCapacity := capacity

	type group struct {
		entry drainEntry
		count float64
	}
	groups := omap.New[string, *group]()
	for _, d := range drains {
		key := d.key()
		if existing, ok := groups.Get(key); ok {
			existing.count++
		} else {
			groups.Set(key, &group{entry: d, count: 1})
		}
	}

	heap := &minHeap{}
	period := 1.0
	disablePeriod := false
	const stagger = true // Pyfa: sim.stagger = True da fit.py
	for gi := 0; gi < groups.Len(); gi++ {
		_, g := groups.At(gi)
		entry, count := g.entry, g.count
		if entry.clipSize > 0 {
			disablePeriod = true
		}
		if entry.isInjector {
			for i := 0.0; i < count; i++ {
				heap.push(activation{0, entry.cycleMs, entry.capNeed, 0, entry.clipSize, entry.reloadMs, true})
			}
			continue
		}
		cycle := entry.cycleMs
		capNeed := entry.capNeed
		if stagger && !entry.disableStagger {
			if entry.clipSize == 0 {
				cycle = math.Floor(entry.cycleMs / count)
			} else {
				staggerAmount := (float64(entry.cycleMs*entry.clipSize) + entry.reloadMs) / (count * entry.clipSize)
				for i := 1.0; i < count; i++ {
					heap.push(activation{math.Floor(i * staggerAmount), entry.cycleMs, entry.capNeed, 0, entry.clipSize, entry.reloadMs, false})
				}
			}
		} else {
			capNeed *= count
		}
		period = orElse(lcm(period, cycle), period)
		heap.push(activation{0, cycle, capNeed, 0, entry.clipSize, entry.reloadMs, false})
	}

	const tMax = 6 * 60 * 60 * 1000
	if disablePeriod {
		period = tMax
	}

	var awaitingInjectors []injector
	var awaitingSnapshot []injector
	cap := capCapacity
	capLow := capCapacity
	capLowPre := capCapacity
	capWrap := capCapacity
	tWrap := period
	tLast := 0.0

	for {
		act, ok := heap.pop()
		if !ok {
			break
		}
		tNow, cycleMs, capNeed, shot, clipSize, reloadMs, isInjector := act.tNow, act.cycleMs, act.capNeed, act.shot, act.clipSize, act.reloadMs, act.isInjector
		if tNow >= tMax {
			break
		}

		// Rigenera il cap dall'evento precedente (integrazione chiusa della curva a S).
		if tNow > tLast {
			base := 1.0 + float64((math.Sqrt(cap/capCapacity)-1.0)*math.Exp((tLast-tNow)/tau))
			cap = (base * base) * capCapacity
		}

		if tNow != tLast {
			if cap < capLowPre {
				capLowPre = cap
			}
			if tNow == tWrap {
				if cap >= capWrap && awaitingMatches(awaitingInjectors, awaitingSnapshot) {
					break
				}
				capWrap = jsmath.Round(cap*10) / 10
				awaitingSnapshot = append([]injector(nil), awaitingInjectors...)
				tWrap += period
			}
		}
		tLast = tNow

		if isInjector && cap-capNeed > capCapacity {
			awaitingInjectors = append(awaitingInjectors, injector{cycleMs, capNeed, shot, clipSize, reloadMs})
		} else {
			if capNeed > cap && cap < capCapacity {
				for len(awaitingInjectors) > 0 && capNeed > cap && capCapacity > cap {
					need := jsmath.Min(capNeed-cap, capCapacity-cap)
					bestIdx := -1
					bestVal := math.Inf(-1)
					goodIdx := -1
					goodVal := math.Inf(1)
					for i := 0; i < len(awaitingInjectors); i++ {
						inj := awaitingInjectors[i]
						provided := -inj.capNeed
						if provided >= need && provided < goodVal {
							goodVal = provided
							goodIdx = i
						}
						if provided > bestVal {
							bestVal = provided
							bestIdx = i
						}
					}
					useIdx := bestIdx
					if goodIdx >= 0 {
						useIdx = goodIdx
					}
					if useIdx < 0 {
						break
					}
					inj := awaitingInjectors[useIdx]
					awaitingInjectors = append(awaitingInjectors[:useIdx], awaitingInjectors[useIdx+1:]...)
					cap -= inj.capNeed
					if cap > capCapacity {
						cap = capCapacity
					}
					injTNow := tNow + inj.cycleMs
					injShot := inj.shot + 1
					if inj.clipSize != 0 && math.Mod(injShot, inj.clipSize) == 0 {
						injShot = 0
						injTNow += inj.reloadMs
					}
					heap.push(activation{injTNow, inj.cycleMs, inj.capNeed, injShot, inj.clipSize, inj.reloadMs, true})
				}
			}

			cap -= capNeed
			if cap > capCapacity {
				cap = capCapacity
			}
			if cap < capLow {
				if cap < 0.0 {
					break // instabile
				}
				capLow = cap
			}

			for len(awaitingInjectors) > 0 && cap < capCapacity {
				need := capCapacity - cap
				bestIdx := -1
				bestVal := math.Inf(-1)
				for i := 0; i < len(awaitingInjectors); i++ {
					inj := awaitingInjectors[i]
					provided := -inj.capNeed
					if provided <= need && provided > bestVal {
						bestVal = provided
						bestIdx = i
					}
				}
				if bestIdx < 0 {
					break
				}
				inj := awaitingInjectors[bestIdx]
				awaitingInjectors = append(awaitingInjectors[:bestIdx], awaitingInjectors[bestIdx+1:]...)
				cap -= inj.capNeed
				if cap > capCapacity {
					cap = capCapacity
				}
				injTNow := tNow + inj.cycleMs
				injShot := inj.shot + 1
				if inj.clipSize != 0 && math.Mod(injShot, inj.clipSize) == 0 {
					injShot = 0
					injTNow += inj.reloadMs
				}
				heap.push(activation{injTNow, inj.cycleMs, inj.capNeed, injShot, inj.clipSize, inj.reloadMs, true})
			}

			tNow += cycleMs
			shot += 1
			if clipSize != 0 && !math.IsNaN(clipSize) && math.Mod(shot, clipSize) == 0 {
				shot = 0
				tNow += reloadMs
			}
			heap.push(activation{tNow, cycleMs, capNeed, shot, clipSize, reloadMs, isInjector})
		}
	}

	stable := cap > 0
	if !stable {
		return CapacitorReport{Capacity: capacity, RechargeMs: rechargeMs, PeakRechargeRate: peakRechargeRate, UsagePerSecond: grossDrain,
			Stable: false, StablePercent: 0, SecondsToEmpty: f64p(tLast / 1000)}
	}
	stablePercent := jsmath.Min(1, (capLow+capLowPre)/(2*capCapacity))
	return CapacitorReport{Capacity: capacity, RechargeMs: rechargeMs, PeakRechargeRate: peakRechargeRate, UsagePerSecond: grossDrain, Stable: true, StablePercent: stablePercent}
}

func injKey(a injector) string {
	return strings.Join([]string{jsmath.NumberToString(a.cycleMs), jsmath.NumberToString(a.capNeed), jsmath.NumberToString(a.shot),
		jsmath.NumberToString(a.clipSize), jsmath.NumberToString(a.reloadMs), "true"}, ",")
}

// awaitingMatches: stesso multiinsieme di injector in attesa (conteggio per chiave).
func awaitingMatches(cur, snap []injector) bool {
	if len(cur) != len(snap) {
		return false
	}
	if len(cur) == 0 {
		return true
	}
	counter := map[string]int{} // solo lookup/delete
	size := 0
	for _, a := range cur {
		key := injKey(a)
		if counter[key] == 0 {
			size++
		}
		counter[key]++
	}
	for _, a := range snap {
		key := injKey(a)
		c := counter[key]
		if c == 0 {
			return false
		}
		if c == 1 {
			delete(counter, key)
			size--
		} else {
			counter[key] = c - 1
		}
	}
	return size == 0
}
