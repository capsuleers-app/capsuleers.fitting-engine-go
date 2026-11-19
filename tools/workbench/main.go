// Command workbench harvests a REAL-fit corpus from EVE Workbench for the differential harness
// (`go run ./tools/diff -source=workbench`). Port of scripts/fetch-workbench-corpus.mjs of eve-fit-engine:
// same API, same picking, same output file, so the two repos can share one cached corpus.
//
// Two upstream facts (measured, unchanged since 2026-08-07):
//  1. /v1/fits/public takes ONLY `page`: every filter is ignored, so the index is built by walking pages
//     and bucketing by ShipId on our side.
//  2. It advertises ~382 pages but only serves the first ~60; page 61+ hangs. maxPage encodes that.
//
// The API rate-limits on BURST, not on rate: EFT fetches run at concurrency 2, deliberately slow; the
// corpus is cached and a 429 aborts the pass (keeping what was fetched) rather than hammering through.
//
// Output: .workbench/corpus.json (gitignored: other capsuleers' published fits, read like the site does
// and never redistributed).
//
//	go run ./tools/workbench               # default 3 fits/hull
//	go run ./tools/workbench -per=2        # fewer
//	go run ./tools/workbench -pages=20     # shallower index walk
//	go run ./tools/workbench -refresh      # ignore the cached index
//
// Needs EVEWORKBENCH_API_KEY.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/jsmath"
)

const (
	base    = "https://api.eveworkbench.com"
	ua      = "Capsuleers.App/1.0 (+https://capsuleers.app; info@capsuleers.app)"
	maxPage = 60
	listC   = 4
	eftC    = 2 // burst-limited upstream: keep this at 2
)

var (
	cacheDir   = ".workbench"
	indexFile  = filepath.Join(cacheDir, "index.json")
	corpusFile = filepath.Join(cacheDir, "corpus.json")
	key        string
	errLimited = errors.New("429 from upstream")
)

type listRow struct {
	ID       string  `json:"Id"`
	ShipID   int     `json:"ShipId"`
	ShipName string  `json:"ShipName"`
	Name     string  `json:"Name"`
	TotalDps float64 `json:"TotalDps"`
}

type corpusRow struct {
	ID         string  `json:"id"`
	ShipTypeID int     `json:"shipTypeID"`
	ShipName   string  `json:"shipName"`
	Name       string  `json:"name"`
	Dps        float64 `json:"dps"`
	Eft        string  `json:"eft,omitempty"`
}

func main() {
	per := flag.Int("per", 3, "fits per hull")
	pages := flag.Int("pages", maxPage, "index pages to walk (max 60)")
	refresh := flag.Bool("refresh", false, "ignore the cached index")
	flag.Parse()
	if *pages > maxPage {
		*pages = maxPage
	}
	key = os.Getenv("EVEWORKBENCH_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "[corpus] set EVEWORKBENCH_API_KEY")
		os.Exit(1)
	}
	if err := chdirRoot(); err != nil {
		fail(err)
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		fail(err)
	}
	var rows []listRow
	if raw, err := os.ReadFile(indexFile); err == nil && !*refresh {
		if err := json.Unmarshal(raw, &rows); err != nil {
			fail(err)
		}
		logf("[corpus] reusing cached index (%d rows) — -refresh to re-walk", len(rows))
	} else {
		var err error
		if rows, err = buildIndex(*pages); err != nil {
			fail(err)
		}
		out, _ := json.Marshal(rows)
		if err := os.WriteFile(indexFile, out, 0o644); err != nil {
			fail(err)
		}
		logf("[corpus] indexed %d fits", len(rows))
	}

	picked := pickPerHull(rows, *per)
	logf("[corpus] %d fits across %d hulls — fetching EFT at concurrency %d", len(picked), countHulls(picked), eftC)

	known := map[string]corpusRow{}
	if raw, err := os.ReadFile(corpusFile); err == nil {
		var have []corpusRow
		if err := json.Unmarshal(raw, &have); err != nil {
			fail(err)
		}
		for _, f := range have {
			known[f.ID] = f
		}
	}
	var todo []corpusRow
	for _, p := range picked {
		if _, ok := known[p.ID]; !ok {
			todo = append(todo, p)
		}
	}
	logf("[corpus] %d cached, %d to fetch", len(known), len(todo))

	var mu sync.Mutex
	n, failed := 0, 0
	err := mapLimit(len(todo), eftC, func(i int) error {
		p := todo[i]
		var j struct {
			Eft string `json:"Eft"`
		}
		if err := wb("/v1/fits/"+url.PathEscape(p.ID)+"/eft", 20*time.Second, &j); err != nil {
			return err
		}
		eft := strings.TrimSpace(j.Eft)
		mu.Lock()
		defer mu.Unlock()
		n++
		if n%50 == 0 {
			logf("[corpus]   %d/%d EFTs", n, len(todo))
		}
		if eft == "" {
			failed++
			return nil
		}
		p.Eft = eft
		known[p.ID] = p
		return nil
	}, func() { mu.Lock(); failed++; mu.Unlock() })
	if err != nil {
		logf("[corpus] aborted: %v — writing what we have", err)
	}

	corpus := make([]corpusRow, 0, len(known))
	for _, f := range known {
		corpus = append(corpus, f)
	}
	sort.Slice(corpus, func(i, j int) bool {
		if corpus[i].ShipTypeID != corpus[j].ShipTypeID {
			return corpus[i].ShipTypeID < corpus[j].ShipTypeID
		}
		return jsmath.LocaleCompare(corpus[i].ID, corpus[j].ID) < 0
	})
	out, _ := json.Marshal(corpus)
	if err := os.WriteFile(corpusFile, out, 0o644); err != nil {
		fail(err)
	}
	suffix := ""
	if failed > 0 {
		suffix = fmt.Sprintf(" (%d failed)", failed)
	}
	logf("[corpus] wrote %d fits across %d hulls → %s%s", len(corpus), countHulls(corpus), corpusFile, suffix)
}

