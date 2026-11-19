// Package omap fornisce Map e Set con ordine di inserimento, sostituti di Map/Set
// di JS (le map di Go hanno ordine casuale: nei package dell'engine e' vietato
// il range su map, lo verifica tools/lint). Semantica JS: Set su una chiave gia'
// presente sovrascrive il valore e NON ne cambia la posizione ("last wins");
// Delete seguito da Set rimette la chiave in coda.
//
// Le chiavi sono comparable: per float64 NaN non e' supportato (in JS SameValueZero
// lo unirebbe a se stesso, in Go NaN != NaN); l'engine usa chiavi int e string.
package omap

// Map e' una mappa con ordine di inserimento.
type Map[K comparable, V any] struct {
	keys []K
	vals []V
	idx  map[K]int
}

// New crea una Map vuota.
func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{idx: map[K]int{}}
}

// Len e' il numero di voci.
func (m *Map[K, V]) Len() int { return len(m.keys) }

// Set inserisce o sovrascrive (la posizione di una chiave esistente non cambia).
func (m *Map[K, V]) Set(k K, v V) {
	if i, ok := m.idx[k]; ok {
		m.vals[i] = v
		return
	}
	m.idx[k] = len(m.keys)
	m.keys = append(m.keys, k)
	m.vals = append(m.vals, v)
}

// Get legge il valore.
func (m *Map[K, V]) Get(k K) (V, bool) {
	if i, ok := m.idx[k]; ok {
		return m.vals[i], true
	}
	var zero V
	return zero, false
}

// Ptr restituisce un puntatore al valore (valido fino al prossimo Set di una chiave nuova o Delete).
func (m *Map[K, V]) Ptr(k K) *V {
	if i, ok := m.idx[k]; ok {
		return &m.vals[i]
	}
	return nil
}

// Has dice se la chiave e' presente.
func (m *Map[K, V]) Has(k K) bool { _, ok := m.idx[k]; return ok }

// Delete rimuove la chiave (O(n): usata raramente). Restituisce se era presente.
func (m *Map[K, V]) Delete(k K) bool {
	i, ok := m.idx[k]
	if !ok {
		return false
	}
	m.keys = append(m.keys[:i], m.keys[i+1:]...)
	m.vals = append(m.vals[:i], m.vals[i+1:]...)
	delete(m.idx, k)
	for j := i; j < len(m.keys); j++ {
		m.idx[m.keys[j]] = j
	}
	return true
}

// At restituisce la i-esima voce in ordine di inserimento. Si itera con
// `for i := 0; i < m.Len(); i++ { k, v := m.At(i) }`.
func (m *Map[K, V]) At(i int) (K, V) { return m.keys[i], m.vals[i] }

// Keys restituisce una copia delle chiavi in ordine di inserimento.
func (m *Map[K, V]) Keys() []K { return append([]K(nil), m.keys...) }

// Clear svuota la mappa.
func (m *Map[K, V]) Clear() {
	m.keys, m.vals = nil, nil
	m.idx = map[K]int{}
}

// Set e' un insieme con ordine di inserimento.
type Set[K comparable] struct{ m *Map[K, struct{}] }

// NewSet crea un Set con gli elementi dati (nell'ordine, duplicati ignorati).
func NewSet[K comparable](items ...K) *Set[K] {
	s := &Set[K]{m: New[K, struct{}]()}
	for _, it := range items {
		s.Add(it)
	}
	return s
}

// Add aggiunge l'elemento (no-op se presente: la posizione resta quella originale).
func (s *Set[K]) Add(k K) { s.m.Set(k, struct{}{}) }

// Has dice se l'elemento e' presente.
func (s *Set[K]) Has(k K) bool { return s.m.Has(k) }

// Delete rimuove l'elemento.
func (s *Set[K]) Delete(k K) bool { return s.m.Delete(k) }

// Len e' il numero di elementi.
func (s *Set[K]) Len() int { return s.m.Len() }

// At restituisce l'i-esimo elemento in ordine di inserimento.
func (s *Set[K]) At(i int) K { k, _ := s.m.At(i); return k }

// Values restituisce una copia degli elementi in ordine di inserimento.
func (s *Set[K]) Values() []K { return s.m.Keys() }

// EntriesAny restituisce le voci come coppie {chiave, valore} in ordine di inserimento (per i dump di verifica).
func (m *Map[K, V]) EntriesAny() [][2]any {
	out := make([][2]any, len(m.keys))
	for i := range m.keys {
		out[i] = [2]any{m.keys[i], m.vals[i]}
	}
	return out
}
