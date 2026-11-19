// Command syncdata keeps the embedded SDE bundle (data/) identical to the one eve-fit-engine ships.
//
// Versions are aligned: module tag vX.Y.Z carries the same data/ as the npm package eve-fit-engine@X.Y.Z.
// The bundle is CCP data (see data/SDE-LICENSE.md), not engine code, so it is taken verbatim from the
// published npm tarball rather than rebuilt: a rebuild could only target CCP's *latest* SDE, never the
// build a past version shipped, and a second bundle generator would be one more thing to keep at parity.
//
//	go run ./tools/syncdata                  # data/ := data/ of eve-fit-engine@$(cat VERSION)
//	                                         # (VERSION not on npm yet = this module is ahead of the TS
//	                                         # twin: the highest published version <= VERSION is used)
//	go run ./tools/syncdata -version 0.2.27  # another version (then write it into VERSION)
//	go run ./tools/syncdata -check           # drift check: exit 1 if data/ differs from the tarball
//	go run ./tools/syncdata -latest          # print the latest published version and exit
//
// The tarball is verified against the registry's sha512 `integrity` before anything is read from it.
// data/embed.go is this module's own file and is never touched or compared.
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	pkg      = "eve-fit-engine"
	registry = "https://registry.npmjs.org/"
	dataDir  = "data"
	ownFile  = "embed.go"
)

var client = &http.Client{Timeout: 5 * time.Minute}

func main() {
	version := flag.String("version", "", "eve-fit-engine version (default: contents of VERSION)")
	check := flag.Bool("check", false, "only compare data/ with the tarball; exit 1 on drift")
	latest := flag.Bool("latest", false, "print the latest published eve-fit-engine version and exit")
	flag.Parse()
	if err := chdirRoot(); err != nil {
		fail(err)
	}
	if *latest {
		v, err := latestVersion()
		if err != nil {
			fail(err)
		}
		fmt.Println(v)
		return
	}
	if *version == "" {
		raw, err := os.ReadFile("VERSION")
		if err != nil {
			fail(err)
		}
		want := strings.TrimSpace(string(raw))
		*version, err = publishedAtMost(want)
		if err != nil {
			fail(err)
		}
		if *version != want {
			fmt.Printf("[syncdata] %s@%s is not published yet (module ahead of the TS twin): data/ follows %s@%s\n", pkg, want, pkg, *version)
		}
	}
	files, err := tarballData(*version)
	if err != nil {
		fail(err)
	}
	local, err := localData()
	if err != nil {
		fail(err)
	}
	drift := compare(files, local)
	if *check {
		if len(drift) > 0 {
			fmt.Fprintf(os.Stderr, "[syncdata] data/ DRIFTS from %s@%s (%d file(s)):\n", pkg, *version, len(drift))
			for _, d := range drift {
				fmt.Fprintln(os.Stderr, "   ", d)
			}
			fmt.Fprintln(os.Stderr, "[syncdata] fix: go run ./tools/syncdata (never edit data/ by hand)")
			os.Exit(1)
		}
		fmt.Printf("[syncdata] data/ == %s@%s (%d files)\n", pkg, *version, len(files))
		return
	}
	if len(drift) == 0 {
		fmt.Printf("[syncdata] data/ already == %s@%s (%d files)\n", pkg, *version, len(files))
		return
	}
	// Replace wholesale: a stale v<hash>/ left next to the new one would be embedded too (go:embed v*).
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		fail(err)
	}
	for _, e := range entries {
		if e.Name() == ownFile {
			continue
		}
		if err := os.RemoveAll(filepath.Join(dataDir, e.Name())); err != nil {
			fail(err)
		}
	}
	for _, name := range sortedNames(files) {
		dst := filepath.Join(dataDir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			fail(err)
		}
		if err := os.WriteFile(dst, files[name], 0o644); err != nil {
			fail(err)
		}
	}
	fmt.Printf("[syncdata] data/ := %s@%s (%d files; %d changed)\n", pkg, *version, len(files), len(drift))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "[syncdata]", err)
	os.Exit(2)
}

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

var errNotFound = errors.New("not found")

func getJSON(url string, v any) error {
	res, err := client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("GET %s: HTTP 404: %w", url, errNotFound)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: HTTP %d", url, res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(v)
}

