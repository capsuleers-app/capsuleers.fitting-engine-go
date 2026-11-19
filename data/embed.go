// Package data incorpora il bundle SDE (manifest.json + v<hash>/...) nel
// binario. go:embed non esce dal package, per questo l'embed sta qui dentro.
// Il bundle e' di CCP (vedi SDE-LICENSE.md), aggregato e non coperto dalla GPL.
package data

import (
	"embed"
	"encoding/json"
	"sync"
)

// FS contiene manifest.json e le cartelle v<hash>/ del bundle (radice = data/).
//
//go:embed manifest.json v*
var FS embed.FS

var (
	versionOnce sync.Once
	version     string
)

// Version e' il `version` di manifest.json (hash di 12 cifre esadecimali).
// Il manifest e' incorporato a build time: se manca o e' illeggibile e' un
// errore di build, quindi qui si va in panic.
func Version() string {
	versionOnce.Do(func() {
		raw, err := FS.ReadFile("manifest.json")
		if err != nil {
			panic("data: manifest.json non incorporato: " + err.Error())
		}
		var m struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal(raw, &m); err != nil || m.Version == "" {
			panic("data: manifest.json non valido")
		}
		version = m.Version
	})
	return version
}
