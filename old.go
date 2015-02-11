// +build ignore

package main

import (
	"bytes"
	"crypto"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"strconv"
)

func main() {
	flag.Parse()

	tarInfos := []TarInfo{}
	for _, arg := range flag.Args() {
		func() {
			// Open the tar archive
			fh, err := os.Open(arg)
			if err != nil {
				log.Fatal(err)
			}
			defer fh.Close()

			fi, err := fh.Stat()
			if err != nil {
				log.Fatal(err)
			}

			ti := TarInfo{
				Name: arg,
				Size: fi.Size(),
			}

			for {
				buf, err := readHeader(fh)
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Fatal(err)
				}

				if !verifyChecksum(buf) {
					log.Fatal(ErrHeader)
				}

				s := slicer(buf)
				name := cString(s.next(100))
				s.next(8) // mode
				s.next(8) // uid
				s.next(8) // gid
				size, err := octal(s.next(12))
				if err != nil {
					log.Fatal(err)
				}
				e := Entry{
					Header: buf,
					Name:   name,
					Size:   size,
				}
				log.Printf("%#v", e)
				ti.Entries = append(ti.Entries, e)

				// TODO(vbatts) some pax types need further reading, for their headers ...
				// XXX this where it is broken
				if _, err := fh.Seek(size, 1); err != nil {
					log.Fatal(err)
				}
			}

			tarInfos = append(tarInfos, ti)
		}()
	}
	if *flOutputJson != "" {
		fh, err := os.Create(*flOutputJson)
		if err != nil {
			log.Fatal(err)
		}
		defer fh.Close()
		jsonBuf, err := json.Marshal(tarInfos)
		if err != nil {
			log.Fatal(err)
		}
		_, err = fh.Write(jsonBuf)
		if err != nil {
			log.Fatal(err)
		}
	}
}

const BlockSize = 512

var (
	zeroBlock = make([]byte, BlockSize)
	hdrBuff   = make([]byte, BlockSize)
	ErrHeader = errors.New("archive/tar: invalid tar header")

	flOutputJson = flag.String("o", "", "output json of the tar archives")
)

// cString parses bytes as a NUL-terminated C-style string.
// If a NUL byte is not found then the whole slice is returned as a string.
//
// copied from 'archive/tar/reader.go'
func cString(b []byte) string {
	n := 0
	for n < len(b) && b[n] != 0 {
		n++
	}
	return string(b[0:n])
}

// parse the octal value from the byte array
//
// copied from 'archive/tar/reader.go'
func octal(b []byte) (int64, error) {
	// Check for binary format first.
	if len(b) > 0 && b[0]&0x80 != 0 {
		var x int64
		for i, c := range b {
			if i == 0 {
				c &= 0x7f // ignore signal bit in first byte
			}
			x = x<<8 | int64(c)
		}
		return x, nil
	}

	// Because unused fields are filled with NULs, we need
	// to skip leading NULs. Fields may also be padded with
	// spaces or NULs.
	// So we remove leading and trailing NULs and spaces to
	// be sure.
	b = bytes.Trim(b, " \x00")

	if len(b) == 0 {
		return 0, nil
	}
	x, err := strconv.ParseUint(cString(b), 8, 64)
	return int64(x), err
}

// copied from 'archive/tar/reader.go'
func verifyChecksum(header []byte) bool {
	given, err := octal(header[148:156])
	if err != nil {
		return false
	}
	unsigned, signed := checksum(header)
	return given == unsigned || given == signed
}

// POSIX specifies a sum of the unsigned byte values, but the Sun tar uses signed byte values.
// We compute and return both.
//
// copied from 'archive/tar/reader.go'
func checksum(header []byte) (unsigned int64, signed int64) {
	for i := 0; i < len(header); i++ {
		if i == 148 {
			// The chksum field (header[148:156]) is special: it should be treated as space bytes.
			unsigned += ' ' * 8
			signed += ' ' * 8
			i += 7
			continue
		}
		unsigned += int64(header[i])
		signed += int64(int8(header[i]))
	}
	return
}

// copied from 'archive/tar/reader.go'
type slicer []byte

// copied from 'archive/tar/reader.go'
func (sp *slicer) next(n int) (b []byte) {
	s := *sp
	b, *sp = s[0:n], s[n:]
	return
}

// readHeader looks for the first header segement from the provided reader
//
// partially copied from 'archive/tar/reader.go'
func readHeader(r io.Reader) ([]byte, error) {
	// prep our buffer
	buf := hdrBuff[:]
	copy(buf, zeroBlock)

	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}

	// Two blocks of zero bytes marks the end of the archive.
	if bytes.Equal(buf, zeroBlock[0:BlockSize]) {
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, err
		}
		if bytes.Equal(buf, zeroBlock[0:BlockSize]) {
			return nil, io.EOF
		}
		return nil, ErrHeader // zero block and then non-zero block
	}

	return buf, nil
}

type (
	// for a whole tar archive
	TarInfo struct {
		Name    string
		Size    int64
		Entries []Entry

		// TODO(vbatts) would be nice to satisfy the Reader interface, so that this could be passed directly to tar.Reader
	}

	// each file from the tar archive has it's header copied exactly,
	//and payload of it's file Checksummed if the file size is greater than 0
	Entry struct {
		Pos      int64
		Name     string
		Header   []byte
		Size     int64
		Checksum []byte
		Hash     crypto.Hash

		// TODO(vbatts) perhaps have info to find the file on disk, to provide an io.Reader
	}
)
