package main

import (
	"crypto"
	"flag"
	"io"
	"log"
	"os"
)

func main() {
	flag.Parse()

	hdrBuff := make([]byte, BlockSize)

	for _, arg := range flag.Args() {
		func() {
			// Open the tar archive
			fh, err := os.Open(arg)
			if err != nil {
				log.Fatal(err)
			}
			defer fh.Close()

			// prep our buffer
			buf := hdrBuff[:]
			copy(buf, zeroBlock)

			if _, err := io.ReadFull(fh, buf); err != nil {
				log.Fatal(err)
			}
		}()
	}
}

const BlockSize = 512

var (
	zeroBlock = make([]byte, BlockSize)

	flOutputJson = flag.String("o", "", "output json of the tar archives")
)

type (
	// for a whole tar archive
	TarInfo struct {
		Name    string
		Entries []Entry

		// TODO(vbatts) would be nice to satisfy the Reader interface, so that this could be passed directly to tar.Reader
	}

	// each file from the tar archive has it's header copied exactly,
	//and payload of it's file Checksummed if the file size is greater than 0
	Entry struct {
		Pos      int64
		Header   []byte
		Size     int64
		Checksum []byte
		Hash     crypto.Hash

		// TODO(vbatts) perhaps have info to find the file on disk, to provide an io.Reader
	}
)
