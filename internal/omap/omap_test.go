package omap

import (
	"reflect"
	"testing"
)

func TestMapOrderAndLastWins(t *testing.T) {
	m := New[string, int]()
	m.Set("b", 1)
	m.Set("a", 2)
	m.Set("b", 3) // sovrascrive senza spostare
	if !reflect.DeepEqual(m.Keys(), []string{"b", "a"}) {
		t.Fatalf("ordine: %v", m.Keys())
	}
	if v, _ := m.Get("b"); v != 3 {
		t.Fatalf("last wins: %d", v)
	}
	m.Delete("b")
	m.Set("b", 9) // dopo delete va in coda, come JS
	if !reflect.DeepEqual(m.Keys(), []string{"a", "b"}) {
		t.Fatalf("delete+set: %v", m.Keys())
	}
	if k, v := m.At(1); k != "b" || v != 9 || m.Len() != 2 || m.Has("z") {
		t.Fatal("At/Len/Has")
	}
	*m.Ptr("a") = 7
	if v, _ := m.Get("a"); v != 7 {
		t.Fatal("Ptr")
	}
}

func TestSet(t *testing.T) {
	s := NewSet(3, 1, 3, 2)
	if !reflect.DeepEqual(s.Values(), []int{3, 1, 2}) || s.Len() != 3 || !s.Has(1) || s.Has(9) {
		t.Fatalf("set: %v", s.Values())
	}
	s.Delete(1)
	if !reflect.DeepEqual(s.Values(), []int{3, 2}) {
		t.Fatal("delete")
	}
}
