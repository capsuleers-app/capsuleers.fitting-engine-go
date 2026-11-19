// Command diff is the differential Pyfa-parity harness (Go counterpart of `npm run diff` in
// eve-fit-engine).
//
// For every ship it builds a corpus of fits (generator.go, sources.go), computes each with THIS engine
// and with pyfa-org/Pyfa (oracle/pyfa_oracle.py through the .pyfa venv built by
// scripts/setup-pyfa-oracle.sh at the commit in PYFA_REF), diffs every stat (schema.go) and prints every
// difference. Exit 0 iff there are no UNEXPECTED diffs and no oracle staticdata drift; 1 otherwise;
// 2 on a harness error (as a built binary: `go run` reports every non-zero exit as 1).
//
//	go run ./tools/diff                        # all ships, generated corpus
//	go run ./tools/diff -source=workbench      # real published fits (go run ./tools/workbench first)
//	go run ./tools/diff -source=implants       # implant sets + boosters
//	go run ./tools/diff -source=all            # every corpus in one run
//	go run ./tools/diff -ships=587,29990       # specific ships
//	go run ./tools/diff -limit=20              # first N ships
//	go run ./tools/diff -group=Loki            # ship group name contains "Loki"
//	go run ./tools/diff -tol=0.02 -json        # custom tolerance / machine output
//	go run ./tools/diff -only=bonused          # only some fit types
//	go run ./tools/diff -stats=offense,defense # only some stat groups
//	go run ./tools/diff -strict                # every diff fails (ignore test/diff/known-diffs.json)
//	go run ./tools/diff -recalibrate           # after a pin bump: rewrite known-diffs.json + PYFA_REF
//	go run ./tools/diff -dump-corpus=c.json    # write the corpus (oracle specs) and stop
//
// $PYFA_DIR overrides the oracle location (default .pyfa).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine/bundle"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine/profiles"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/types"
)

type args struct {
	tol, eps    float64
	source      string
	ships       []int
	limit       int
	group       string
	only, stats []string
	json        bool
	strict      bool
	perHull     int
	recalibrate bool
	dump        string
}

func parseArgs() args {
	a := args{eps: 0.01}
	ships := flag.String("ships", "", "comma-separated ship typeIDs")
	only := flag.String("only", "", "comma-separated fit types")
	stats := flag.String("stats", "", "comma-separated stat group prefixes")
	flag.Float64Var(&a.tol, "tol", 0.01, "relative tolerance")
	flag.StringVar(&a.source, "source", "generated", "generated|workbench|implants|all (comma-separated)")
	flag.IntVar(&a.limit, "limit", 0, "first N ships")
	flag.StringVar(&a.group, "group", "", "ship group name contains")
	flag.BoolVar(&a.json, "json", false, "machine output")
	flag.BoolVar(&a.strict, "strict", false, "every diff fails (ignore known-diffs)")
	flag.IntVar(&a.perHull, "per-hull", 3, "workbench fits per hull")
	flag.StringVar(&a.dump, "dump-corpus", "", "write the oracle specs (the corpus) to this JSON file and exit")
	flag.BoolVar(&a.recalibrate, "recalibrate", false, "rewrite known-diffs.json + PYFA_REF from a strict run")
	flag.Parse()
	a.group = strings.ToLower(a.group)
	if *ships != "" {
		for _, s := range strings.Split(*ships, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(s))
			if err != nil {
				fatal("bad -ships value " + s)
			}
			a.ships = append(a.ships, n)
		}
	}
	if *only != "" {
		a.only = strings.Split(*only, ",")
	}
	if *stats != "" {
		a.stats = strings.Split(*stats, ",")
	}
	return a
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "[diff]", msg)
	os.Exit(2)
}

func logf(format string, a ...any) { fmt.Fprintf(os.Stderr, format+"\n", a...) }

func itoa(n int) string { return strconv.Itoa(n) }

