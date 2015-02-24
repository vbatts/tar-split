package asm

import (
	"io"
	"os"
	"path"

	"github.com/vbatts/tar-split/tar/storage"
)

func NewTarStream(relpath string, up storage.Unpacker) io.ReadCloser {
	pr, pw := io.Pipe()
	go func() {
		for {
			entry, err := up.Next()
			if err != nil {
				pw.CloseWithError(err)
				break
			}
			switch entry.Type {
			case storage.SegmentType:
				if _, err := pw.Write(entry.Payload); err != nil {
					pw.CloseWithError(err)
					break
				}
			case storage.FileType:
				if err := writeEntryFromRelPath(pw, relpath, entry); err != nil {
					pw.CloseWithError(err)
					break
				}
			}
		}
	}()
	return pr
}

func writeEntryFromRelPath(w io.Writer, root string, entry *storage.Entry) error {
	if entry.Size == 0 {
		return nil
	}

	// FIXME might should have a check for '../../../../etc/passwd' attempts?
	fh, err := os.Open(path.Join(root, entry.Name))
	if err != nil {
		return err
	}
	defer fh.Close()
	if _, err := io.Copy(w, fh); err != nil {
		return err
	}

	return nil
}
