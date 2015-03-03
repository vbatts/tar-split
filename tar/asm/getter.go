package asm

import (
	"bytes"
	"errors"
	"hash/crc64"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type FileGetter interface {
	// Get returns a stream for the provided file path
	Get(string) (io.ReadCloser, error)
}

type FilePutter interface {
	// Put returns a stream for the provided file path
	Put(string, io.Reader) ([]byte, error)
}

type FileGetPutter interface {
	FileGetter
	FilePutter
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

type bufferFileGetPutter struct {
	files map[string][]byte
}

func (bfgp bufferFileGetPutter) Get(name string) (io.ReadCloser, error) {
	if _, ok := bfgp.files[name]; !ok {
		return nil, errors.New("no such file")
	}
	b := bytes.NewBuffer(bfgp.files[name])
	return &readCloserWrapper{b}, nil
}

func (bfgp *bufferFileGetPutter) Put(name string, r io.Reader) ([]byte, error) {
	c := crc64.New(crcTable)
	tRdr := io.TeeReader(r, c)
	b := bytes.NewBuffer([]byte{})
	if _, err := io.Copy(b, tRdr); err != nil {
		return nil, err
	}
	bfgp.files[name] = b.Bytes()
	return c.Sum(nil), nil
}

type readCloserWrapper struct {
	io.Reader
}

func (w *readCloserWrapper) Close() error { return nil }

// NewBufferFileGetPutter is simple in memory FileGetPutter
//
// Implication is this is memory intensive...
// Probably best for testing or light weight cases.
func NewBufferFileGetPutter() FileGetPutter {
	return &bufferFileGetPutter{
		files: map[string][]byte{},
	}
}

// NewDiscardFilePutter is a bit bucket FilePutter
func NewDiscardFilePutter() FilePutter {
	return &bitBucketFilePutter{}
}

type bitBucketFilePutter struct {
}

func (bbfp *bitBucketFilePutter) Put(name string, r io.Reader) ([]byte, error) {
	c := crc64.New(crcTable)
	tRdr := io.TeeReader(r, c)
	_, err := io.Copy(ioutil.Discard, tRdr)
	return c.Sum(nil), err
}

var crcTable = crc64.MakeTable(crc64.ISO)
