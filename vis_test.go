package mtree

import "testing"

func TestVisBasic(t *testing.T) {
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
		got, err := Vis(testset[i].Src, DefaultVisFlags)
		if err != nil {
			t.Errorf("working with %q: %s", testset[i].Src, err)
		}
		if got != testset[i].Dest {
			t.Errorf("%q: expected %#v; got %#v", testset[i].Src, testset[i].Dest, got)
			continue
		}

		got, err = Unvis(got)
		if err != nil {
			t.Errorf("working with %q: %s: %q", testset[i].Src, err, got)
			continue
		}
		if got != testset[i].Src {
			t.Errorf("%q: expected %#v; got %#v", testset[i].Dest, testset[i].Src, got)
			continue
		}
	}
}