// publishedAtMost returns v when eve-fit-engine@v is on npm, otherwise the highest published X.Y.Z below
// it. The module may be released ahead of the TS twin (versions then re-align when the twin catches up);
// its data/ is the latest published one in the meantime.
func publishedAtMost(v string) (string, error) {
	var one struct {
		Version string `json:"version"`
	}
	err := getJSON(registry+pkg+"/"+v, &one)
	if err == nil {
		return v, nil
	}
	if !errors.Is(err, errNotFound) {
		return "", err
	}
	var all struct {
		Versions map[string]json.RawMessage `json:"versions"`
	}
	if err := getJSON(registry+pkg, &all); err != nil {
		return "", err
	}
	names := make([]string, 0, len(all.Versions))
	for k := range all.Versions {
		names = append(names, k)
	}
	best, ok := highestAtMost(names, v)
	if !ok {
		return "", fmt.Errorf("registry: no published %s version <= %s", pkg, v)
	}
	return best, nil
}

// parseVersion reads a plain X.Y.Z (prereleases and anything else are rejected).
func parseVersion(v string) ([3]int, bool) {
	var out [3]int
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || strconv.Itoa(n) != p {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

func versionLess(a, b [3]int) bool {
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i]
		}
	}
	return false
}

// highestAtMost picks the highest plain X.Y.Z in names that is <= limit.
func highestAtMost(names []string, limit string) (string, bool) {
	lim, ok := parseVersion(limit)
	if !ok {
		return "", false
	}
	best, found := "", false
	var bestV [3]int
	for _, n := range names {
		pv, ok := parseVersion(n)
		if !ok || versionLess(lim, pv) {
			continue
		}
		if !found || versionLess(bestV, pv) {
			best, bestV, found = n, pv, true
		}
	}
	return best, found
}

func latestVersion() (string, error) {
	var meta struct {
		Version string `json:"version"`
	}
	if err := getJSON(registry+pkg+"/latest", &meta); err != nil {
		return "", err
	}
	if meta.Version == "" {
		return "", errors.New("registry: no latest version")
	}
	return meta.Version, nil
}

// tarballData downloads eve-fit-engine@version, verifies its sha512 integrity and returns the files
// under package/data/ keyed by their slash path relative to data/.
func tarballData(version string) (map[string][]byte, error) {
	var meta struct {
		Dist struct {
			Tarball   string `json:"tarball"`
			Integrity string `json:"integrity"`
		} `json:"dist"`
	}
	if err := getJSON(registry+pkg+"/"+version, &meta); err != nil {
		return nil, err
	}
	want, ok := strings.CutPrefix(meta.Dist.Integrity, "sha512-")
	if !ok || meta.Dist.Tarball == "" {
		return nil, fmt.Errorf("registry: %s@%s has no tarball/sha512 integrity", pkg, version)
	}
	res, err := client.Get(meta.Dist.Tarball)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: HTTP %d", meta.Dist.Tarball, res.StatusCode)
	}
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	sum := sha512.Sum512(raw)
	if got := base64.StdEncoding.EncodeToString(sum[:]); got != want {
		return nil, fmt.Errorf("tarball integrity mismatch for %s@%s", pkg, version)
	}
	gz, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	files := map[string][]byte{}
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		name := path.Clean(h.Name)
		rel, ok := strings.CutPrefix(name, "package/data/")
		if !ok || h.Typeflag != tar.TypeReg {
			continue
		}
		if !fs.ValidPath(rel) || rel == ownFile {
			return nil, fmt.Errorf("unexpected tarball entry %q", h.Name)
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		files[rel] = b
	}
	if _, ok := files["manifest.json"]; !ok {
		return nil, fmt.Errorf("%s@%s: tarball carries no data/manifest.json", pkg, version)
	}
	return files, nil
}

func localData() (map[string][]byte, error) {
	files := map[string][]byte{}
	err := filepath.WalkDir(dataDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dataDir, p)
		rel = filepath.ToSlash(rel)
		if rel == ownFile {
			return nil
		}
		b, err := os.ReadFile(p)
		files[rel] = b
		return err
	})
	return files, err
}

func compare(want, have map[string][]byte) []string {
	var out []string
	for _, n := range sortedNames(want) {
		h, ok := have[n]
		switch {
		case !ok:
			out = append(out, "missing  "+n)
		case !bytes.Equal(h, want[n]):
			out = append(out, "differs  "+n)
		}
	}
	for _, n := range sortedNames(have) {
		if _, ok := want[n]; !ok {
			out = append(out, "extra    "+n)
		}
	}
	return out
}

func sortedNames(m map[string][]byte) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
