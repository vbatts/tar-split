package asm

import (
	"testing"

	"github.com/vbatts/tar-split/tar/storage"
)

var entries = storage.Entries{
	storage.Entry{
		Type:     storage.SegmentType,
		Payload:  []byte("how"),
		Position: 0,
	},
	storage.Entry{
		Type:     storage.SegmentType,
		Payload:  []byte("y'all"),
		Position: 1,
	},
	storage.Entry{
		Type:     storage.FileType,
		Name:     "./hurr.txt",
		Payload:  []byte("deadbeef"),
		Size:     8,
		Position: 2,
	},
	storage.Entry{
		Type:     storage.SegmentType,
		Payload:  []byte("doin"),
		Position: 3,
	},
	storage.Entry{
		Type:     storage.FileType,
		Name:     "./ermahgerd.txt",
		Payload:  []byte("cafebabe"),
		Size:     8,
		Position: 4,
	},
}

func TestNewOutputTarStream(t *testing.T) {
	// TODO disassembly
	fgp := NewBufferFileGetPutter()
	_ = NewOutputTarStream(fgp, nil)
}

func TestNewInputTarStream(t *testing.T) {
}
