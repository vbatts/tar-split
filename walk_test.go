package mtree

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestWalk(t *testing.T) {
	dh, err := Walk(".", nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}
	numEntries = countTypes(dh)

	fh, err := ioutil.TempFile("", "walk.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fh.Name())
	defer fh.Close()

	if _, err = dh.WriteTo(fh); err != nil {
		t.Fatal(err)
	}
	if _, err := fh.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	if dh, err = ParseSpec(fh); err != nil {
		t.Fatal(err)
	}
	for k, v := range countTypes(dh) {
		if numEntries[k] != v {
			t.Errorf("for type %s: expected %d, got %d", k, numEntries[k], v)
		}
	}
}
