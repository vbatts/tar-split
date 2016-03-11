package mtree

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// DirectoryHierarchy is the mapped structure for an mtree directory hierarchy
// spec
type DirectoryHierarchy struct {
	Entries []Entry
}

// WriteTo simplifies the output of the resulting hierarchy spec
func (dh DirectoryHierarchy) WriteTo(w io.Writer) (n int64, err error) {
	sort.Sort(byPos(dh.Entries))
	var sum int64
	for _, e := range dh.Entries {
		i, err := io.WriteString(w, e.String()+"\n")
		if err != nil {
			return sum, err
		}
		sum += int64(i)
	}
	return sum, nil
}

type byPos []Entry

func (bp byPos) Len() int           { return len(bp) }
func (bp byPos) Less(i, j int) bool { return bp[i].Pos < bp[j].Pos }
func (bp byPos) Swap(i, j int)      { bp[i], bp[j] = bp[j], bp[i] }

// Entry is each component of content in the mtree spec file
type Entry struct {
	Pos      int      // order in the spec
	Raw      string   // file or directory name
	Name     string   // file or directory name
	Keywords []string // TODO(vbatts) maybe a keyword typed set of values?
	Type     EntryType
}

func (e Entry) String() string {
	if e.Raw != "" {
		return e.Raw
	}
	// TODO(vbatts) if type is RelativeType and a keyword of not type=dir
	return fmt.Sprintf("%s %s", e.Name, strings.Join(e.Keywords, " "))
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