func pyfaDir() string {
	if d := os.Getenv("PYFA_DIR"); d != "" {
		abs, err := filepath.Abs(d)
		if err == nil {
			return abs
		}
		return d
	}
	abs, _ := filepath.Abs(".pyfa")
	return abs
}

func chdirRoot() {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	gomod := strings.TrimSpace(string(out))
	if err != nil || gomod == "" || gomod == os.DevNull {
		fatal("not inside the Go module")
	}
	if err := os.Chdir(filepath.Dir(gomod)); err != nil {
		fatal(err.Error())
	}
}

func resolveShips(ds *types.FittingDataset, a args) []*types.SdeType {
	ships := bucket(ds, types.BucketShips)
	if a.ships != nil {
		want := setOf(a.ships...)
		ships = filter(ships, func(s *types.SdeType) bool { return want[s.ID] })
	}
	if a.group != "" {
		ships = filter(ships, func(s *types.SdeType) bool {
			g, ok := ds.Groups.Get(s.GroupID)
			name := ""
			if ok && g.Name != nil {
				name = *g.Name
			}
			return strings.Contains(strings.ToLower(name), a.group)
		})
	}
	sort.SliceStable(ships, func(i, j int) bool { return ships[i].ID < ships[j].ID })
	if a.limit > 0 && a.limit < len(ships) {
		ships = ships[:a.limit]
	}
	return ships
}

func toOurFit(spec fitSpec) *types.Fit {
	f := &types.Fit{ShipTypeID: spec.ShipTypeID, Name: spec.FitType, Visibility: "PRIVATE", Tags: []string{},
		Cargo: []types.FitCargo{}, ModeTypeID: spec.ModeTypeID,
		Modules: []types.FitModule{}, Fighters: []types.FitFighter{}, Implants: []types.FitImplant{},
		Boosters: []types.FitBooster{}, Drones: []types.FitDrone{}, Subsystems: []types.FitSubsystem{}}
	for _, m := range spec.Modules {
		f.Modules = append(f.Modules, types.FitModule{ID: m.ID, Position: float64(m.Position), SlotType: m.SlotType,
			TypeID: m.TypeID, State: m.State, ChargeTypeID: m.ChargeTypeID})
	}
	for i, x := range spec.Fighters {
		f.Fighters = append(f.Fighters, types.FitFighter{ID: "f" + itoa(i), TypeID: x.TypeID, Count: float64(x.Count), AbilityState: map[int]bool{}})
	}
	for i, im := range spec.Implants {
		slot := float64(i + 1)
		if im.Slot != nil {
			slot = *im.Slot
		}
		f.Implants = append(f.Implants, types.FitImplant{ID: "i" + itoa(i), TypeID: im.TypeID, Slot: slot})
	}
	for i, b := range spec.Boosters {
		slot := float64(i + 1)
		if b.Slot != nil {
			slot = *b.Slot
		}
		side := b.ActiveSideEffects
		if side == nil {
			side = []int{}
		}
		f.Boosters = append(f.Boosters, types.FitBooster{ID: "b" + itoa(i), TypeID: b.TypeID, Slot: slot, ActiveSideEffects: side})
	}
	for i, d := range spec.Drones {
		f.Drones = append(f.Drones, types.FitDrone{ID: "d" + itoa(i), TypeID: d.TypeID, CountTotal: float64(d.Count), CountActive: float64(d.Active)})
	}
	for i, s := range spec.Subsystems {
		f.Subsystems = append(f.Subsystems, types.FitSubsystem{ID: "s" + itoa(i), Slot: i + 1, TypeID: s})
	}
	return f
}

