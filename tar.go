package mtree

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Streamer creates a file hierarchy out of a tar stream
type Streamer interface {
	io.ReadCloser
	Hierarchy() (*DirectoryHierarchy, error)
}

var tarDefaultSetKeywords = []string{"type=file", "flags=none", "mode=0664"}

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
	rootComment := Entry{
		Raw:  "# .",
		Type: CommentType,
	}
	root := Entry{
		Name: ".",
		Type: RelativeType,
		Prev: &rootComment,
		Set: &Entry{
			Name: "meta-set",
			Type: SpecialType,
		},
	}
	metadataEntries := signatureEntries("<user specified tar archive>")
	for _, e := range metadataEntries {
		e.Pos = len(ts.creator.DH.Entries)
		ts.creator.DH.Entries = append(ts.creator.DH.Entries, e)
	}
	for {
		hdr, err := ts.tarReader.Next()
		if err != nil {
			flatten(&root, ts)
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
		encodedName, err := Vis(filepath.Base(hdr.Name))
		if err != nil {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			ts.pipeReader.CloseWithError(err)
			return
		}
		e := Entry{
			Name: encodedName,
			Type: RelativeType,
		}

		// now collect keywords on the file
		for _, keyword := range ts.keywords {
			if keyword == "time" {
				keyword = "tar_time"
			}
			if keyFunc, ok := KeywordFuncs[keyword]; ok {
				// We can't extract directories on to disk, so "size" keyword
				// is irrelevant for now
				if hdr.FileInfo().IsDir() && keyword == "size" {
					continue
				}

				if string(hdr.Typeflag) == string('1') {
					// TODO: get number of hardlinks for a file
				}
				val, err := keyFunc(hdr.Name, hdr.FileInfo(), tmpFile)
				if err != nil {
					ts.setErr(err)
				}
				// for good measure, check that we actually get a value for a keyword
				if val != "" {
					e.Keywords = append(e.Keywords, val)
				}

				// don't forget to reset the reader
				if _, err := tmpFile.Seek(0, 0); err != nil {
					tmpFile.Close()
					os.Remove(tmpFile.Name())
					ts.pipeReader.CloseWithError(err)
					return
				}
			}
		}
		// collect meta-set keywords for a directory so that we can build the
		// actual sets in `flatten`
		if hdr.FileInfo().IsDir() {
			s := Entry{
				Name: "meta-set",
				Type: SpecialType,
			}
			for _, setKW := range SetKeywords {
				if setKW == "time" {
					setKW = "tar_time"
				}
				if keyFunc, ok := KeywordFuncs[setKW]; ok {
					val, err := keyFunc(hdr.Name, hdr.FileInfo(), tmpFile)
					if err != nil {
						ts.setErr(err)
					}
					if val != "" {
						s.Keywords = append(s.Keywords, val)
					}
					if _, err := tmpFile.Seek(0, 0); err != nil {
						tmpFile.Close()
						os.Remove(tmpFile.Name())
						ts.pipeReader.CloseWithError(err)
					}
				}
			}
			if filepath.Dir(filepath.Clean(hdr.Name)) == "." {
				root.Set = &s
			} else {
				e.Set = &s
			}
		}
		populateTree(&root, &e, hdr, ts)
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}
}

type relationship int

const (
	unknownDir relationship = iota
	sameDir
	childDir
	parentDir
)

// populateTree creates a file tree hierarchy using an Entry's Parent and
// Children fields. When examining the Entry e to insert in the tree, we
// determine if the path to that Entry exists yet. If it does, insert it in the
// appropriate position in the tree. If not, create a path with "placeholder"
// directories, and then insert the Entry. populateTree does not consider
// symbolic links yet.
func populateTree(root, e *Entry, hdr *tar.Header, ts *tarStream) {
	isDir := hdr.FileInfo().IsDir()
	wd := filepath.Clean(hdr.Name)

	if !isDir {
		// If entry is a file, we only want the directory it's in.
		wd = filepath.Dir(wd)
	}
	if filepath.Dir(wd) == "." {
		if isDir {
			root.Keywords = e.Keywords
		} else {
			root.Children = append([]*Entry{e}, root.Children...)
			e.Parent = root
		}
		return
	}

	dirNames := strings.Split(wd, "/")
	parent := root
	for _, name := range dirNames[1:] {
		if node := parent.Descend(name); node == nil {
			// Entry for directory doesn't exist in tree relative to root
			var newEntry *Entry
			if isDir {
				newEntry = e
			} else {
				encodedName, err := Vis(name)
				if err != nil {
					ts.setErr(err)
					return
				}
				newEntry = &Entry{
					Name: encodedName,
					Type: RelativeType,
				}
			}
			newEntry.Parent = parent
			parent.Children = append(parent.Children, newEntry)
			parent = newEntry
		} else {
			// Entry for directory exists in tree, just keep going
			parent = node
		}
	}
	if !isDir {
		parent.Children = append([]*Entry{e}, parent.Children...)
		e.Parent = parent
	} else {
		commentpath, err := e.Path()
		if err != nil {
			ts.setErr(err)
			return
		}
		commentEntry := Entry{
			Raw:  "# " + commentpath,
			Type: CommentType,
		}
		e.Prev = &commentEntry
	}
}