func logf(format string, a ...any) { fmt.Fprintf(os.Stderr, format+"\n", a...) }

func fail(err error) {
	fmt.Fprintln(os.Stderr, "[corpus]", err)
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

func countHulls(rows []corpusRow) int {
	h := map[int]bool{}
	for _, r := range rows {
		h[r.ShipTypeID] = true
	}
	return len(h)
}

func wb(path string, timeout time.Duration, out any) error {
	req, err := http.NewRequest(http.MethodGet, base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-API-KEY", key)
	req.Header.Set("User-Agent", ua)
	res, err := (&http.Client{Timeout: timeout}).Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusTooManyRequests {
		return errLimited
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s -> %d", path, res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(out)
}

// mapLimit runs fn over [0,n) with bounded concurrency and stops at the first 429 (returned). Other
// errors are counted through onErr and do not stop the pass.
func mapLimit(n, limit int, fn func(i int) error, onErr func()) error {
	var mu sync.Mutex
	next := 0
	var fatal error
	var wg sync.WaitGroup
	for w := 0; w < limit && w < n; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				mu.Lock()
				if next >= n || fatal != nil {
					mu.Unlock()
					return
				}
				i := next
				next++
				mu.Unlock()
				if err := fn(i); err != nil {
					if errors.Is(err, errLimited) {
						mu.Lock()
						fatal = err
						mu.Unlock()
					} else if onErr != nil {
						onErr()
					}
				}
			}
		}()
	}
	wg.Wait()
	return fatal
}

// buildIndex walks the public list and returns every fit row it serves.
func buildIndex(pages int) ([]listRow, error) {
	var first struct {
		NumberOfPages int       `json:"NumberOfPages"`
		TotalFits     int       `json:"TotalFits"`
		Fits          []listRow `json:"Fits"`
	}
	if err := wb("/v1/fits/public?page=1", 30*time.Second, &first); err != nil {
		return nil, err
	}
	total := first.NumberOfPages
	if total < 1 {
		total = 1
	}
	if total > pages {
		total = pages
	}
	logf("[corpus] upstream declares %d pages / %d fits; walking %d", first.NumberOfPages, first.TotalFits, total)
	rows := append([]listRow(nil), first.Fits...)
	pageRows := make([][]listRow, total+1)
	var mu sync.Mutex
	done := 1
	err := mapLimit(total-1, listC, func(i int) error {
		p := i + 2
		var j struct {
			Fits []listRow `json:"Fits"`
		}
		if err := wb(fmt.Sprintf("/v1/fits/public?page=%d", p), 30*time.Second, &j); err != nil {
			return err
		}
		mu.Lock()
		pageRows[p] = j.Fits
		done++
		if done%10 == 0 {
			logf("[corpus]   %d/%d pages", done, total)
		}
		mu.Unlock()
		return nil
	}, nil)
	if err != nil {
		return nil, err
	}
	for _, pr := range pageRows {
		rows = append(rows, pr...)
	}
	return rows, nil
}

// pickPerHull picks up to `per` fits per hull, spread across the hull's own list (three fits published
// the same week by the same meta are three copies of one fit). Deterministic: index arithmetic, no RNG.
func pickPerHull(rows []listRow, per int) []corpusRow {
	var order []int
	byShip := map[int][]listRow{}
	for _, f := range rows {
		if f.ID == "" || f.ShipID == 0 {
			continue
		}
		if _, ok := byShip[f.ShipID]; !ok {
			order = append(order, f.ShipID)
		}
		byShip[f.ShipID] = append(byShip[f.ShipID], f)
	}
	var picked []corpusRow
	for _, ship := range order {
		list := byShip[ship]
		sort.SliceStable(list, func(i, j int) bool { return jsmath.LocaleCompare(list[i].ID, list[j].ID) < 0 })
		n := per
		if len(list) < n {
			n = len(list)
		}
		for i := 0; i < n; i++ {
			f := list[int(math.Floor(float64(i*len(list))/float64(n)))]
			picked = append(picked, corpusRow{ID: f.ID, ShipTypeID: ship, ShipName: f.ShipName, Name: f.Name, Dps: f.TotalDps})
		}
	}
	sort.SliceStable(picked, func(i, j int) bool {
		if picked[i].ShipTypeID != picked[j].ShipTypeID {
			return picked[i].ShipTypeID < picked[j].ShipTypeID
		}
		return jsmath.LocaleCompare(picked[i].ID, picked[j].ID) < 0
	})
	return picked
}
