// Command stacking generates the stacking-penalty registry from pyfa's own effect handlers.
//
// In EVE the stacking penalty is not a property of an ATTRIBUTE: it is a property of the (effect,
// attribute) pair, and pyfa writes it one call at a time. `boostItemAttr('damageMultiplier', ...)`
// multiplies in full, while `boostItemAttr('mass', ..., stackingPenalties=True)` joins a penalised
// chain. Getting the granularity wrong in either direction is visible: penalising every modifier that
// shares an attribute chains bonuses pyfa keeps apart (a Rokh with an MWD and an MJD read 6 335 m of
// signature radius against 6 875), penalising a whole effect because one of its calls does breaks the
// Siege Module (unpenalised turret damage, penalised mass/repair bonuses).
//
// Encoding, per effect id (attributes are ATTRIBUTE IDS, resolved from the raw SDE table: the shipped
// bundle carries only published attributes, and some penalised ones like `warpSpeedMultiplier` are not):
//   - "attrId#group"            every call touching that attribute penalises, in that penalty group;
//   - "receiver:attrId#group"   only calls with that receiver penalise (Siege/Bastion shape);
//   - "*"                       the attribute name is computed at runtime ('%sDamage' % type);
//   - "...!ctx+ctx"             the pyfa contexts exempt from the penalty.
//
// One run, two outputs, both rendered from the same canonical model (never edit them by hand):
//
//	generated/stacking-penalised.json        canonical model
//	internal/gen/stacking_penalised_gen.go   Go registry + StackPenaltyGroup
//
// Usage:
//
//	go run ./tools/stacking                      # analyse pyfa at .pyfa + raw SDE, rewrite both outputs
//	go run ./tools/stacking -check               # analyse, fail (non-zero) if either output would change
//	go run ./tools/stacking -from-json           # no pyfa: re-render the Go file from the committed JSON
//	go run ./tools/stacking -check -from-json    # no pyfa: verify the Go file is what the JSON renders
//	go run ./tools/stacking -fetch-sde           # (re)stage the two raw SDE tables into .sde-src/
//
// The analysis reads the pyfa checkout the diff oracle pins (.pyfa, built by scripts/setup-pyfa-oracle.sh
// at the commit in PYFA_REF), so the registry and the oracle can never describe two pyfa versions.
// It is a line-for-line port of scripts/build-stacking-registry.mjs of eve-fit-engine: both repos must
// produce the SAME generated/stacking-penalised.json at the same pin.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	effectsPy = filepath.Join(pyfaDir(), "eos", "effects.py")
	sdeDir    = ".sde-src"
	outGo     = filepath.Join("internal", "gen", "stacking_penalised_gen.go")
	outJSON   = filepath.Join("generated", "stacking-penalised.json")
)

// overrides: corrections where pyfa's handler BODY doesn't describe the SDE effect of the same id.
// pyfa's Effect5998 (freighterSMACapacityBonusO1) boosts AGILITY and penalises it, while the SDE's 5998
// moves the maintenance-bay capacity and it is 6001 that moves agility: the bodies are transposed.
// Without it a Bowhead read 18.21 s of align time against pyfa's 18.81. [sdeEffectId, attributeId, group].
var overrides = []struct {
	effect, attr int
	group        string
}{
	{6001, 70, "default"}, // freighterAgilityBonus2O2 <- pyfa's Effect5998
}

// pyfaDir is the pyfa checkout: $PYFA_DIR, else .pyfa (scripts/setup-pyfa-oracle.sh).
func pyfaDir() string {
	if d := os.Getenv("PYFA_DIR"); d != "" {
		return d
	}
	return ".pyfa"
}

const sdeIndexURL = "https://sde.evetech.net/tranquility/latest/"

var sdeTables = []string{"dogmaAttributes", "dogmaEffects"}

