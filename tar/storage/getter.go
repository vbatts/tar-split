package storage

import (
	"bytes"
	"errors"
	"hash/crc64"
	"io"
	"io/ioutil"
	"os"
	"path"
)

// FileGetter is the interface for getting a stream of a file payload,
// addressed by name/filepath. Presumably, the names will be scoped to relative
// file paths.
type FileGetter interface {
	// Get returns a stream for the provided file path
	Get(filepath string) (output io.ReadCloser, err error)
}

// FilePutter is the interface for storing a stream of a file payload,
// addressed by name/filepath.
type FilePutter interface {
	// Put returns the size of the stream received, and the crc64 checksum for
	// the provided stream
	Put(filepath string, input io.Reader) (size int64, checksum []byte, err error)
}

// FileGetPutter is the interface that groups both Getting and Putting file
// payloads.
type FileGetPutter interface {
	FileGetter
	FilePutter
}

// NewPathFileGetter returns a FileGetter that is for files relative to path
// relpath.
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

func (bfgp *bufferFileGetPutter) Put(name string, r io.Reader) (int64, []byte, error) {
	c := crc64.New(CRCTable)
	tRdr := io.TeeReader(r, c)
	b := bytes.NewBuffer([]byte{})
	i, err := io.Copy(b, tRdr)
	if err != nil {
		return 0, nil, err
	}
	bfgp.files[name] = b.Bytes()
	return i, c.Sum(nil), nil
}

type readCloserWrapper struct {
	io.Reader
}

func (w *readCloserWrapper) Close() error { return nil }

// NewBufferFileGetPutter is a simple in-memory FileGetPutter
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

func (bbfp *bitBucketFilePutter) Put(name string, r io.Reader) (int64, []byte, error) {
	c := crc64.New(CRCTable)
	tRdr := io.TeeReader(r, c)
	i, err := io.Copy(ioutil.Discard, tRdr)
	return i, c.Sum(nil), err
}

// CRCTable is the default table used for crc64 sum calculations
var CRCTable = crc64.MakeTable(crc64.ISO)
