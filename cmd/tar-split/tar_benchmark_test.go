package main

import (
	"io"
	"os"
	"testing"

	upTar "archive/tar"

	ourTar "github.com/vbatts/tar-split/internal/archive/tar"
)

var testfile = "../../archive/tar/testdata/sparse-formats.tar"

func BenchmarkUpstreamTar(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fh, err := os.Open(testfile)
		if err != nil {
			b.Fatal(err)
		}
		tr := upTar.NewReader(fh)
		for {
			_, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				fh.Close()
				b.Fatal(err)
			}
			_, err = io.Copy(io.Discard, tr)
			if err != nil {
				b.Fatal(err)
			}
		}
		if err := fh.Close(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOurTarNoAccounting(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fh, err := os.Open(testfile)
		if err != nil {
			b.Fatal(err)
		}
		tr := ourTar.NewReader(fh)
		tr.RawAccounting = false // this is default, but explicit here
		for {
			_, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				fh.Close()
				b.Fatal(err)
			}
			_, err = io.Copy(io.Discard, tr)
			if err != nil {
				b.Fatal(err)
			}
		}
		if err := fh.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
func BenchmarkOurTarYesAccounting(b *testing.B) {
	for n := 0; n < b.N; n++ {
		fh, err := os.Open(testfile)
		if err != nil {
			b.Fatal(err)
		}
		tr := ourTar.NewReader(fh)
		tr.RawAccounting = true // This enables mechanics for collecting raw bytes
		for {
			_ = tr.RawBytes()
			_, err := tr.Next()
			_ = tr.RawBytes()
			if err != nil {
				if err == io.EOF {
					break
				}
				fh.Close()
				b.Fatal(err)
			}
			_, err = io.Copy(io.Discard, tr)
			if err != nil {
				b.Fatal(err)
			}
			_ = tr.RawBytes()
		}
		if err := fh.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
