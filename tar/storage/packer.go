package storage

import (
	"bufio"
	"encoding/json"
	"io"
)

// Packer describes the methods to pack Entries to a storage destination
type Packer interface {
	// AddEntry packs the Entry and returns its position
	AddEntry(e Entry) (int, error)
}

// Unpacker describes the methods to read Entries from a source
type Unpacker interface {
	// Next returns the next Entry being unpacked, or error, until io.EOF
	Next() (*Entry, error)
}

/* TODO(vbatts) figure out a good model for this
type PackUnpacker interface {
	Packer
	Unpacker
}
*/

type jsonUnpacker struct {
	r     io.Reader
	b     *bufio.Reader
	isEOF bool
}

func (jup *jsonUnpacker) Next() (*Entry, error) {
	var e Entry
	if jup.isEOF {
		// since ReadBytes() will return read bytes AND an EOF, we handle it this
		// round-a-bout way so we can Unmarshal the tail with relevant errors, but
		// still get an io.EOF when the stream is ended.
		return nil, io.EOF
	}
	line, err := jup.b.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, err
	} else if err == io.EOF {
		jup.isEOF = true
	}
	err = json.Unmarshal(line, &e)
	if err != nil && jup.isEOF {
		// if the remainder actually _wasn't_ a remaining json structure, then just EOF
		return nil, io.EOF
	}
	return &e, err
}

// NewJsonUnpacker provides an Unpacker that reads Entries (SegmentType and
// FileType) as a json document.
//
// Each Entry read are expected to be delimited by new line.
func NewJsonUnpacker(r io.Reader) Unpacker {
	return &jsonUnpacker{
		r: r,
		b: bufio.NewReader(r),
	}
}

type jsonPacker struct {
	w   io.Writer
	e   *json.Encoder
	pos int
}

func (jp *jsonPacker) AddEntry(e Entry) (int, error) {
	e.Position = jp.pos
	err := jp.e.Encode(e)
	if err == nil {
		jp.pos++
	}
	return e.Position, err
}

// NewJsonPacker provides an Packer that writes each Entry (SegmentType and
// FileType) as a json document.
//
// The Entries are delimited by new line.
func NewJsonPacker(w io.Writer) Packer {
	return &jsonPacker{
		w: w,
		e: json.NewEncoder(w),
	}
}
