package asm

import (
	"io"

	"github.com/vbatts/tar-split/archive/tar"
	"github.com/vbatts/tar-split/tar/storage"
)

func NewInputTarStream(r io.Reader, fp FilePutter, p storage.Packer) (io.Reader, error) {
	// What to do here... folks will want their own access to the Reader that is
	// their tar archive stream, but we'll need that same stream to use our
	// forked 'archive/tar'.
	// Perhaps do an io.TeeReader that hand back an io.Reader for them to read
	// from, and we'll mitm the stream to store metadata.
	// We'll need a FilePutter too ...

	// Another concern, whether to do any FilePutter operations, such that we
	// don't extract any amount of the archive. But then again, we're not making
	// files/directories, hardlinks, etc. Just writing the io to the FilePutter.
	// Perhaps we have a DiscardFilePutter that is a bit bucket.

	// we'll return the pipe reader, since TeeReader does not buffer and will
	// only read what the outputRdr Read's. Since Tar archive's have padding on
	// the end, we want to be the one reading the padding, even if the user's
	// `archive/tar` doesn't care.
	pR, pW := io.Pipe()
	outputRdr := io.TeeReader(r, pW)

	tr := tar.NewReader(outputRdr)
	tr.RawAccounting = true

	return pR, nil
}