func main() {
	check := flag.Bool("check", false, "fail (exit 2) if an output is stale")
	fromJSON := flag.Bool("from-json", false, "render from the committed JSON instead of analysing pyfa")
	fetch := flag.Bool("fetch-sde", false, "(re)download the raw SDE tables into .sde-src/ and exit")
	flag.Parse()
	if err := chdirRoot(); err != nil {
		fail(1, err.Error())
	}
	if *fetch {
		if err := fetchSDE(); err != nil {
			fail(1, err.Error())
		}
		return
	}
	var m model
	if *fromJSON {
		raw, err := os.ReadFile(outJSON)
		if err != nil {
			fail(1, err.Error())
		}
		if m, err = parseModel(raw); err != nil {
			fail(1, err.Error())
		}
		if renderJSON(m) != string(raw) {
			fail(2, outJSON+" is not in canonical form")
		}
	} else {
		var err error
		if m, err = analyse(); err != nil {
			fail(1, err.Error())
		}
	}
	outputs := [][2]string{{outJSON, renderJSON(m)}, {outGo, renderGo(m)}}
	if *check {
		var stale []string
		for _, o := range outputs {
			cur, _ := os.ReadFile(o[0])
			if string(cur) != o[1] {
				stale = append(stale, o[0])
			}
		}
		if len(stale) > 0 {
			fail(2, fmt.Sprintf("stale: %s — run `go run ./tools/stacking`", strings.Join(stale, ", ")))
		}
		how := ""
		if *fromJSON {
			how = ", rendered from the JSON"
		}
		fmt.Printf("[stacking] up to date (%d effects, %d penalised calls; json + go%s)\n", len(m.effects), m.penalisedCalls, how)
		return
	}
	for _, o := range outputs {
		if *fromJSON && o[0] == outJSON {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(o[0]), 0o755); err != nil {
			fail(1, err.Error())
		}
		if err := os.WriteFile(o[0], []byte(o[1]), 0o644); err != nil {
			fail(1, err.Error())
		}
	}
	fmt.Printf("[stacking] wrote json + go — %d/%d calls penalise across %d effects (%d receiver-qualified)\n",
		m.penalisedCalls, m.calls, len(m.effects), m.ambiguous)
}

func fail(code int, msg string) {
	fmt.Fprintln(os.Stderr, "[stacking]", msg)
	os.Exit(code)
}

// chdirRoot moves to the module root so the relative paths above hold from any subdirectory.
func chdirRoot() error {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return err
	}
	gomod := strings.TrimSpace(string(out))
	if gomod == "" || gomod == os.DevNull {
		return errors.New("not inside the Go module")
	}
	return os.Chdir(filepath.Dir(gomod))
}

type model struct {
	calls, penalisedCalls, ambiguous int
	effects                          []effectKeys
}

type effectKeys struct {
	id   int
	keys []string
}

type use struct {
	recv, group string
	pen         bool
	exempt      []string
}

var (
	classRe      = regexp.MustCompile(`\nclass (Effect\d+)\(BaseEffect\):`)
	callRe       = regexp.MustCompile(`(\w+)\.(boostItemAttr|multiplyItemAttr|filteredItemBoost|filteredChargeBoost|filteredItemMultiply|filteredChargeMultiply)\(`)
	penVarRe     = regexp.MustCompile(`stackingPenalties=(\w+)`)
	groupRe      = regexp.MustCompile(`penaltyGroup='([A-Za-z]+)'`)
	litRe        = regexp.MustCompile(`^'([A-Za-z0-9_]+)'$`)
	ifLineRe     = regexp.MustCompile(`^\s*if\s.*:\s*$`)
	contextRe    = regexp.MustCompile(`'(\w+)'\s+(?:not\s+)?in\s+context`)
	trueBoundary = "True"
)

// penVariable is JS /stackingPenalties=(?!True\b|False\b)(\w+)/: the first assignment that is neither
// literal True nor False. A word equal to "True" is exactly "True\b" because \w+ is greedy.
func penVariable(s string) (string, bool) {
	for _, m := range penVarRe.FindAllStringSubmatch(s, -1) {
		if m[1] != "True" && m[1] != "False" {
			return m[1], true
		}
	}
	return "", false
}

func penTrue(s string) bool {
	for _, m := range penVarRe.FindAllStringSubmatch(s, -1) {
		if m[1] == trueBoundary {
			return true
		}
	}
	return false
}

