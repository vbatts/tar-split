package storage

import (
	"bufio"
	"encoding/json"
	"io"
)

type Packer interface {
	// AddSegment packs the segment bytes provided and returns the position of
	// the entry
	//AddSegment([]byte) (int, error)
	// AddFile packs the File provided and returns the position of the entry. The
	// Position is set in the stored File.
	//AddFile(File) (int, error)

	//
	AddEntry(e Entry) (int, error)
}

type Unpacker interface {
	Next() (*Entry, error)
}

type PackUnpacker interface {
	Packer
	Unpacker
}

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

// jsonUnpacker writes each entry (SegmentType and FileType) as a json document.
// Each entry on a new line.
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

func NewJsonPacker(w io.Writer) Packer {
	return &jsonPacker{
		w: w,
		e: json.NewEncoder(w),
	}
}
