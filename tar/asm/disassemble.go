package asm

import (
	"io"
	"io/ioutil"

	"github.com/vbatts/tar-split/archive/tar"
	"github.com/vbatts/tar-split/tar/storage"
)

// NewInputTarStream wraps the Reader stream of a tar archive and provides a
// Reader stream of the same.
//
// In the middle it will pack the segments and file metadata to storage.Packer
// `p`.
//
// The the FilePutter is where payload of files in the stream are stashed. If
// this stashing is not needed, fp can be nil or use NewDiscardFilePutter.
func NewInputTarStream(r io.Reader, p storage.Packer, fp FilePutter) (io.Reader, error) {
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

	if fp == nil {
		fp = NewDiscardFilePutter()
	}

	go func() {
		tr := tar.NewReader(outputRdr)
		tr.RawAccounting = true
		for {
			hdr, err := tr.Next()
			if err != nil {
				if err != io.EOF {
					pW.CloseWithError(err)
					return
				}
				// even when an EOF is reached, there is often 1024 null bytes on
				// the end of an archive. Collect them too.
				_, err := p.AddEntry(storage.Entry{
					Type:    storage.SegmentType,
					Payload: tr.RawBytes(),
				})
				if err != nil {
					pW.CloseWithError(err)
				} else {
					pW.Close()
				}
				return
			}

			if _, err := p.AddEntry(storage.Entry{
				Type:    storage.SegmentType,
				Payload: tr.RawBytes(),
			}); err != nil {
				pW.CloseWithError(err)
			}

			var csum []byte
			if hdr.Size > 0 {
				// if there is a file payload to write, then write the file to the FilePutter
				fileRdr, fileWrtr := io.Pipe()
				go func() {
					var err error
					csum, err = fp.Put(hdr.Name, fileRdr)
					if err != nil {
						pW.CloseWithError(err)
					}
				}()
				if _, err = io.Copy(fileWrtr, tr); err != nil {
					pW.CloseWithError(err)
					return
				}
			}
			// File entries added, regardless of size
			if _, err := p.AddEntry(storage.Entry{
				Type:    storage.FileType,
				Name:    hdr.Name,
				Size:    hdr.Size,
				Payload: csum,
			}); err != nil {
				pW.CloseWithError(err)
			}

			if _, err := p.AddEntry(storage.Entry{
				Type:    storage.SegmentType,
				Payload: tr.RawBytes(),
			}); err != nil {
				pW.CloseWithError(err)
			}
		}

		// it is allowable, and not uncommon that there is further padding on the
		// end of an archive, apart from the expected 1024 null bytes
		remainder, err := ioutil.ReadAll(outputRdr)
		if err != nil && err != io.EOF {
			pW.CloseWithError(err)
		}
		_, err = p.AddEntry(storage.Entry{
			Type:    storage.SegmentType,
			Payload: remainder,
		})
		if err != nil {
			pW.CloseWithError(err)
		} else {
			pW.Close()
		}
	}()

	return pR, nil
}