// splitArgs splits an argument list on top-level commas (lambdas carry their own).
func splitArgs(text string) []string {
	var args []string
	depth := 0
	var cur strings.Builder
	var quote byte
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if quote != 0 {
			cur.WriteByte(ch)
			if ch == quote && (i == 0 || text[i-1] != '\\') {
				quote = 0
			}
			continue
		}
		switch {
		case ch == '\'' || ch == '"':
			quote = ch
			cur.WriteByte(ch)
		case strings.IndexByte("([{", ch) >= 0:
			depth++
			cur.WriteByte(ch)
		case strings.IndexByte(")]}", ch) >= 0:
			depth--
			cur.WriteByte(ch)
		case ch == ',' && depth == 0:
			args = append(args, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteByte(ch)
		}
	}
	if strings.TrimSpace(cur.String()) != "" {
		args = append(args, strings.TrimSpace(cur.String()))
	}
	return args
}

// callBody is the text inside the parens of the call whose '(' is at open.
func callBody(text string, open int) string {
	depth := 0
	var quote byte
	for i := open; i < len(text); i++ {
		ch := text[i]
		if quote != 0 {
			if ch == quote && text[i-1] != '\\' {
				quote = 0
			}
			continue
		}
		switch ch {
		case '\'', '"':
			quote = ch
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return text[open+1 : i]
			}
		}
	}
	return ""
}

// receiverOf maps pyfa's receiver expression + call name onto the engine's target kinds.
func receiverOf(receiver, fn string) string {
	switch receiver {
	case "ship":
		return "ship"
	case "drones":
		return "drone"
	case "fighters":
		return "fighter"
	case "character":
		return "character"
	case "modules":
		if strings.HasPrefix(fn, "filteredCharge") {
			return "charge"
		}
		return "module"
	}
	return "self"
}

// exemptContexts: the source kinds pyfa EXEMPTS from an effect's penalty (`penalized = False if 'skill'
// in context ...`, `penalties = 'ship' not in context`, or the if/else spelling).
func exemptContexts(body string) []string {
	varName, ok := penVariable(body)
	if !ok {
		return nil
	}
	assign := regexp.MustCompile(`\b` + regexp.QuoteMeta(varName) + `\s*=`)
	all := strings.Split(body, "\n")
	names := map[string]bool{}
	for i, l := range all {
		keep := assign.MatchString(l)
		if !keep && ifLineRe.MatchString(l) {
			for j := i + 1; j < i+3 && j < len(all); j++ {
				if assign.MatchString(all[j]) {
					keep = true
					break
				}
			}
		}
		if !keep {
			continue
		}
		for _, c := range contextRe.FindAllStringSubmatch(l, -1) {
			names[c[1]] = true
		}
	}
	out := make([]string, 0, len(names))
	for n := range names {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

type attrUses struct {
	attr string
	uses []use
}

type effectUses struct {
	id    int
	attrs []*attrUses
}

func classify(src string) (perEffect []*effectUses, calls, penalisedCalls int) {
	idx := map[int]*effectUses{}
	ms := classRe.FindAllStringSubmatchIndex(src, -1)
	for k, m := range ms {
		id, _ := strconv.Atoi(src[m[2]+len("Effect") : m[3]])
		end := len(src)
		if k+1 < len(ms) {
			end = ms[k+1][0]
		}
		body := src[m[1]:end]
		exempt := exemptContexts(body)
		for _, c := range callRe.FindAllStringSubmatchIndex(body, -1) {
			receiver, fn := body[c[2]:c[3]], body[c[4]:c[5]]
			inner := callBody(body, c[1]-1)
			if inner == "" {
				continue
			}
			calls++
			args := splitArgs(inner)
			_, penVar := penVariable(inner)
			pen := penTrue(inner) || penVar
			if pen {
				penalisedCalls++
			}
			group := "default"
			if g := groupRe.FindStringSubmatch(inner); g != nil {
				group = g[1]
			}
			ai := 0
			if strings.HasPrefix(fn, "filtered") {
				ai = 1
			}
			attrArg := ""
			if ai < len(args) {
				attrArg = args[ai]
			}
			attr := "*"
			if lit := litRe.FindStringSubmatch(attrArg); lit != nil {
				attr = lit[1]
			}
			e := idx[id]
			if e == nil {
				e = &effectUses{id: id}
				idx[id] = e
				perEffect = append(perEffect, e)
			}
			var au *attrUses
			for _, a := range e.attrs {
				if a.attr == attr {
					au = a
					break
				}
			}
			if au == nil {
				au = &attrUses{attr: attr}
				e.attrs = append(e.attrs, au)
			}
			au.uses = append(au.uses, use{recv: receiverOf(receiver, fn), pen: pen, group: group, exempt: exempt})
		}
	}
	return perEffect, calls, penalisedCalls
}

type registry struct {
	order []int
	keys  map[int][]string
}

func (r *registry) set(id int, keys []string) {
	if _, ok := r.keys[id]; !ok {
		r.order = append(r.order, id)
	}
	r.keys[id] = keys
}

func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func toRegistry(perEffect []*effectUses, attrIDs map[string]int) (reg *registry, ambiguous int, unresolved []string) {
	reg = &registry{keys: map[int][]string{}}
	seenUnresolved := map[string]bool{}
	for _, e := range perEffect {
		keys := map[string]bool{}
		for _, au := range e.attrs {
			var pen []use
			for _, u := range au.uses {
				if u.pen {
					pen = append(pen, u)
				}
			}
			if len(pen) == 0 {
				continue
			}
			key := au.attr
			if au.attr != "*" {
				aid, ok := attrIDs[au.attr]
				if !ok {
					if !seenUnresolved[au.attr] {
						seenUnresolved[au.attr] = true
						unresolved = append(unresolved, au.attr)
					}
					continue
				}
				key = strconv.Itoa(aid)
			}
			suffix := func(u use) string {
				if len(u.exempt) == 0 {
					return ""
				}
				return "!" + strings.Join(u.exempt, "+")
			}
			groups := map[string]bool{}
			for _, u := range pen {
				groups[u.group] = true
			}
			if len(pen) == len(au.uses) && len(groups) == 1 {
				keys[key+"#"+pen[0].group+suffix(pen[0])] = true
			} else {
				ambiguous++
				for _, u := range pen {
					keys[u.recv+":"+key+"#"+u.group+suffix(u)] = true
				}
			}
		}
		if len(keys) > 0 {
			reg.set(e.id, sortedKeys(keys))
		}
	}
	return reg, ambiguous, unresolved
}

// readJSONL calls fn for every non-empty line of an SDE jsonl table.
func readJSONL(path string, fn func(raw []byte) error) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 64<<20)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		if err := fn(line); err != nil {
			return err
		}
	}
	return sc.Err()
}

