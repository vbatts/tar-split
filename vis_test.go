package mtree

import "testing"

func TestVis(t *testing.T) {
	testset := []struct {
		Src, Dest string
	}{
		{"[", "\\133"},
		{" ", "\\040"},
		{"	", "\\011"},
	}

	for i := range testset {
		got, err := Vis(testset[i].Src)
		if err != nil {
			t.Errorf("working with %q: %s", testset[i].Src, err)
			continue
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
