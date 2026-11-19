package engine

import "github.com/capsuleers-app/capsuleers.fitting-engine-go/internal/omap"

// intSet: insieme di interi con ordine di inserimento (Set di JS con soli numeri).
type intSet = omap.Set[int]

func newIntSet(items []int) *intSet { return omap.NewSet(items...) }

// intMap: mappa int -> int con ordine di inserimento, "last wins" come `new Map([[k, v], ...])`.
type intMap struct{ m *omap.Map[int, int] }

func newIntMap(pairs [][2]int) *intMap {
	m := omap.New[int, int]()
	for _, p := range pairs {
		m.Set(p[0], p[1])
	}
	return &intMap{m: m}
}

func (t *intMap) get(k int) (int, bool) { return t.m.Get(k) }