// sdeKey reads the row key: `_key` (legacy zip shape, as staged by eve-fit-engine) or the API's own
// `<table>ID` field.
func sdeKey(row map[string]json.RawMessage, apiField string) (int, bool) {
	for _, f := range []string{"_key", apiField} {
		if v, ok := row[f]; ok {
			var n int
			if json.Unmarshal(v, &n) == nil {
				return n, true
			}
		}
	}
	return 0, false
}

// attributeIDs maps pyfa's attribute NAMES onto ids from the FULL SDE table (unpublished included).
func attributeIDs() (map[string]int, error) {
	byName := map[string]int{}
	err := readJSONL(filepath.Join(sdeDir, "dogmaAttributes.jsonl"), func(raw []byte) error {
		var row map[string]json.RawMessage
		if err := json.Unmarshal(raw, &row); err != nil {
			return err
		}
		var name string
		_ = json.Unmarshal(row["name"], &name)
		if name == "" {
			return nil
		}
		if id, ok := sdeKey(row, "attributeID"); ok {
			byName[name] = id
		}
		return nil
	})
	return byName, err
}

// sdeEffectAttributes: SDE effect id -> the attribute ids its modifierInfo actually touches.
func sdeEffectAttributes() (map[int]map[int]bool, error) {
	byEffect := map[int]map[int]bool{}
	err := readJSONL(filepath.Join(sdeDir, "dogmaEffects.jsonl"), func(raw []byte) error {
		var row map[string]json.RawMessage
		if err := json.Unmarshal(raw, &row); err != nil {
			return err
		}
		id, ok := sdeKey(row, "effectID")
		if !ok {
			return nil
		}
		var mods []struct {
			ModifiedAttributeID *int `json:"modifiedAttributeID"`
		}
		_ = json.Unmarshal(row["modifierInfo"], &mods)
		ids := map[int]bool{}
		for _, m := range mods {
			if m.ModifiedAttributeID != nil {
				ids[*m.ModifiedAttributeID] = true
			}
		}
		byEffect[id] = ids
		return nil
	})
	return byEffect, err
}

