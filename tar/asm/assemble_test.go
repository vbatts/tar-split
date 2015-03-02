package asm

import "testing"

func TestNewOutputTarStream(t *testing.T) {
	// TODO disassembly
	fgp := NewBufferFileGetPutter()
	_ = NewOutputTarStream(fgp, nil)
}
