package asm

import (
	"io"
	"os"
	"path"
)

type FileGetter interface {
	// Get returns a stream for the provided file path
	Get(string) (io.ReadCloser, error)
}

// NewPathFileGetter returns a FileGetter that is for files relative to path relpath.
func NewPathFileGetter(relpath string) FileGetter {
	return &pathFileGetter{root: relpath}
}

type pathFileGetter struct {
	root string
}

func (pfg pathFileGetter) Get(filename string) (io.ReadCloser, error) {
	// FIXME might should have a check for '../../../../etc/passwd' attempts?
	return os.Open(path.Join(pfg.root, filename))
}