func analyse() (model, error) {
	for _, req := range []struct{ what, file, how string }{
		{"the pinned pyfa checkout", effectsPy, "PYFA_REF=$(cat PYFA_REF) bash scripts/setup-pyfa-oracle.sh"},
		{"the raw SDE", filepath.Join(sdeDir, "dogmaAttributes.jsonl"), "go run ./tools/stacking -fetch-sde"},
		{"the raw SDE", filepath.Join(sdeDir, "dogmaEffects.jsonl"), "go run ./tools/stacking -fetch-sde"},
	} {
		if _, err := os.Stat(req.file); err != nil {
			fmt.Fprintf(os.Stderr, "[stacking] %s is missing (%s) — run `%s` first.\n", req.what, req.file, req.how)
			fmt.Fprintln(os.Stderr, "[stacking]   the registry maps pyfa attribute NAMES onto ids, and the shipped")
			fmt.Fprintln(os.Stderr, "[stacking]   bundle carries only PUBLISHED attributes, so the raw SDE is required.")
			os.Exit(1)
		}
	}
	src, err := os.ReadFile(effectsPy)
	if err != nil {
		return model{}, err
	}
	attrIDs, err := attributeIDs()
	if err != nil {
		return model{}, err
	}
	perEffect, calls, penalisedCalls := classify(string(src))
	reg, ambiguous, unresolved := toRegistry(perEffect, attrIDs)
	// Cross-check every key against the SDE: a penalty pinned to an attribute the effect doesn't modify
	// means pyfa's handler and the SDE effect have drifted apart. Reported, never silent.
	sdeAttrs, err := sdeEffectAttributes()
	if err != nil {
		return model{}, err
	}
	var mismatched []string
	for _, eid := range reg.order {
		known := sdeAttrs[eid]
		if len(known) == 0 {
			continue
		}
		for _, key := range reg.keys[eid] {
			head := strings.SplitN(key, "!", 2)[0]
			head = strings.SplitN(head, "#", 2)[0]
			parts := strings.Split(head, ":")
			attr := parts[len(parts)-1]
			if attr == "*" {
				continue
			}
			n, err := strconv.Atoi(attr)
			if err != nil || !known[n] {
				mismatched = append(mismatched, fmt.Sprintf("%d→attr %s", eid, attr))
			}
		}
	}
	if len(mismatched) > 0 {
		fmt.Fprintf(os.Stderr, "[stacking] %d key(s) pin a penalty to an attribute the SDE effect does not modify (pyfa handler drift): %s\n",
			len(mismatched), strings.Join(mismatched, ", "))
		fmt.Fprintln(os.Stderr, "[stacking]   → if one of them matters, add an overrides row mapping it onto the SDE effect that does.")
	}
	for _, o := range overrides {
		keys := map[string]bool{}
		for _, k := range reg.keys[o.effect] {
			keys[k] = true
		}
		keys[fmt.Sprintf("%d#%s", o.attr, o.group)] = true
		reg.set(o.effect, sortedKeys(keys))
	}
	if len(unresolved) > 0 {
		show := unresolved
		if len(show) > 8 {
			show = show[:8]
		}
		fmt.Fprintf(os.Stderr, "[stacking] %d attribute name(s) absent from the SDE, skipped: %s\n", len(unresolved), strings.Join(show, ", "))
	}
	ids := append([]int(nil), reg.order...)
	sort.Ints(ids)
	m := model{calls: calls, penalisedCalls: penalisedCalls, ambiguous: ambiguous}
	for _, id := range ids {
		m.effects = append(m.effects, effectKeys{id, reg.keys[id]})
	}
	return m, nil
}

func quoteKeys(keys []string, sep string) string {
	q := make([]string, len(keys))
	for i, k := range keys {
		q[i] = strconv.Quote(k)
	}
	return strings.Join(q, sep)
}

