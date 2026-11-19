package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintFlagsViolations(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "internal", "engine")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := `package engine
func f(m map[int]float64, a, b, c float64) float64 {
	s := 0.0
	for k := range m { s += float64(k) }       // no-map-range
	s += a * b                                 // fma-cast
	s = a*b + c                                // fma-cast
	s = c - (a * b)                            // fma-cast
	s = float64(a*b) + c                       // ok
	s += float64(a * b)                        // ok
	s = a*b*c + 0                              // fma-cast (a*b*c e' un prodotto)
	const k = 2.0 * 3.0
	s = k*2 + 1                                // ok: costante
	return s
}
`
	if err := os.WriteFile(filepath.Join(dir, "x.go"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	ps, err := checkDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps) != 5 {
		t.Fatalf("violazioni attese 5, trovate %d: %v", len(ps), ps)
	}
}