// Oracle wire format (oracle/pyfa_oracle.py).
type oModule struct {
	TypeID       int    `json:"typeID"`
	State        string `json:"state"`
	ChargeTypeID *int   `json:"chargeTypeID,omitempty"`
}
type oDrone struct {
	TypeID int `json:"typeID"`
	Count  int `json:"count"`
	Active int `json:"active"`
}
type oFighter struct {
	TypeID int `json:"typeID"`
	Count  int `json:"count"`
}
type oTyped struct {
	TypeID int `json:"typeID"`
}
type oBooster struct {
	TypeID      int   `json:"typeID"`
	SideEffects []int `json:"sideEffects"`
}
type oracleSpec struct {
	ID         string     `json:"id"`
	ShipTypeID int        `json:"shipTypeID"`
	Modules    []oModule  `json:"modules"`
	Drones     []oDrone   `json:"drones"`
	Subsystems []oTyped   `json:"subsystems"`
	ModeTypeID *int       `json:"modeTypeID,omitempty"`
	Fighters   []oFighter `json:"fighters"`
	Implants   []oTyped   `json:"implants"`
	Boosters   []oBooster `json:"boosters"`
}

func toOracleSpec(spec fitSpec, id string) oracleSpec {
	o := oracleSpec{ID: id, ShipTypeID: spec.ShipTypeID, ModeTypeID: spec.ModeTypeID,
		Modules: []oModule{}, Drones: []oDrone{}, Subsystems: []oTyped{}, Fighters: []oFighter{}, Implants: []oTyped{}, Boosters: []oBooster{}}
	for _, m := range spec.Modules {
		o.Modules = append(o.Modules, oModule{m.TypeID, string(m.State), m.ChargeTypeID})
	}
	for _, d := range spec.Drones {
		o.Drones = append(o.Drones, oDrone{d.TypeID, d.Count, d.Active})
	}
	for _, s := range spec.Subsystems {
		o.Subsystems = append(o.Subsystems, oTyped{s})
	}
	for _, f := range spec.Fighters {
		o.Fighters = append(o.Fighters, oFighter{f.TypeID, f.Count})
	}
	for _, im := range spec.Implants {
		o.Implants = append(o.Implants, oTyped{im.TypeID})
	}
	for _, b := range spec.Boosters {
		side := b.ActiveSideEffects
		if side == nil {
			side = []int{}
		}
		o.Boosters = append(o.Boosters, oBooster{b.TypeID, side})
	}
	return o
}

type oracleResult struct {
	ID      string         `json:"id"`
	OK      bool           `json:"ok"`
	Stats   map[string]any `json:"stats"`
	Dropped []struct {
		TypeID int    `json:"typeID"`
		Kind   string `json:"kind"`
	} `json:"dropped"`
}

func callOracle(payload any, out any) {
	pyfa := pyfaDir()
	script, _ := filepath.Abs(filepath.Join("oracle", "pyfa_oracle.py"))
	in, err := json.Marshal(payload)
	if err != nil {
		fatal(err.Error())
	}
	cmd := exec.Command(filepath.Join(pyfa, ".venv", "bin", "python"), script)
	cmd.Dir = pyfa
	cmd.Env = append(os.Environ(), "PYTHONPATH="+pyfa+":"+filepath.Join(pyfa, "_oracle_stubs"), "PYTHONDONTWRITEBYTECODE=1")
	cmd.Stdin = strings.NewReader(string(in))
	var stderr strings.Builder
	cmd.Stderr = &stderr
	stdout, err := cmd.Output()
	if err != nil {
		s := stderr.String()
		if len(s) > 2000 {
			s = s[len(s)-2000:]
		}
		fatal("oracle failed: " + err.Error() + "\n" + s)
	}
	if err := json.Unmarshal(stdout, out); err != nil {
		fatal("oracle output: " + err.Error())
	}
}

// oracleKnownTypes: which of these typeIDs the pinned pyfa staticdata can supply (pyfa's SDE trails
// CCP's, so a source can leave the unknown ones out and DECLARE it).
func oracleKnownTypes(ids []int) map[int]bool {
	var res struct {
		Known []int `json:"known"`
	}
	callOracle(map[string]any{"op": "known", "typeIDs": ids}, &res)
	return setOf(res.Known...)
}

type entry struct {
	ship *types.SdeType
	spec fitSpec
}

type item struct {
	id      string
	ship    *types.SdeType
	fitType string
	ours    *flat
	ourErr  string
}

