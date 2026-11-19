package main

// Known, accepted differences vs pyfa (test/diff/known-diffs.json) and the pin recalibration
// (port of scripts/recalibrate-pyfa-pin.mjs of eve-fit-engine).
//
// The file is AUTO-MANAGED by -recalibrate, but the REASONS are human-owned: each entry carries the ROOT
// CAUSE of a pyfa float/modelling/per-ship quirk an independent engine can't match without regressing
// the 662-assertion fixture suite. It is not a place to hide real bugs. Matching is by
// (ship, fitType, statKey); a matched diff does not fail the run, ANY unlisted diff does. Entries with the
// PENDING reason were added at a pin bump and MUST be classified before a release.
//
// Invariant: never regress test/parity (662/0) to satisfy this harness, never widen the list to turn a
// red run green.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
)

var knownDiffsFile = filepath.Join("test", "diff", "known-diffs.json")

const pinFile = "PYFA_REF"

const pendingReason = "PENDING REVIEW — classify this diff (real bug → fix engine; pyfa quirk → write the root cause) before release"

const knownDiffsNote = "Accepted differences vs pyfa-org/Pyfa, calibrated against pyfaRef (= PYFA_REF). " +
	"Auto-managed by `go run ./tools/diff -recalibrate`; reasons are human-owned root causes, never a place to hide bugs. " +
	"Never regress test/parity (662/0) to satisfy the harness."

type knownDiff struct {
	Ship    string `json:"ship"`
	FitType string `json:"fitType"`
	Key     string `json:"key"`
	Reason  string `json:"reason"`
}

type knownDiffsDoc struct {
	Note    string      `json:"_note"`
	PyfaRef string      `json:"pyfaRef"`
	Entries []knownDiff `json:"entries"`
}

type knownSet struct {
	doc     knownDiffsDoc
	reasons map[string]string
}

func diffKey(ship, fitType, key string) string { return ship + "|" + fitType + "|" + key }

func loadKnownDiffs() (*knownSet, error) {
	raw, err := os.ReadFile(knownDiffsFile)
	if err != nil {
		return nil, err
	}
	ks := &knownSet{reasons: map[string]string{}}
	if err := json.Unmarshal(raw, &ks.doc); err != nil {
		return nil, fmt.Errorf("%s: %v", knownDiffsFile, err)
	}
	for _, e := range ks.doc.Entries {
		ks.reasons[diffKey(e.Ship, e.FitType, e.Key)] = e.Reason
	}
	return ks, nil
}

func (k *knownSet) has(d statDiff) bool {
	_, ok := k.reasons[diffKey(d.Ship, d.FitType, d.Key)]
	return ok
}

func (k *knownSet) reason(d statDiff) string { return k.reasons[diffKey(d.Ship, d.FitType, d.Key)] }

func writeKnownDiffs(doc knownDiffsDoc) error {
	if doc.Entries == nil {
		doc.Entries = []knownDiff{}
	}
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(knownDiffsFile, append(out, '\n'), 0o644)
}

func pyfaHead() (string, error) {
	out, err := exec.Command("git", "-C", pyfaDir(), "rev-parse", "HEAD").Output()
	if err != nil {
		return "", errors.New(".pyfa not found — run `bash scripts/setup-pyfa-oracle.sh` first")
	}
	return strings.TrimSpace(string(out)), nil
}

// recalibrate: the MECHANICAL half of a deliberate pin bump. Reads the current .pyfa HEAD as the new pin,
// runs the diff in strict mode, carries forward reasons for kept entries, drops resolved ones, marks new
// ones PENDING REVIEW, rewrites known-diffs.json and PYFA_REF. Exits 1 while anything is PENDING: the
// human still classifies new diffs (auto-accepting them would launder regressions).
func recalibrate(a args) {
	pin, err := pyfaHead()
	if err != nil {
		fatal(err.Error())
	}
	logf("[recalibrate] oracle pin (.pyfa HEAD): %s", pin)
	old, err := loadKnownDiffs()
	if err != nil {
		fatal(err.Error())
	}
	a.strict = true
	diffs, sum := run(a)
	if len(sum.Drift) > 0 {
		logf("[recalibrate] WARNING: %d type(s) absent from the pinned pyfa; %d fit(s) skipped (see `go run ./tools/diff`).", len(sum.Drift), sum.OracleSkipped)
	}
	current := map[string]bool{}
	var next, added, kept []knownDiff
	for _, d := range diffs {
		k := diffKey(d.Ship, d.FitType, d.Key)
		if current[k] {
			continue
		}
		current[k] = true
		r, had := old.reasons[k]
		reason := pendingReason
		if had && r != pendingReason {
			reason = r
		}
		e := knownDiff{d.Ship, d.FitType, d.Key, reason}
		next = append(next, e)
		if had {
			kept = append(kept, e)
		} else {
			added = append(added, e)
		}
	}
	var removed []knownDiff
	for _, e := range old.doc.Entries {
		if !current[diffKey(e.Ship, e.FitType, e.Key)] {
			removed = append(removed, e)
		}
	}
	sort.SliceStable(next, func(i, j int) bool {
		return jsmath.LocaleCompare(diffKey(next[i].Ship, next[i].FitType, next[i].Key), diffKey(next[j].Ship, next[j].FitType, next[j].Key)) < 0
	})
	if err := writeKnownDiffs(knownDiffsDoc{Note: knownDiffsNote, PyfaRef: pin, Entries: next}); err != nil {
		fatal(err.Error())
	}
	if err := os.WriteFile(pinFile, []byte(pin+"\n"), 0o644); err != nil {
		fatal(err.Error())
	}
	fmt.Printf("\n=== recalibration vs pin %s ===\n", pin[:7])
	fmt.Printf("  kept    : %d\n", len(kept))
	fmt.Printf("  removed : %d%s\n", len(removed), map[bool]string{true: " (resolved — dropped)", false: ""}[len(removed) > 0])
	for _, d := range removed {
		fmt.Printf("      - %s | %s | %s\n", d.Ship, d.FitType, d.Key)
	}
	fmt.Printf("  added   : %d%s\n", len(added), map[bool]string{true: "  ← NEEDS HUMAN CLASSIFICATION", false: ""}[len(added) > 0])
	for _, d := range added {
		fmt.Printf("      + %s | %s | %s\n", d.Ship, d.FitType, d.Key)
	}
	fmt.Printf("\nWrote %d entries to %s; PYFA_REF set to %s.\n", len(next), knownDiffsFile, pin)
	if len(added) > 0 {
		fmt.Printf("\nWARNING: %d entr%s PENDING REVIEW — classify each (real bug → fix the engine; pyfa quirk → write the root cause), then re-run.\n",
			len(added), map[bool]string{true: "y is", false: "ies are"}[len(added) == 1])
		os.Exit(1)
	}
	fmt.Println("\nOK: no new diffs — registry + pin updated, nothing to classify.")
}
