package mtree

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Streamer interface that wraps an io.ReadCloser with a function that will
// return it's Hierarchy
type Streamer interface {
	io.ReadCloser
	Hierarchy() (*DirectoryHierarchy, error)
}

// NewTarStreamer streams a tar archive and creates a file hierarchy based off
// of the tar metadata headers
func NewTarStreamer(r io.Reader, keywords []string) Streamer {
	pR, pW := io.Pipe()
	ts := &tarStream{
		pipeReader: pR,
		pipeWriter: pW,
		creator:    dhCreator{DH: &DirectoryHierarchy{}},
		teeReader:  io.TeeReader(r, pW),
		tarReader:  tar.NewReader(pR),
		keywords:   keywords,
	}
	go ts.readHeaders() // I don't like this
	return ts
}

type tarStream struct {
	creator    dhCreator
	pipeReader *io.PipeReader
	pipeWriter *io.PipeWriter
	teeReader  io.Reader
	tarReader  *tar.Reader
	keywords   []string
	err        error
}

func (ts *tarStream) readHeaders() {
	// We have to start with the directory we're in, and anything beyond these
	// items is determined at the time a tar is extracted.
	e := Entry{
		Name:     ".",
		Keywords: []string{"size=0", "type=dir"},
	}
	ts.creator.curDir = &e
	ts.creator.DH.Entries = append(ts.creator.DH.Entries, e)
	for {
		hdr, err := ts.tarReader.Next()
		if err != nil {
			ts.pipeReader.CloseWithError(err)
			return
		}

		// Because the content of the file may need to be read by several
		// KeywordFuncs, it needs to be an io.Seeker as well. So, just reading from
		// ts.tarReader is not enough.
		tmpFile, err := ioutil.TempFile("", "ts.payload.")
		if err != nil {
			ts.pipeReader.CloseWithError(err)
			return
		}
		// for good measure
		if err := tmpFile.Chmod(0600); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			ts.pipeReader.CloseWithError(err)
			return
		}
		if _, err := io.Copy(tmpFile, ts.tarReader); err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			ts.pipeReader.CloseWithError(err)
			return
		}

		// Alright, it's either file or directory
		e := Entry{
			Name: filepath.Base(hdr.Name),
			Pos:  len(ts.creator.DH.Entries),
			Type: RelativeType,
		}
		// now collect keywords on the file
		for _, keyword := range ts.keywords {
			if keyFunc, ok := KeywordFuncs[keyword]; ok {
				val, err := keyFunc(hdr.Name, hdr.FileInfo(), tmpFile)
				if err != nil {
					ts.setErr(err)
				}
				e.Keywords = append(e.Keywords, val)

				// don't forget to reset the reader
				if _, err := tmpFile.Seek(0, 0); err != nil {
					tmpFile.Close()
					os.Remove(tmpFile.Name())
					ts.pipeReader.CloseWithError(err)
					return
				}
			}
		}
		tmpFile.Close()
		os.Remove(tmpFile.Name())

		// compare directories, to determine parent of the current entry
		cd := compareDir(filepath.Dir(hdr.Name), ts.creator.curDir.Path())
		switch {
		case cd == sameDir:
			e.Parent = ts.creator.curDir
			if e.Parent != nil {
				e.Parent.Children = append(e.Parent.Children, &e)
			}
		case cd == parentDir:
			e.Parent = ts.creator.curDir.Parent
			if e.Parent != nil {
				e.Parent.Children = append(e.Parent.Children, &e)
			}
		}

		if hdr.FileInfo().IsDir() {
			ts.creator.curDir = &e
		}
		// TODO getting the parent child relationship of these entries!
		if hdr.FileInfo().IsDir() {
			log.Println(strings.Split(hdr.Name, "/"), strings.Split(ts.creator.curDir.Path(), "/"))
		}

		ts.creator.DH.Entries = append(ts.creator.DH.Entries, e)

		// Now is the wacky part of building out the entries. Since we can not
		// control how the archive was assembled, can only take in the order given.
		// Using `/set` will be tough. Hopefully i can do the directory stepping
		// with relative paths, but even then I may get a new directory, and not
		// the files first, but its directories first. :-\
	}
}

type relationship int

const (
	unknownDir relationship = iota
	sameDir
	childDir
	parentDir
)

func compareDir(curDir, prevDir string) relationship {
	curDir = filepath.Clean(curDir)
	prevDir = filepath.Clean(prevDir)
	if curDir == prevDir {
		return sameDir
	}
	if filepath.Dir(curDir) == prevDir {
		return childDir
	}
	if curDir == filepath.Dir(prevDir) {
		return parentDir
	}
	return unknownDir
}

func (ts *tarStream) setErr(err error) {
	ts.err = err
}

func (ts *tarStream) Read(p []byte) (n int, err error) {
	return ts.teeReader.Read(p)
}

func (ts *tarStream) Close() error {
	return ts.pipeReader.Close()
}

func (ts *tarStream) Hierarchy() (*DirectoryHierarchy, error) {
	if ts.err != nil && ts.err != io.EOF {
		return nil, ts.err
	}
	return ts.creator.DH, nil
}
