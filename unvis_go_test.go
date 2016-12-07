package mtree

import "testing"

type runeCheck func(rune) bool

func TestUnvisHelpers(t *testing.T) {
	testset := []struct {
		R      rune
		Check  runeCheck
		Expect bool
	}{
		{'a', ishex, true},
		{'A', ishex, true},
		{'z', ishex, false},
		{'Z', ishex, false},
		{'G', ishex, false},
		{'1', ishex, true},
		{'0', ishex, true},
		{'9', ishex, true},
		{'0', isoctal, true},
		{'3', isoctal, true},
		{'7', isoctal, true},
		{'9', isoctal, false},
		{'a', isoctal, false},
		{'z', isoctal, false},
		{'3', isalnum, true},
		{'a', isalnum, true},
		{';', isalnum, false},
		{'!', isalnum, false},
		{' ', isalnum, false},
		{'3', isgraph, true},
		{'a', isgraph, true},
		{';', isgraph, true},
		{'!', isgraph, true},
		{' ', isgraph, false},
	}

	for i, ts := range testset {
		got := ts.Check(ts.R)
		if got != ts.Expect {
			t.Errorf("%d: %q expected: %t; got %t", i, string(ts.R), ts.Expect, got)
		}
	}
}