type drift struct {
	TypeID int      `json:"typeID"`
	Kind   string   `json:"kind"`
	Name   string   `json:"name"`
	Fits   []string `json:"fits"`
}

type summary struct {
	Total         int        `json:"total"`
	OkFits        int        `json:"okFits"`
	OracleFail    int        `json:"oracleFail"`
	OurFail       int        `json:"ourFail"`
	OracleSkipped int        `json:"oracleSkipped"`
	Accepted      []statDiff `json:"accepted"`
	Drift         []drift    `json:"drift"`
	CorpusNotes   []string   `json:"corpusNotes"`
}

func main() {
	a := parseArgs()
	chdirRoot()
	if a.recalibrate {
		recalibrate(a)
		return
	}
	unexpected, sum := run(a)
	report(unexpected, sum, a)
	// Drift fails the run too: a stale pin silently shrinks coverage. Reported SEPARATELY from stat diffs.
	if len(unexpected) > 0 || len(sum.Drift) > 0 {
		os.Exit(1)
	}
}

func run(a args) ([]statDiff, summary) {
	if _, err := os.Stat(filepath.Join(pyfaDir(), "eve.db")); err != nil {
		fatal("Oracle not set up. Run: bash scripts/setup-pyfa-oracle.sh")
	}
	known, err := loadKnownDiffs()
	if err != nil {
		fatal(err.Error())
	}
	ds, err := bundle.LoadDataset(bundle.FS())
	if err != nil {
		fatal(err.Error())
	}
	allV := bundle.BuildAllVSkillProfile(ds)
	// The oracle sets a Uniform 25/25/25/25 damagePattern; pass the same so a Reactive Armor Hardener
	// adapts identically and EHP is comparable.
	var uniform *types.DamageProfile
	for i := range profiles.DamageProfilePresets {
		if profiles.DamageProfilePresets[i].Name == "Uniform" {
			uniform = &profiles.DamageProfilePresets[i]
			break
		}
	}
	ships := resolveShips(ds, a)
	sources := strings.Split(a.source, ",")
	if a.source == "all" {
		sources = []string{"generated", "workbench", "implants"}
	}
	logf("[diff] %d ships, source=%s, tol=%v", len(ships), strings.Join(sources, "+"), a.tol)

	var specs []entry
	var notes []string
	for _, source := range sources {
		switch source {
		case "generated":
			for _, ship := range ships {
				func() {
					defer func() {
						if r := recover(); r != nil {
							logf("[gen] %s: %v", ship.Name, r)
						}
					}()
					for _, spec := range generateFits(ds, ship, allV) {
						specs = append(specs, entry{ship, spec})
					}
				}()
			}
		case "workbench":
			shipIDs := map[int]bool{}
			for _, s := range ships {
				shipIDs[s.ID] = true
			}
			fits, failures, missing, err := loadWorkbenchFits(ds, allV, a.perHull)
			if err != nil {
				fatal("workbench corpus: " + err.Error())
			}
			if missing {
				notes = append(notes, "workbench corpus absent — run `go run ./tools/workbench` (needs EVEWORKBENCH_API_KEY)")
			}
			for _, spec := range fits {
				if !shipIDs[spec.ShipTypeID] {
					continue
				}
				ship, _ := ds.TypesByBucket[types.BucketShips].Get(spec.ShipTypeID)
				specs = append(specs, entry{ship, spec})
			}
			if len(failures) > 0 {
				var parts []string
				for i, f := range failures {
					if i == 3 {
						break
					}
					parts = append(parts, f.ship+": "+f.reason)
				}
				more := ""
				if len(failures) > 3 {
					more = ", …"
				}
				notes = append(notes, fmt.Sprintf("workbench: %d fit(s) unparseable — %s%s", len(failures), strings.Join(parts, "; "), more))
			}
			hulls := map[int]bool{}
			for _, f := range fits {
				hulls[f.ShipTypeID] = true
			}
			logf("[diff]   workbench: %d real fits across %d hulls", len(fits), len(hulls))
		case "implants":
			// The FULL matrix once on the first hull, then a rotating slice per hull for breadth.
			var built []entry
			if len(ships) > 0 {
				ref := ships[0]
				for _, spec := range generateImplantMatrix(ds, ref) {
					built = append(built, entry{ref, spec})
				}
				for i, ship := range ships[1:] {
					for _, spec := range generateImplantFits(ds, ship, i) {
						built = append(built, entry{ship, spec})
					}
				}
			}
			wantSet := map[int]bool{}
			var want []int
			for _, e := range built {
				for _, im := range e.spec.Implants {
					if !wantSet[im.TypeID] {
						wantSet[im.TypeID] = true
						want = append(want, im.TypeID)
					}
				}
				for _, b := range e.spec.Boosters {
					if !wantSet[b.TypeID] {
						wantSet[b.TypeID] = true
						want = append(want, b.TypeID)
					}
				}
			}
			knownTypes := oracleKnownTypes(want)
			absent := map[int]bool{}
			kept := 0
			for _, e := range built {
				ok := true
				for _, im := range e.spec.Implants {
					if !knownTypes[im.TypeID] {
						absent[im.TypeID], ok = true, false
					}
				}
				for _, b := range e.spec.Boosters {
					if !knownTypes[b.TypeID] {
						absent[b.TypeID], ok = true, false
					}
				}
				if ok {
					specs = append(specs, e)
					kept++
				}
			}
			if len(absent) > 0 {
				notes = append(notes, fmt.Sprintf("implants: %d item(s) absent from the pinned pyfa staticdata — excluded (pyfa's SDE trails CCP's; not an engine difference)", len(absent)))
			}
			refName := "—"
			if len(ships) > 0 {
				refName = ships[0].Name
			}
			rest := 0
			if len(ships) > 1 {
				rest = len(ships) - 1
			}
			logf("[diff]   implants: full matrix on %s + rotating slice on %d hulls → %d comparable fits", refName, rest, kept)
		default:
			fatal(fmt.Sprintf("unknown source %q", source))
		}
	}

	var items []item
	var oracleSpecs []oracleSpec
	for _, e := range specs {
		if a.only != nil && !contains(a.only, e.spec.FitType) {
			continue
		}
		id := itoa(e.ship.ID) + ":" + e.spec.FitType
		it := item{id: id, ship: e.ship, fitType: e.spec.FitType}
		func() {
			defer func() {
				if r := recover(); r != nil {
					it.ourErr = fmt.Sprint(r)
				}
			}()
			c, err := fitengine.ComputeFit(toOurFit(e.spec), ds, &fitengine.ComputeFitOptions{SkillProfile: allV, DamageProfile: uniform})
			if err != nil {
				it.ourErr = err.Error()
				return
			}
			it.ours = oursToSchema(&c.Derived)
		}()
		items = append(items, it)
		oracleSpecs = append(oracleSpecs, toOracleSpec(e.spec, id))
	}

	if a.dump != "" {
		out, err := json.MarshalIndent(oracleSpecs, "", " ")
		if err == nil {
			err = os.WriteFile(a.dump, out, 0o644)
		}
		if err != nil {
			fatal(err.Error())
		}
		logf("[diff] wrote %d specs to %s", len(oracleSpecs), a.dump)
		os.Exit(0)
	}
	logf("[diff] running oracle on %d fits...", len(oracleSpecs))
	var results []oracleResult
	callOracle(oracleSpecs, &results)
	oracle := map[string]*oracleResult{}
	for i := range results {
		oracle[results[i].ID] = &results[i]
	}

	var all []statDiff
	driftBy := map[int]*drift{}
	var driftOrder []int
	sum := summary{Total: len(items), CorpusNotes: notes}
	for _, it := range items {
		o := oracle[it.id]
		if it.ourErr != "" {
			sum.OurFail++
			continue
		}
		if o == nil || !o.OK {
			sum.OracleFail++
			continue
		}
		// The pinned pyfa didn't know some type this fit uses: the two sides computed different fits.
		// Record it as pin drift and skip (diffing it would manufacture phantom diffs).
		if len(o.Dropped) > 0 {
			sum.OracleSkipped++
			for _, d := range o.Dropped {
				e := driftBy[d.TypeID]
				if e == nil {
					name := "?"
					if t, ok := ds.GetType(d.TypeID); ok && t.Name != "" {
						name = t.Name
					}
					e = &drift{TypeID: d.TypeID, Kind: d.Kind, Name: name}
					driftBy[d.TypeID] = e
					driftOrder = append(driftOrder, d.TypeID)
				}
				e.Fits = append(e.Fits, it.ship.Name+" / "+it.fitType)
			}
			continue
		}
		ours := it.ours
		pyfa := flattenAny(o.Stats, ours)
		if a.stats != nil {
			keep := func(k string) bool {
				for _, g := range a.stats {
					if strings.HasPrefix(k, g) {
						return true
					}
				}
				return false
			}
			// Filter BOTH sides (filtering only pyfa's turned every ours-only key into a diff).
			ours, pyfa = ours.filter(keep), pyfa.filter(keep)
		}
		diffs := diffStats(ours, pyfa, a.tol, a.eps)
		if len(diffs) == 0 {
			sum.OkFits++
		}
		for _, d := range diffs {
			d.Ship, d.ShipID, d.FitType = it.ship.Name, it.ship.ID, it.fitType
			all = append(all, d)
		}
	}
	var unexpected []statDiff
	for _, d := range all {
		if a.strict || !known.has(d) {
			unexpected = append(unexpected, d)
		} else {
			sum.Accepted = append(sum.Accepted, d)
		}
	}
	for _, id := range driftOrder {
		sum.Drift = append(sum.Drift, *driftBy[id])
	}
	sort.SliceStable(sum.Drift, func(i, j int) bool { return len(sum.Drift[i].Fits) > len(sum.Drift[j].Fits) })
	return unexpected, sum
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func fmtVal(v any) string {
	switch x := v.(type) {
	case nil:
		return "—"
	case float64:
		// jsmath.ToFixed, not FormatFloat: the report prints the same digits as the TS harness.
		if x >= 100 || x <= -100 {
			return jsmath.ToFixed(x, 1)
		}
		return jsmath.ToFixed(x, 3)
	default:
		return fmt.Sprint(x)
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

func report(diffs []statDiff, sum summary, a args) {
	if a.json {
		out, _ := json.MarshalIndent(map[string]any{"summary": sum, "diffs": diffs}, "", "  ")
		fmt.Println(string(out))
		return
	}
	// Coverage caveats first and never silently: a shrinking corpus that still says "no differences" is
	// the one failure this harness must not produce.
	for _, n := range sum.CorpusNotes {
		fmt.Printf("\nWARNING: %s\n", n)
	}
	if len(sum.Drift) > 0 {
		fmt.Printf("\n=== ORACLE SDE DRIFT (%d type%s absent from the pinned pyfa) ===\n", len(sum.Drift), plural(len(sum.Drift), "", "s"))
		fmt.Printf("  The pinned oracle cannot fit these, so %d fit(s) were SKIPPED as not comparable.\n", sum.OracleSkipped)
		fmt.Println("  Fix: move PYFA_REF to a pyfa whose staticdata carries them (the same pin as eve-fit-engine),")
		fmt.Println("  re-run `bash scripts/setup-pyfa-oracle.sh`, then `go run ./tools/diff -recalibrate`.")
		for _, d := range sum.Drift {
			fits := d.Fits
			more := ""
			if len(fits) > 4 {
				fits, more = fits[:4], ", …"
			}
			fmt.Printf("    %-7d %-30s %-10s %d fit%s: %s%s\n", d.TypeID, d.Name, d.Kind, len(d.Fits), plural(len(d.Fits), "", "s"), strings.Join(fits, ", "), more)
		}
	}
	if len(sum.Accepted) > 0 {
		fmt.Printf("\n=== ACCEPTED known differences (%d; documented in test/diff/known-diffs.json) ===\n", len(sum.Accepted))
		known, _ := loadKnownDiffs()
		var reasons []string
		byReason := map[string][]statDiff{}
		for _, d := range sum.Accepted {
			r := known.reason(d)
			if r == "" {
				r = "documented"
			}
			if _, ok := byReason[r]; !ok {
				reasons = append(reasons, r)
			}
			byReason[r] = append(byReason[r], d)
		}
		for _, r := range reasons {
			fmt.Printf("  • %s\n", r)
			for _, d := range byReason[r] {
				fmt.Printf("      %-22s %-12s %s  ours=%s pyfa=%s\n", d.Ship, d.FitType, d.Key, fmtVal(d.Ours), fmtVal(d.Pyfa))
			}
		}
	}
	var keys []string
	byStat := map[string][]statDiff{}
	for _, d := range diffs {
		if _, ok := byStat[d.Key]; !ok {
			keys = append(keys, d.Key)
		}
		byStat[d.Key] = append(byStat[d.Key], d)
	}
	sort.SliceStable(keys, func(i, j int) bool { return len(byStat[keys[i]]) > len(byStat[keys[j]]) })
	if len(diffs) > 0 {
		fmt.Println("\n=== UNEXPECTED DIFFERENCES (ours vs pyfa) ===")
		for _, k := range keys {
			list := byStat[k]
			fmt.Printf("\n● %s  (%d fit%s)\n", k, len(list), plural(len(list), "", "s"))
			sort.SliceStable(list, func(i, j int) bool { return absPct(list[i]) > absPct(list[j]) })
			for i, d := range list {
				if i == 12 {
					fmt.Printf("    … +%d more\n", len(list)-12)
					break
				}
				pct := ""
				if d.PctDelta != nil {
					sign := ""
					if *d.PctDelta > 0 {
						sign = "+"
					}
					pct = fmt.Sprintf("  (%s%.1f%%)", sign, *d.PctDelta)
				}
				fmt.Printf("    %-22s %-12s ours=%12s pyfa=%12s%s\n", d.Ship, d.FitType, fmtVal(d.Ours), fmtVal(d.Pyfa), pct)
			}
		}
	}
	fmt.Printf("\nSummary: %d/%d fits match | %d unexpected diff%s across %d stats | %d accepted | our-fail %d | oracle-fail %d | oracle-skipped %d\n",
		sum.OkFits, sum.Total, len(diffs), plural(len(diffs), "", "s"), len(byStat), len(sum.Accepted), sum.OurFail, sum.OracleFail, sum.OracleSkipped)
	if len(sum.Drift) > 0 {
		fmt.Printf("WARNING: oracle SDE drift — the pin is stale (see above). Engine parity was NOT assessed on %d fit(s).\n", sum.OracleSkipped)
	}
	if len(diffs) == 0 && len(sum.Drift) == 0 {
		acc := ""
		switch {
		case sum.Total == 0:
			fmt.Println("WARNING: nothing was compared — 0 fits in the corpus. This is NOT a pass.")
		case len(sum.CorpusNotes) > 0:
			if len(sum.Accepted) > 0 {
				acc = fmt.Sprintf(" (%d documented quirk%s accepted)", len(sum.Accepted), plural(len(sum.Accepted), "", "s"))
			}
			fmt.Printf("OK: no unexpected differences across the %d fits compared%s — but the corpus is INCOMPLETE, see the note%s above.\n",
				sum.Total, acc, plural(len(sum.CorpusNotes), "", "s"))
		default:
			if len(sum.Accepted) > 0 {
				acc = fmt.Sprintf(" (%d documented quirks accepted; run -strict to list them as failures)", len(sum.Accepted))
			}
			fmt.Printf("OK: no unexpected differences — engine matches pyfa across all %d sampled fits%s.\n", sum.Total, acc)
		}
	}
}

func absPct(d statDiff) float64 {
	if d.PctDelta == nil {
		return 0
	}
	if *d.PctDelta < 0 {
		return -*d.PctDelta
	}
	return *d.PctDelta
}
