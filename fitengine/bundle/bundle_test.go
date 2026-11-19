package bundle

import (
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/capsuleers-app/capsuleers.fitting-engine-go/fitengine"
)

func TestEmbeddedAndDirAgree(t *testing.T) {
	a, err := LoadDataset(FS())
	if err != nil {
		t.Fatal(err)
	}
	if a.Version != Version() {
		t.Fatalf("versione dataset %s != Version() %s", a.Version, Version())
	}
	b, err := LoadDatasetDir("../../data")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.Effects.Keys(), b.Effects.Keys()) || a.TypesByBucket[fitengine.BucketModules].Len() != b.TypesByBucket[fitengine.BucketModules].Len() {
		t.Fatal("incorporato e cartella divergono")
	}
	// Chiavi in ordine ascendente e GetType attraverso i bucket.
	for _, k := range [][]int{a.Attributes.Keys(), a.Effects.Keys(), a.TypesByBucket[fitengine.BucketModules].Keys()} {
		for i := 1; i < len(k); i++ {
			if k[i-1] >= k[i] {
				t.Fatalf("chiavi non ascendenti: %d, %d", k[i-1], k[i])
			}
		}
	}
	if ty, ok := a.GetType(587); !ok || ty.Name != "Rifter" {
		t.Fatalf("GetType(587) = %v, %v", ty, ok)
	}
	if _, ok := a.GetType(-1); ok {
		t.Fatal("GetType(-1) non deve esistere")
	}
}

func miniFS(extra map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{"manifest.json": {Data: []byte(`{"version":"abc123"}`)}}
	for _, f := range []string{"attributes", "units", "effects", "meta-groups", "categories", "groups",
		"clone-grades", "dbuff-collections", "dynamic-attributes"} {
		fsys["vabc123/"+f+".json"] = &fstest.MapFile{Data: []byte(`{}`)}
	}
	for _, f := range []string{"ships", "modules", "charges", "drones", "fighters", "implants", "subsystems", "skills"} {
		fsys["vabc123/types/"+f+".json"] = &fstest.MapFile{Data: []byte(`{}`)}
	}
	for k, v := range extra {
		fsys[k] = &fstest.MapFile{Data: []byte(v)}
	}
	return fsys
}

func TestOptionalFilesMissing(t *testing.T) {
	ds, err := LoadDataset(miniFS(nil)) // niente market-groups ne' bucket opzionali
	if err != nil {
		t.Fatal(err)
	}
	if ds.MarketGroups.Len() != 0 || ds.TypesByBucket[fitengine.BucketMutaplasmids].Len() != 0 {
		t.Fatal("i file opzionali mancanti devono dare tabelle vuote")
	}
}

func TestRequiredFileMissingAndBadKey(t *testing.T) {
	fsys := miniFS(nil)
	delete(fsys, "vabc123/groups.json")
	if _, err := LoadDataset(fsys); err == nil || !strings.Contains(err.Error(), "groups.json") {
		t.Fatalf("errore atteso su groups.json, ottenuto %v", err)
	}
	for _, bad := range []string{`{"01":{"id":1}}`, `{"x":{"id":1}}`, `{"-1":{"id":1}}`} {
		if _, err := LoadDataset(miniFS(map[string]string{"vabc123/units.json": bad})); err == nil {
			t.Fatalf("chiave non canonica accettata: %s", bad)
		}
	}
	ds, err := LoadDataset(miniFS(map[string]string{"vabc123/units.json": `{"10":{"id":10,"name":"b"},"9":{"id":9,"name":"a"}}`}))
	if err != nil {
		t.Fatal(err)
	}
	if k := ds.Units.Keys(); len(k) != 2 || k[0] != 9 || k[1] != 10 {
		t.Fatalf("ordine per id numerico, non lessicografico: %v", k)
	}
}
