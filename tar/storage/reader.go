package storage

import (
	"io"

	"github.com/vbatts/tar-split/archive/tar"
)

func NewReader(r io.Reader, p Packer) *Reader {
	return &Reader{
		tr: tar.NewReader(r),
		p:  p,
	}
}

// Reader resembles the tar.Reader struct, and is handled the same. Though it
// takes an Packer which write the stored records and file info
type Reader struct {
	tr *tar.Reader
	p  Packer
}

func (r *Reader) Next() (*tar.Header, error) {
	// TODO read RawBytes
	return r.tr.Next()
}

func (r *Reader) Read(b []byte) (i int, e error) {
	return r.tr.Read(b)
}
