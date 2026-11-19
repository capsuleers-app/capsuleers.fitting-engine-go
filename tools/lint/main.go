// Command lint e' l'analyzer dell'engine (gate bloccante in CI), solo libreria standard.
//
//	go run ./tools/lint
//
// Regole, sui package dell'engine (internal/engine, internal/derived, internal/effects, fitengine):
//
//	no-map-range  vietato `range` su map Go: l'ordine e' casuale, si usa internal/omap.
//	fma-cast      un prodotto float che entra in una somma/sottrazione va scritto float64(x*y) + z:
//	              Go puo' fondere x*y+z in un FMA (arm64, ppc64le, s390x, riscv64, amd64 con GOAMD64=v3)
//	              e il risultato cambierebbe nell'ultimo bit rispetto a V8; la conversione esplicita lo vieta.
//
// Salta i file _test.go e quelli generati (internal/gen). Esce con 1 se trova violazioni.
package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var packages = []string{"internal/engine", "internal/derived", "internal/effects", "fitengine"}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	var problems []string
	for _, rel := range packages {
		dir := filepath.Join(root, rel)
		if _, err := os.Stat(dir); err != nil {
			continue // package non ancora portato
		}
		ps, err := checkDir(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "lint:", err)
			os.Exit(2)
		}
		problems = append(problems, ps...)
	}
	sort.Strings(problems)
	for _, p := range problems {
		fmt.Println(p)
	}
	if len(problems) > 0 {
		fmt.Fprintf(os.Stderr, "lint: %d violazioni\n", len(problems))
		os.Exit(1)
	}
}

func checkDir(dir string) ([]string, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, pkg := range pkgs {
		var files []*ast.File
		for _, f := range pkg.Files {
			files = append(files, f)
		}
		info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}}
		conf := types.Config{Importer: importer.ForCompiler(fset, "source", nil), Error: func(error) {}}
		_, _ = conf.Check(pkg.Name, fset, files, info) // gli errori di tipo non fermano l'analisi
		for _, f := range files {
			ast.Inspect(f, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.RangeStmt:
					if tv, ok := info.Types[x.X]; ok && tv.Type != nil {
						if _, isMap := tv.Type.Underlying().(*types.Map); isMap {
							out = append(out, fmt.Sprintf("%s: no-map-range: range su map (usare internal/omap)", fset.Position(x.Pos())))
						}
					}
				case *ast.BinaryExpr:
					if x.Op == token.ADD || x.Op == token.SUB {
						for _, side := range []ast.Expr{x.X, x.Y} {
							if p := unprotectedProduct(info, side); p != nil {
								out = append(out, fmt.Sprintf("%s: fma-cast: prodotto in una somma, scrivere float64(a*b)", fset.Position(p.Pos())))
							}
						}
					}
				case *ast.AssignStmt:
					if (x.Tok == token.ADD_ASSIGN || x.Tok == token.SUB_ASSIGN) && len(x.Rhs) == 1 {
						if p := unprotectedProduct(info, x.Rhs[0]); p != nil {
							out = append(out, fmt.Sprintf("%s: fma-cast: prodotto in += / -=, scrivere float64(a*b)", fset.Position(p.Pos())))
						}
					}
				}
				return true
			})
		}
	}
	return out, nil
}

// unprotectedProduct restituisce e se e' (a meno di parentesi) una moltiplicazione float non costante:
// una conversione esplicita float64(...) e' una CallExpr e quindi non viene riconosciuta.
func unprotectedProduct(info *types.Info, e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			break
		}
		e = p.X
	}
	b, ok := e.(*ast.BinaryExpr)
	if !ok || b.Op != token.MUL {
		return nil
	}
	tv, ok := info.Types[e]
	if !ok || tv.Type == nil || tv.Value != nil { // costante: calcolata a compile time
		return nil
	}
	if basic, ok := tv.Type.Underlying().(*types.Basic); ok && basic.Info()&types.IsFloat != 0 {
		return e
	}
	return nil
}
