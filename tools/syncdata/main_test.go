package main

import "testing"

func TestHighestAtMost(t *testing.T) {
	names := []string{"0.2.25", "0.2.9", "0.2.27", "0.3.0-beta.1", "0.2.26", "junk"}
	cases := []struct {
		limit string
		want  string
		ok    bool
	}{
		{"0.2.26", "0.2.26", true},
		{"0.2.28", "0.2.27", true},
		{"0.2.10", "0.2.9", true},
		{"1.0.0", "0.2.27", true},
		{"0.1.0", "", false},
		{"nope", "", false},
	}
	for _, c := range cases {
		got, ok := highestAtMost(names, c.limit)
		if got != c.want || ok != c.ok {
			t.Errorf("highestAtMost(%q) = %q,%v; want %q,%v", c.limit, got, ok, c.want, c.ok)
		}
	}
}