// renderJSON: canonical JSON, one effect per line, stable and diff-friendly (byte-identical to the TS
// generator's output).
func renderJSON(m model) string {
	rows := make([]string, len(m.effects))
	for i, e := range m.effects {
		rows[i] = fmt.Sprintf("    [%d, [%s]]", e.id, quoteKeys(e.keys, ","))
	}
	return fmt.Sprintf(`{
  "pyfaAttributeBoostCalls": %d,
  "penalisedCalls": %d,
  "receiverQualifiedKeys": %d,
  "effects": [
%s
  ]
}
`, m.calls, m.penalisedCalls, m.ambiguous, strings.Join(rows, ",\n"))
}

func parseModel(raw []byte) (model, error) {
	var j struct {
		Calls          int                 `json:"pyfaAttributeBoostCalls"`
		PenalisedCalls int                 `json:"penalisedCalls"`
		Ambiguous      int                 `json:"receiverQualifiedKeys"`
		Effects        [][]json.RawMessage `json:"effects"`
	}
	if err := json.Unmarshal(raw, &j); err != nil {
		return model{}, err
	}
	m := model{calls: j.Calls, penalisedCalls: j.PenalisedCalls, ambiguous: j.Ambiguous}
	for _, row := range j.Effects {
		if len(row) != 2 {
			return model{}, errors.New("malformed effects row")
		}
		var e effectKeys
		if err := json.Unmarshal(row[0], &e.id); err != nil {
			return model{}, err
		}
		if err := json.Unmarshal(row[1], &e.keys); err != nil {
			return model{}, err
		}
		m.effects = append(m.effects, e)
	}
	return m, nil
}

// fetchSDE stages the raw SDE tables the analysis needs (attribute names -> ids, effect modifierInfo)
// from CCP's SDE API into .sde-src/, as served (API shape; sdeKey reads both shapes).
func fetchSDE() error {
	client := &http.Client{Timeout: 120 * time.Second}
	get := func(url string) ([]byte, error) {
		var last error
		for attempt := 1; attempt <= 5; attempt++ {
			res, err := client.Get(url)
			if err == nil {
				body, rerr := io.ReadAll(res.Body)
				res.Body.Close()
				if res.StatusCode == http.StatusOK && rerr == nil {
					return body, nil
				}
				err = fmt.Errorf("HTTP %d %v", res.StatusCode, rerr)
			}
			last = err
			if attempt < 5 {
				wait := time.Duration(2<<(attempt-1)) * time.Second
				fmt.Fprintf(os.Stderr, "[stacking] %s attempt %d failed (%v); retry in %s\n", url, attempt, err, wait)
				time.Sleep(wait)
			}
		}
		return nil, fmt.Errorf("%s: %v", url, last)
	}
	raw, err := get(sdeIndexURL)
	if err != nil {
		return err
	}
	var index struct {
		Release     string `json:"release"`
		BuildNumber int    `json:"buildNumber"`
		Tables      map[string]struct {
			JSONL struct {
				URL string `json:"url"`
			} `json:"jsonl"`
		} `json:"tables"`
	}
	if err := json.Unmarshal(raw, &index); err != nil {
		return fmt.Errorf("SDE index: %v", err)
	}
	if index.Release == "" || index.BuildNumber == 0 {
		return errors.New("SDE index: missing release/buildNumber")
	}
	if err := os.MkdirAll(sdeDir, 0o755); err != nil {
		return err
	}
	for _, t := range sdeTables {
		url := index.Tables[t].JSONL.URL
		if url == "" {
			return fmt.Errorf("SDE index: table %s has no jsonl url", t)
		}
		body, err := get(url)
		if err != nil {
			return err
		}
		if len(bytes.TrimSpace(body)) == 0 || bytes.HasPrefix(body, []byte(`{"error"`)) {
			return fmt.Errorf("%s: empty or error payload", t)
		}
		if err := os.WriteFile(filepath.Join(sdeDir, t+".jsonl"), body, 0o644); err != nil {
			return err
		}
		fmt.Printf("[stacking] %s: %.2f MB\n", t, float64(len(body))/1048576)
	}
	_ = os.WriteFile(filepath.Join(sdeDir, ".build"), []byte(strconv.Itoa(index.BuildNumber)), 0o644)
	_ = os.WriteFile(filepath.Join(sdeDir, ".release"), []byte(index.Release), 0o644)
	fmt.Printf("[stacking] staged %d tables into %s (release %s, build %d)\n", len(sdeTables), sdeDir, index.Release, index.BuildNumber)
	return nil
}

