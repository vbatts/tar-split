package mtree

import (
	"archive/tar"
	"fmt"
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

	go ts.readHeaders()
	return ts
}

type tarStream struct {
	root       *Entry
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
	ts.root = &Entry{
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
				// TODO: handle hardlinks
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
				ts.root.Set = &s
			} else {
				e.Set = &s
			}
		}
		err = populateTree(ts.root, &e, hdr)
		if err != nil {
			ts.setErr(err)
		}
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

// populateTree creates a pseudo file tree hierarchy using an Entry's Parent and
// Children fields. When examining the Entry e to insert in the tree, we
// determine if the path to that Entry exists yet. If it does, insert it in the
// appropriate position in the tree. If not, create a path up until the Entry's
// directory that it is contained in. Then, insert the Entry.
// root: the "." Entry
//    e: the Entry we are looking to insert
//  hdr: the tar header struct associated with e
func populateTree(root, e *Entry, hdr *tar.Header) error {
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
		return nil
	}
	// TODO: what about directory/file names with "/" in it?
	dirNames := strings.Split(wd, "/")
	parent := root
	for _, name := range dirNames[1:] {
		encoded, err := Vis(name)
		if err != nil {
			return err
		}
		if node := parent.Descend(encoded); node == nil {
			// Entry for directory doesn't exist in tree relative to root
			newEntry := Entry{
				Name:   encoded,
				Type:   RelativeType,
				Parent: parent,
			}
			parent.Children = append(parent.Children, &newEntry)
			parent = &newEntry
		} else {
			// Entry for directory exists in tree, just keep going
			parent = node
		}
	}
	if !isDir {
		parent.Children = append([]*Entry{e}, parent.Children...)
		e.Parent = parent
	} else {
		// the "placeholder" directory already exists in the Entry "parent",
		// so now we have to replace it's underlying data with that from e,
		// as well as set the Parent field. Note that we don't set parent = e
		// because parent is already in the pseudo tree, we just need to
		// complete it's data.
		e.Parent = parent.Parent
		*parent = *e
		commentpath, err := parent.Path()
		if err != nil {
			return err
		}
		parent.Prev = &Entry{
			Raw:  "# " + commentpath,
			Type: CommentType,
		}
	}
	return nil
}

// After constructing a pseudo file hierarchy tree, we want to "flatten" this
// tree by putting the Entries into a slice with appropriate positioning.
// root: the "head" of the sub-tree to flatten
// creator: a dhCreator that helps with the '/set' keyword
// keywords: keywords specified by the user that should be evaluated
func flatten(root *Entry, creator *dhCreator, keywords []string) {
	if root == nil {
		return
	}
	if root.Prev != nil {
		// root.Prev != nil implies root is a directory
		creator.DH.Entries = append(creator.DH.Entries,
			Entry{
				Type: BlankType,
				Pos:  len(creator.DH.Entries),
			})
		root.Prev.Pos = len(creator.DH.Entries)
		creator.DH.Entries = append(creator.DH.Entries, *root.Prev)

		// Check if we need a new set
		if creator.curSet == nil {
			creator.curSet = &Entry{
				Type:     SpecialType,
				Name:     "/set",
				Keywords: keywordSelector(append(tarDefaultSetKeywords, root.Set.Keywords...), keywords),
				Pos:      len(creator.DH.Entries),
			}
			creator.DH.Entries = append(creator.DH.Entries, *creator.curSet)
		} else {
			needNewSet := false
			for _, k := range root.Set.Keywords {
				if !inSlice(k, creator.curSet.Keywords) {
					needNewSet = true
					break
				}
			}
			if needNewSet {
				creator.curSet = &Entry{
					Name:     "/set",
					Type:     SpecialType,
					Pos:      len(creator.DH.Entries),
					Keywords: keywordSelector(append(tarDefaultSetKeywords, root.Set.Keywords...), keywords),
				}
				creator.DH.Entries = append(creator.DH.Entries, *creator.curSet)
			}
		}
	}
	root.Set = creator.curSet
	root.Keywords = setDifference(root.Keywords, creator.curSet.Keywords)
	root.Pos = len(creator.DH.Entries)
	creator.DH.Entries = append(creator.DH.Entries, *root)

	for _, c := range root.Children {
		flatten(c, creator, keywords)
	}

	if root.Prev != nil {
		// Show a comment when stepping out
		root.Prev.Pos = len(creator.DH.Entries)
		creator.DH.Entries = append(creator.DH.Entries, *root.Prev)
		dotEntry := Entry{
			Type: DotDotType,
			Name: "..",
			Pos:  len(creator.DH.Entries),
		}
		creator.DH.Entries = append(creator.DH.Entries, dotEntry)
	}
	return
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
					// prepend files
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
	if ts.root == nil {
		return nil, fmt.Errorf("root Entry not found. Nothing to flatten")
	}
	flatten(ts.root, &ts.creator, ts.keywords)
	return ts.creator.DH, nil
}
