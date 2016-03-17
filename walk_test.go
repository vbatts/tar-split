package mtree

import (
	"io/ioutil"
	"testing"
)

func TestWalk(t *testing.T) {
	dh, err := Walk(".", nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	fh, err := ioutil.TempFile("", "walk.")
	if err != nil {
		t.Fatal(err)
	}

	if _, err = dh.WriteTo(fh); err != nil {
		t.Error(err)
	}
	fh.Close()
	t.Fatal(fh.Name())
	//os.Remove(fh.Name())
}

func TestReadNames(t *testing.T) {
	names, err := readOrderedDirNames(".")
	if err != nil {
		t.Error(err)
	}
	t.Errorf("names: %q", names)
}
