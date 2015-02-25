package asm

import (
	"io"

	"github.com/vbatts/tar-split/tar/storage"
)

func NewTarStream(fg FileGetter, up storage.Unpacker) io.ReadCloser {
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
				if entry.Size == 0 {
					continue
				}
				fh, err := fg.Get(entry.Name)
				if err != nil {
					pw.CloseWithError(err)
					break
				}
				defer fh.Close()
				if _, err := io.Copy(pw, fh); err != nil {
					pw.CloseWithError(err)
					break
				}
			}
		}
	}()
	return pr
}
