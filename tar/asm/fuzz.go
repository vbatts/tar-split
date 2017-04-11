// +build gofuzz

package asm

import (
	"bytes"
	"log"

	"github.com/vbatts/tar-split/archive/tar"
	"github.com/vbatts/tar-split/tar/storage"
)

func Fuzz(data []byte) int {
	sp := storage.NewJSONPacker(bytes.NewBuffer([]byte{}))
	fgp := storage.NewBufferFileGetPutter()
	tarStream, err := NewInputTarStream(bytes.NewReader(data), sp, fgp)
	if err != nil {
		if tarStream != nil {
			panic("tarStream is != nil on error")
		}
		log.Println(err)
		return 0
	}
	rdr := tar.NewReader(tarStream)

	for {
		hdr, err := rdr.Next()
		if err != nil {
			if hdr != nil {
				panic("hdr is != nil on error")
			}
			log.Println(err)
			return 0
		}
		log.Printf("%v", hdr)
	}
	return 1
}
