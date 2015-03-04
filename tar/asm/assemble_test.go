package asm

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/vbatts/tar-split/tar/storage"
)

var entries = []struct {
	Entry storage.Entry
	Body  []byte
}{
	{
		Entry: storage.Entry{
			Type:    storage.SegmentType,
			Payload: []byte("y'all"), // FIXME need real header here
		},
	},
	{
		Entry: storage.Entry{
			Type:    storage.FileType,
			Name:    "./hurr.txt",
			Payload: []byte{2, 116, 164, 177, 171, 236, 107, 78},
			Size:    20,
		},
		Body: []byte("imma hurr til I derp"),
	},
	{
		Entry: storage.Entry{
			Type:    storage.SegmentType,
			Payload: []byte("doin"), // FIXME need real header here
		},
	},
	{
		Entry: storage.Entry{
			Type:    storage.FileType,
			Name:    "./ermahgerd.txt",
			Payload: []byte{126, 72, 89, 239, 230, 252, 160, 187},
			Size:    26,
		},

		Body: []byte("café con leche, por favor"),
	},
	{
		Entry: storage.Entry{
			Type: storage.SegmentType,
			Payload: []byte{00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00, 00,
				00, 00},
		},
	},
}

func TestTarStream(t *testing.T) {
	fgp := NewBufferFileGetPutter()

	// first lets prep a GetPutter and Packer
	for i := range entries {
		if entries[i].Entry.Type == storage.FileType {
			j, csum, err := fgp.Put(entries[i].Entry.Name, bytes.NewBuffer(entries[i].Body))
			if err != nil {
				t.Error(err)
			}
			if j != entries[i].Entry.Size {
				t.Errorf("size %q: expected %d; got %d",
					entries[i].Entry.Name,
					entries[i].Entry.Size,
					j)
			}
			if !bytes.Equal(csum, entries[i].Entry.Payload) {
				t.Errorf("checksum %q: expected %v; got %v",
					entries[i].Entry.Name,
					entries[i].Entry.Payload,
					csum)
			}
		}
	}

	// next we'll use these to produce a tar stream.
	_ = NewOutputTarStream(fgp, nil)
	// TODO finish this
}

func TestInputTarStream(t *testing.T) {
	var (
		expectedSum        = "1eb237ff69bca6e22789ecb05b45d35ca307adbd"
		expectedSize int64 = 10240
	)

	fh, err := os.Open("./testdata/t.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	gzRdr, err := gzip.NewReader(fh)
	if err != nil {
		t.Fatal(err)
	}
	defer gzRdr.Close()

	// Setup where we'll store the metadata
	w := bytes.NewBuffer([]byte{})
	sp := storage.NewJsonPacker(w)
	fgp := NewBufferFileGetPutter()

	// check the tar on the front end too
	h0 := sha1.New()
	tRdr0 := io.TeeReader(gzRdr, h0)

	// wrap the disassembly stream
	tarStream, err := NewInputTarStream(tRdr0, sp, fgp)
	if err != nil {
		t.Fatal(err)
	}

	// get a sum of the stream after it has passed through to ensure it's the same.
	h1 := sha1.New()
	tRdr1 := io.TeeReader(tarStream, h1)

	// read it all to the bit bucket
	i, err := io.Copy(ioutil.Discard, tRdr1)
	if err != nil {
		t.Fatal(err)
	}

	if i != expectedSize {
		t.Errorf("size of tar: expected %d; got %d", expectedSize, i)
	}

	if fmt.Sprintf("%x", h1.Sum(nil)) != expectedSum {
		t.Logf("h0 was %x", h0.Sum(nil))
		t.Errorf("checksum of tar: expected %q; got %x", expectedSum, h1.Sum(nil))
	}

}
