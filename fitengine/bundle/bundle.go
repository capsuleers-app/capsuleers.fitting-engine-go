// Package bundle e' l'equivalente di src/node.ts: carica il bundle SDE (quello
// incorporato in data/, una cartella o un fs.FS) in un fitengine.FittingDataset.
package bundle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/data"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine/checks"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine/eft"
	"github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"
)

// FS e' il bundle incorporato: manifest.json + v<hash>/ alla radice.
func FS() fs.FS { return data.FS }

// Version e' la versione (hash SDE) del bundle incorporato.
func Version() string { return data.Version() }

// LoadDatasetDir carica un bundle da una cartella (stesso layout di data/).
func LoadDatasetDir(dir string) (*fitengine.FittingDataset, error) {
	return LoadDataset(os.DirFS(dir))
}

type manifest struct {
	Version string `json:"version"`
}

// LoadDataset legge manifest.json e v<version>/ da fsys. Come loadBundledDataset
// del TS: market-groups.json e i bucket opzionali possono mancare (tabella
// vuota), ogni altro file mancante e' un errore. I campi sconosciuti nel JSON
// sono ignorati.
func LoadDataset(fsys fs.FS) (*fitengine.FittingDataset, error) {
	raw, err := fs.ReadFile(fsys, "manifest.json")
	if err != nil {
		return nil, fmt.Errorf("bundle: manifest.json: %w", err)
	}
	var m manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("bundle: manifest.json: %w", err)
	}
	if m.Version == "" {
		return nil, errors.New("bundle: manifest.json senza version")
	}
	vdir := "v" + m.Version
	ds := &fitengine.FittingDataset{Version: m.Version}

	var steps = []struct {
		file     string
		optional bool
		load     func(raw []byte) error
	}{
		{"attributes.json", false, tableLoader(&ds.Attributes)},
		{"units.json", false, tableLoader(&ds.Units)},
		{"effects.json", false, tableLoader(&ds.Effects)},
		{"meta-groups.json", false, tableLoader(&ds.MetaGroups)},
		{"categories.json", false, tableLoader(&ds.Categories)},
		{"groups.json", false, tableLoader(&ds.Groups)},
		{"market-groups.json", true, tableLoader(&ds.MarketGroups)},
		{"clone-grades.json", false, tableLoader(&ds.CloneGrades)},
		{"dbuff-collections.json", false, tableLoader(&ds.DbuffCollections)},
		{"dynamic-attributes.json", false, tableLoader(&ds.DynamicAttributes)},
	}
	for b := fitengine.TypeBucket(0); b < fitengine.NumBuckets; b++ {
		steps = append(steps, struct {
			file     string
			optional bool
			load     func(raw []byte) error
		}{fitengine.BucketFiles[b], fitengine.BucketOptional[b], tableLoader(&ds.TypesByBucket[b])})
	}
	for _, s := range steps {
		raw, err := fs.ReadFile(fsys, vdir+"/"+s.file)
		if err != nil {
			if s.optional && errors.Is(err, fs.ErrNotExist) {
				if err := s.load([]byte("{}")); err != nil {
					return nil, err
				}
				continue
			}
			return nil, fmt.Errorf("bundle: %s: %w", s.file, err)
		}
		if err := s.load(raw); err != nil {
			return nil, fmt.Errorf("bundle: %s: %w", s.file, err)
		}
	}
	return ds, nil
}

func tableLoader[V any](dst *fitengine.Table[V]) func([]byte) error {
	return func(raw []byte) error {
		t, err := decodeTable[V](raw)
		if err != nil {
			return err
		}
		*dst = t
		return nil
	}
}

// decodeTable decodifica un oggetto `{ "<id>": {...} }` in una Table ordinata per
// id ascendente, che e' l'ordine di iterazione JS per chiavi intere. Le chiavi
// che JS tratterebbe diversamente (non intere canoniche, o oltre l'indice di
// array 2^32-2: ordine di inserimento, dopo gli interi) non sono supportate e
// sono un errore, cosi' un cambio di formato del bundle non passa in silenzio.
func decodeTable[V any](raw []byte) (fitengine.Table[V], error) {
	var rec map[string]V
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&rec); err != nil {
		return fitengine.Table[V]{}, err
	}
	keys := make([]int, 0, len(rec))
	byKey := make(map[int]V, len(rec))
	for k, v := range rec {
		n, err := strconv.ParseUint(k, 10, 32)
		if err != nil || n > 4294967294 || strconv.FormatUint(n, 10) != k {
			return fitengine.Table[V]{}, fmt.Errorf("chiave %q non e' un indice intero canonico (ordine JS non riproducibile)", k)
		}
		keys = append(keys, int(n))
		byKey[int(n)] = v
	}
	sort.Ints(keys)
	values := make([]V, len(keys))
	for i, k := range keys {
		values[i] = byKey[k]
	}
	return fitengine.NewTable(keys, values), nil
}

// BuildAllVSkillProfile e' buildAllVSkillProfile di node.ts: ogni skill del bucket a livello 5, in ordine di id.
func BuildAllVSkillProfile(ds *fitengine.FittingDataset) *fitengine.SkillProfile {
	skills := omap.New[int, int]()
	bucket := &ds.TypesByBucket[fitengine.BucketSkills]
	for i := 0; i < bucket.Len(); i++ {
		id, _ := bucket.At(i)
		skills.Set(id, 5)
	}
	return &fitengine.SkillProfile{Name: "All V", IsDefault: true, Source: "preset", Skills: skills}
}

var (
	bundledOnce sync.Once
	bundled     *fitengine.FittingDataset
	bundledErr  error
)

// LoadBundledDataset carica (una volta sola, poi in cache) il bundle incorporato in data/.
func LoadBundledDataset() (*fitengine.FittingDataset, error) {
	bundledOnce.Do(func() { bundled, bundledErr = LoadDataset(FS()) })
	return bundled, bundledErr
}

// ComputeFromEftResult e' ComputeFromEftResult di node.ts.
type ComputeFromEftResult struct {
	Fit      *fitengine.Fit
	Warnings []eft.Warning
	Computed *fitengine.ComputedFit
}

// ComputeFromEft: da testo EFT alle statistiche complete con il bundle incorporato. Stati di default per
// modulo (DefaultStateForModule); skill All-V se options.SkillProfile e' nil.
func ComputeFromEft(text string, options *fitengine.ComputeFitOptions) (*ComputeFromEftResult, error) {
	dataset, err := LoadBundledDataset()
	if err != nil {
		return nil, err
	}
	parsed, err := eft.ParseEft(text, dataset)
	if err != nil {
		return nil, err
	}
	fit := parsed.Fit
	for i := range fit.Modules {
		m := &fit.Modules[i]
		if t, ok := dataset.GetType(m.TypeID); ok {
			m.State = checks.DefaultStateForModule(t, &dataset.Effects)
		}
	}
	opts := fitengine.ComputeFitOptions{}
	if options != nil {
		opts = *options
	}
	if opts.SkillProfile == nil {
		opts.SkillProfile = BuildAllVSkillProfile(dataset)
	}
	computed, err := fitengine.ComputeFit(fit, dataset, &opts)
	if err != nil {
		return nil, err
	}
	return &ComputeFromEftResult{Fit: fit, Warnings: parsed.Warnings, Computed: computed}, nil
}
