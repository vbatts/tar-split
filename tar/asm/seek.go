package asm

import "io"

// ReadSeekCloser implements Read(), Seek() and Close()
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}
