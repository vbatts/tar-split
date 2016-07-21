package mtree

import "testing"

func TestVis(t *testing.T) {
	testset := []struct {
		Src, Dest string
	}{
		{"[", "\\133"},
		{" ", "\\040"},
		{"	", "\\011"},
		{"dir with space", "dir\\040with\\040space"},
		{"consec   spaces", "consec\\040\\040\\040spaces"},
		{"trailingsymbol[", "trailingsymbol\\133"},
		{" [ leadingsymbols", "\\040\\133\\040leadingsymbols"},
		{"no_need_for_encoding", "no_need_for_encoding"},
	}

	for i := range testset {
		got, err := Vis(testset[i].Src)
		if err != nil {
			t.Errorf("working with %q: %s", testset[i].Src, err)
		}
		if got != testset[i].Dest {
			t.Errorf("expected %#v; got %#v", testset[i].Dest, got)
			continue
		}

		got, err = Unvis(got)
		if err != nil {
			t.Errorf("working with %q: %s", testset[i].Src, err)
			continue
		}
		if got != testset[i].Src {
			t.Errorf("expected %#v; got %#v", testset[i].Src, got)
			continue
		}
	}
}

// The resulting string of Vis output could potentially be four times longer than
// the original. Vis must handle this possibility.
func TestVisLength(t *testing.T) {
	testString := "All work and no play makes Jack a dull boy\n"
	for i := 0; i < 20; i++ {
		Vis(testString)
		testString = testString + testString
	}
}
