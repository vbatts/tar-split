package mtree

import (
	"fmt"
	"path/filepath"
	"strings"
)

type byPos []Entry

func (bp byPos) Len() int           { return len(bp) }
func (bp byPos) Less(i, j int) bool { return bp[i].Pos < bp[j].Pos }
func (bp byPos) Swap(i, j int)      { bp[i], bp[j] = bp[j], bp[i] }

// Entry is each component of content in the mtree spec file
type Entry struct {
	Parent     *Entry   // up
	Children   []*Entry // down
	Prev, Next *Entry   // left, right
	Set        *Entry   // current `/set` for additional keywords
	Pos        int      // order in the spec
	Raw        string   // file or directory name
	Name       string   // file or directory name
	Keywords   []string // TODO(vbatts) maybe a keyword typed set of values?
	Type       EntryType
}

// Descend searches thru an Entry's children to find the Entry associated with
// `filename`. Directories are stored at the end of an Entry's children so do a
// traverse backwards. If you descend to a "."
func (e Entry) Descend(filename string) *Entry {
	if filename == "." || filename == "" {
		return &e
	}
	numChildren := len(e.Children)
	for i := range e.Children {
		c := e.Children[numChildren-1-i]
		if c.Name == filename {
			return c
		}
	}
	return nil
}

// Ascend gets the parent of an Entry. Serves mainly to maintain readability
// when traversing up and down an Entry tree
func (e Entry) Ascend() *Entry {
	return e.Parent
}

// Path provides the full path of the file, despite RelativeType or FullType
func (e Entry) Path() string {
	if e.Parent == nil || e.Type == FullType {
		return filepath.Clean(e.Name)
	}
	return filepath.Clean(filepath.Join(e.Parent.Path(), e.Name))
}

func (e Entry) String() string {
	if e.Raw != "" {
		return e.Raw
	}
	if e.Type == BlankType {
		return ""
	}
	if e.Type == DotDotType {
		return e.Name
	}
	if e.Type == SpecialType || e.Type == FullType || inSlice("type=dir", e.Keywords) {
		return fmt.Sprintf("%s %s", e.Name, strings.Join(e.Keywords, " "))
	}
	return fmt.Sprintf("    %s %s", e.Name, strings.Join(e.Keywords, " "))
}

// EntryType are the formats of lines in an mtree spec file
type EntryType int

// The types of lines to be found in an mtree spec file
const (
	SignatureType EntryType = iota // first line of the file, like `#mtree v2.0`
	BlankType                      // blank lines are ignored
	CommentType                    // Lines beginning with `#` are ignored
	SpecialType                    // line that has `/` prefix issue a "special" command (currently only /set and /unset)
	RelativeType                   // if the first white-space delimited word does not have a '/' in it. Options/keywords are applied.
	DotDotType                     // .. - A relative path step. keywords/options are ignored
	FullType                       // if the first word on the line has a `/` after the first character, it interpretted as a file pathname with options
)

// String returns the name of the EntryType
func (et EntryType) String() string {
	return typeNames[et]
}

var typeNames = map[EntryType]string{
	SignatureType: "SignatureType",
	BlankType:     "BlankType",
	CommentType:   "CommentType",
	SpecialType:   "SpecialType",
	RelativeType:  "RelativeType",
	DotDotType:    "DotDotType",
	FullType:      "FullType",
}