// renderGo: the Go registry. No maps (lookup by binary search), same semantics as the TS
// stackPenaltyGroup of eve-fit-engine.
func renderGo(m model) string {
	rows := make([]string, len(m.effects))
	for i, e := range m.effects {
		rows[i] = fmt.Sprintf("\t{%d, []string{%s}},", e.id, quoteKeys(e.keys, ", "))
	}
	return fmt.Sprintf(goHeader, m.penalisedCalls, m.calls, len(m.effects), m.ambiguous) + strings.Join(rows, "\n") + goTail
}

const goHeader = `// Code generated by tools/stacking from generated/stacking-penalised.json; DO NOT EDIT.
// Regenerate with "go run ./tools/stacking". pyfa penalises %d of its %d attribute-boosting calls,
// across %d effects; %d keys are receiver-qualified.

package gen

import "sort"

// StackingPenalisedEntry: effect id -> penalised keys, each "[receiver:]attrId#penaltyGroup[!exemptContext+...]"
// ("*" where pyfa computes the attribute name at runtime).
type StackingPenalisedEntry struct {
	EffectID int
	Keys     []string
}

// StackingPenalised is sorted by EffectID ascending.
var StackingPenalised = []StackingPenalisedEntry{
`

const goTail = `
}

// PenaltyReceiver: receiver classes used to qualify an ambiguous key.
type PenaltyReceiver string

const (
	ReceiverShip      PenaltyReceiver = "ship"
	ReceiverModule    PenaltyReceiver = "module"
	ReceiverCharge    PenaltyReceiver = "charge"
	ReceiverDrone     PenaltyReceiver = "drone"
	ReceiverFighter   PenaltyReceiver = "fighter"
	ReceiverCharacter PenaltyReceiver = "character"
	ReceiverSelf      PenaltyReceiver = "self"
)

// contextKinds: pyfa's context names, as the engine's source kinds (CONTEXT_KIND).
func contextKinds(c string) []string {
	switch c {
	case "skill":
		return []string{"skill", "character"}
	case "ship":
		return []string{"ship"}
	case "implant":
		return []string{"implant"}
	case "booster":
		return []string{"booster"}
	case "module":
		return []string{"module"}
	case "drone":
		return []string{"drone"}
	case "fighter":
		return []string{"fighter"}
	case "subsystem":
		return []string{"subsystem"}
	case "mode":
		return []string{"mode"}
	}
	return nil
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func lastIndexByte(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func splitPlus(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '+' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// StackPenaltyGroup: the pyfa penalty GROUP this effect uses for this attribute on this target, or
// ok=false when it does not penalise. attributeID nil = undefined.
func StackPenaltyGroup(effectID int, attributeID *int, receiver PenaltyReceiver, sourceKind string) (group string, ok bool) {
	i := sort.Search(len(StackingPenalised), func(i int) bool { return StackingPenalised[i].EffectID >= effectID })
	if i == len(StackingPenalised) || StackingPenalised[i].EffectID != effectID {
		return "", false
	}
	keys := StackingPenalised[i].Keys
	id := ""
	hasID := attributeID != nil
	if hasID {
		id = itoa(*attributeID)
	}
	for _, key := range keys {
		bang := indexByte(key, '!')
		var exempt []string
		head := key
		if bang >= 0 {
			exempt = splitPlus(key[bang+1:])
			head = key[:bang]
		}
		hash := lastIndexByte(head, '#')
		group := head[hash+1:]
		target := head[:hash]
		colon := indexByte(target, ':')
		recv := ""
		hasRecv := colon >= 0
		attr := target
		if hasRecv {
			recv = target[:colon]
			attr = target[colon+1:]
		}
		if hasRecv && recv != string(receiver) {
			continue
		}
		if attr != "*" && (!hasID || attr != id) {
			continue
		}
		for _, c := range exempt {
			for _, k := range contextKinds(c) {
				if k == sourceKind {
					return "", false
				}
			}
		}
		return group, true
	}
	return "", false
}
`
