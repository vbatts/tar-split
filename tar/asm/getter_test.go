package asm

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestGetter(t *testing.T) {
	fgp := NewBufferFileGetPutter()
	files := map[string][]byte{
		"file1.txt": []byte("foo"),
		"file2.txt": []byte("bar"),
	}
	for n, b := range files {
		if err := fgp.Put(n, bytes.NewBuffer(b)); err != nil {
			t.Error(err)
		}
	}
	for n, b := range files {
		r, err := fgp.Get(n)
		if err != nil {
			t.Error(err)
		}
		buf, err := ioutil.ReadAll(r)
		if err != nil {
			t.Error(err)
		}
		if string(b) != string(buf) {
			t.Errorf("expected %q, got %q", string(b), string(buf))
		}
	}
}
