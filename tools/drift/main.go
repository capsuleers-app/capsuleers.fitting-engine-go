// Command drift is the pyfa drift detector (port of `npm run drift` of eve-fit-engine).
//
// It fetches pyfa master's eos/effects.py, extracts every hardcoded `class Effect<N>(BaseEffect)` (the
// effects whose semantics live in hand-written Python rather than SDE modifierInfo), and:
//  1. diffs the set against the committed snapshot (tools/drift/effects-snapshot.json): effect ids pyfa
//     ADDED or REMOVED since the last baseline;
//  2. cross-references this engine's LegacyEffectIDs: new pyfa effects we do not hardcode (check
//     data-driven vs needs a handler) and changed effects we DO hardcode (re-verify parity).
//
// SDE balance changes flow through data/ by themselves; NEW hardcoded mechanics are code and need a human.
//
//	go run ./tools/drift            # report only
//	go run ./tools/drift -update    # rewrite the snapshot (advance the baseline)
//
// Exit: 0 no drift, 2 drift detected (CI opens an issue), 1 error. `go run` reports every non-zero exit
// as 1: build the binary (as pyfa-drift.yml does) when the code matters.
package main

import (
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

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/engine"
)

const effectsURL = "https://raw.githubusercontent.com/pyfa-org/Pyfa/master/eos/effects.py"

var snapshotFile = filepath.Join("tools", "drift", "effects-snapshot.json")

var classRe = regexp.MustCompile(`(?s)class Effect(\d+)\(BaseEffect\):\s*\n\s*(?:"""(.*?)"""|'''(.*?)''')?`)

// parseEffectsPy: effect id -> first docstring line (the effect name), nil when absent.
func parseEffectsPy(src string) map[int]*string {
	out := map[int]*string{}
	for _, m := range classRe.FindAllStringSubmatchIndex(src, -1) {
		id, _ := strconv.Atoi(src[m[2]:m[3]])
		doc := ""
		switch {
		case m[4] >= 0:
			doc = src[m[4]:m[5]]
		case m[6] >= 0:
			doc = src[m[6]:m[7]]
		}
		name := strings.TrimSpace(strings.SplitN(strings.TrimSpace(doc), "\n", 2)[0])
		if name == "" {
			out[id] = nil
		} else {
			n := name
			out[id] = &n
		}
	}
	return out
}

type snapshot struct {
	Source  string             `json:"source"`
	Count   int                `json:"count"`
	Effects map[string]*string `json:"effects"`
}

func str(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func main() {
	update := flag.Bool("update", false, "rewrite the snapshot")
	flag.Parse()
	if err := chdirRoot(); err != nil {
		fail(err)
	}
	client := &http.Client{Timeout: 2 * time.Minute}
	res, err := client.Get(effectsURL)
	if err != nil {
		fail(err)
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		fail(err)
	}
	if res.StatusCode != http.StatusOK {
		fail(fmt.Errorf("fetch effects.py -> %d", res.StatusCode))
	}
	current := parseEffectsPy(string(body))
	fmt.Printf("Pyfa effects.py: %d hardcoded Effect classes.\n", len(current))

	prev := map[int]*string{}
	if raw, err := os.ReadFile(snapshotFile); err == nil {
		var s snapshot
		if json.Unmarshal(raw, &s) == nil {
			for k, v := range s.Effects {
				if id, err := strconv.Atoi(k); err == nil {
					prev[id] = v
				}
			}
		}
	}
	ours := map[int]bool{}
	for _, e := range engine.LegacyEffectIDs {
		ours[e.ID] = true
	}
	var added, removed, renamedCovered []int
	for id := range current {
		if _, ok := prev[id]; !ok {
			added = append(added, id)
		} else if str(prev[id]) != str(current[id]) || (prev[id] == nil) != (current[id] == nil) {
			if ours[id] {
				renamedCovered = append(renamedCovered, id)
			}
		}
	}
	for id := range prev {
		if _, ok := current[id]; !ok {
			removed = append(removed, id)
		}
	}
	sort.Ints(added)
	sort.Ints(removed)
	sort.Ints(renamedCovered)
	uncovered := 0
	drift := false
	if len(added) > 0 {
		drift = true
		fmt.Printf("\n+ %d effect(s) ADDED upstream:\n", len(added))
		// pyfa hardcodes ~2 400 effects, we hardcode ~80 (the rest is data-driven): "not in our set" means
		// "check data-driven vs needs a handler", not "unhandled".
		for _, id := range added {
			note := "  -> not in our hardcoded set; check data-driven vs needs-handler"
			if ours[id] {
				note = "  (we hardcode this id — re-verify)"
			} else {
				uncovered++
			}
			fmt.Printf("   Effect%d  %q%s\n", id, str(current[id]), note)
		}
	}
	if len(removed) > 0 {
		drift = true
		fmt.Printf("\n- %d effect(s) REMOVED upstream:\n", len(removed))
		for _, id := range removed {
			note := ""
			if ours[id] {
				note = "  WARNING: WE STILL HARDCODE THIS — review"
			}
			fmt.Printf("   Effect%d  %q%s\n", id, str(prev[id]), note)
		}
	}
	if len(renamedCovered) > 0 {
		drift = true
		fmt.Printf("\n~ %d effect(s) we cover were RENAMED/changed upstream — re-verify parity:\n", len(renamedCovered))
		for _, id := range renamedCovered {
			fmt.Printf("   Effect%d: %q -> %q\n", id, str(prev[id]), str(current[id]))
		}
	}
	if !drift {
		fmt.Println("\nOK: no drift vs snapshot. Pyfa hardcoded-effect set unchanged.")
	} else {
		fmt.Printf("\nSummary: %d added (%d uncovered), %d removed, %d covered-and-changed.\n", len(added), uncovered, len(removed), len(renamedCovered))
		fmt.Println("Action: port/verify the affected legacy handlers (in BOTH repos), then run the parity suite before release.")
	}
	if *update {
		if err := writeSnapshot(current); err != nil {
			fail(err)
		}
		fmt.Printf("\nSnapshot rewritten (%d effects).\n", len(current))
		return
	}
	if drift {
		os.Exit(2)
	}
}

// writeSnapshot writes the snapshot like JSON.stringify(_, null, 2) (numeric keys ascending), so the
// file is byte-compatible with eve-fit-engine's scripts/effects-snapshot.json.
func writeSnapshot(effects map[int]*string) error {
	ids := make([]int, 0, len(effects))
	for id := range effects {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	enc := func(v any) string {
		var b bytes.Buffer
		e := json.NewEncoder(&b)
		e.SetEscapeHTML(false)
		_ = e.Encode(v)
		return strings.TrimSuffix(b.String(), "\n")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "{\n  \"source\": %s,\n  \"count\": %d,\n  \"effects\": {\n", enc(effectsURL), len(ids))
	for i, id := range ids {
		sep := ","
		if i == len(ids)-1 {
			sep = ""
		}
		fmt.Fprintf(&b, "    \"%d\": %s%s\n", id, enc(effects[id]), sep)
	}
	b.WriteString("  }\n}\n")
	return os.WriteFile(snapshotFile, []byte(b.String()), 0o644)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "[drift]", err)
	os.Exit(1)
}

func chdirRoot() error {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	gomod := strings.TrimSpace(string(out))
	if err != nil || gomod == "" || gomod == os.DevNull {
		return errors.New("not inside the Go module")
	}
	return os.Chdir(filepath.Dir(gomod))
}