// After constructing the tree from the tar stream, we want to "flatten" this
// tree by appending Entry's into ts.creator.DH.Entries in an appropriate
// manner to simplify writing the output with ts.creator.DH.WriteTo
// root: the "head" of the sub-tree to flatten
// ts  : tarStream to keep track of Entry's
func flatten(root *Entry, ts *tarStream) {
	if root.Prev != nil {
		// root.Prev != nil implies root is a directory
		ts.creator.DH.Entries = append(ts.creator.DH.Entries,
			Entry{
				Type: BlankType,
				Pos:  len(ts.creator.DH.Entries),
			})
		root.Prev.Pos = len(ts.creator.DH.Entries)
		ts.creator.DH.Entries = append(ts.creator.DH.Entries, *root.Prev)

		// Check if we need a new set
		if ts.creator.curSet == nil {
			ts.creator.curSet = &Entry{
				Type:     SpecialType,
				Name:     "/set",
				Keywords: keywordSelector(append(tarDefaultSetKeywords, root.Set.Keywords...), ts.keywords),
				Pos:      len(ts.creator.DH.Entries),
			}
			ts.creator.DH.Entries = append(ts.creator.DH.Entries, *ts.creator.curSet)
		} else {
			needNewSet := false
			for _, k := range root.Set.Keywords {
				if !inSlice(k, ts.creator.curSet.Keywords) {
					needNewSet = true
					break
				}
			}
			if needNewSet {
				ts.creator.curSet = &Entry{
					Name:     "/set",
					Type:     SpecialType,
					Pos:      len(ts.creator.DH.Entries),
					Keywords: keywordSelector(append(tarDefaultSetKeywords, root.Set.Keywords...), ts.keywords),
				}
				ts.creator.DH.Entries = append(ts.creator.DH.Entries, *ts.creator.curSet)
			}
		}
	}
	root.Set = ts.creator.curSet
	root.Keywords = setDifference(root.Keywords, ts.creator.curSet.Keywords)
	root.Pos = len(ts.creator.DH.Entries)
	ts.creator.DH.Entries = append(ts.creator.DH.Entries, *root)

	for _, c := range root.Children {
		flatten(c, ts)
	}

	if root.Prev != nil {
		// Show a comment when stepping out
		root.Prev.Pos = len(ts.creator.DH.Entries)
		ts.creator.DH.Entries = append(ts.creator.DH.Entries, *root.Prev)
		dotEntry := Entry{
			Type: DotDotType,
			Name: "..",
			Pos:  len(ts.creator.DH.Entries),
		}
		ts.creator.DH.Entries = append(ts.creator.DH.Entries, dotEntry)
	}
}

// filter takes in a pointer to an Entry, and returns a slice of Entry's that
// satisfy the predicate p
func filter(root *Entry, p func(*Entry) bool) []Entry {
	var validEntrys []Entry
	if len(root.Children) > 0 || root.Prev != nil {
		for _, c := range root.Children {
			// if an Entry is a directory, filter the directory
			if c.Prev != nil {
				validEntrys = append(validEntrys, filter(c, p)...)
			}
			if p(c) {
				if c.Prev == nil {
					// prepend directories
					validEntrys = append([]Entry{*c}, validEntrys...)
				} else {
					validEntrys = append(validEntrys, *c)
				}
			}
		}
		return validEntrys
	}
	return nil
}

func setDifference(this, that []string) []string {
	if len(this) == 0 {
		return that
	}
	diff := []string{}
	for _, kv := range this {
		if !inSlice(kv, that) {
			diff = append(diff, kv)
		}
	}
	return diff
}

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
