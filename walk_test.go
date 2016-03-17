package mtree

import (
	"os"
	"testing"
)

func TestWalk(t *testing.T) {
	dh, err := Walk(".", nil, append(DefaultKeywords, "xattr"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = dh.WriteTo(os.Stdout); err != nil {
		t.Error(err)
	}
}
