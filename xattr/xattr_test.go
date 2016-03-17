package xattr

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestXattr(t *testing.T) {
	fh, err := ioutil.TempFile(".", "xattr.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fh.Name())
	if err := fh.Close(); err != nil {
		t.Fatal(err)
	}

	expected := []byte("1234")
	if err := Set(fh.Name(), "user.testing", expected); err != nil {
		t.Fatal(fh.Name(), err)
	}
	l, err := List(fh.Name())
	if err != nil {
		t.Error(fh.Name(), err)
	}
	if !(len(l) > 0) {
		t.Errorf("%q: expected a list of at least 1; got %d", len(l))
	}
	got, err := Get(fh.Name(), "user.testing")
	if err != nil {
		t.Fatal(fh.Name(), err)
	}
	if !bytes.Equal(got, expected) {
		t.Errorf("%q: expected %q; got %q", fh.Name(), expected, got)
	}
}
