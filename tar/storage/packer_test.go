package storage

import (
	"bytes"
	"io"
	"testing"
)

func TestJsonPackerUnpacker(t *testing.T) {
	e := []Entry{
		Entry{
			Type:    SegmentType,
			Payload: []byte("how"),
		},
		Entry{
			Type:    SegmentType,
			Payload: []byte("y'all"),
		},
		Entry{
			Type:    FileType,
			Name:    "./hurr.txt",
			Payload: []byte("deadbeef"),
		},
		Entry{
			Type:    SegmentType,
			Payload: []byte("doin"),
		},
	}

	buf := []byte{}
	b := bytes.NewBuffer(buf)

	func() {
		jp := NewJsonPacker(b)
		for i := range e {
			if _, err := jp.AddEntry(e[i]); err != nil {
				t.Error(err)
			}
		}
	}()

	t.Errorf("%#v", b.String())
	b = bytes.NewBuffer(b.Bytes())
	func() {
		jup := NewJsonUnpacker(b)
		for {
			entry, err := jup.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Error(err)
			}
			t.Errorf("%#v", entry)
		}
	}()

}
