package mtree

import "io"

// Positioner responds with the newline delimited position
type Positioner interface {
	Pos() int
}

// DirectoryHierarchy is the mapped structure for an mtree directory hierarchy
// spec
type DirectoryHierarchy struct {
	Comments []Comment
	Entries  []Entry
}

// WriteTo simplifies the output of the resulting hierarchy spec
func (dh DirectoryHierarchy) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

// Comment stores arbitrary metadata for the spec. Sometimes "user", "machine",
// "tree", and "date". But most of the time, it includes the relative path of
// the directory being stepped into. Or a "signature" like `#mtree v2.0`,
type Comment struct {
	Position int
	Str      string
	// TODO(vbatts) include a comment line parser
}

// Pos returns the line of this comment
func (c Comment) Pos() int {
	return c.Position
}

// Entry is each component of content in the mtree spec file
type Entry struct {
	Position int      // order in the spec
	Name     string   // file or directory name
	Keywords []string // TODO(vbatts) maybe a keyword typed set of values?
	str      string   // raw string. needed?
	Type     EntryType
}

// Pos returns the line of this comment
func (e Entry) Pos() int {
	return e.Position
}

type EntryType int

const (
	SpecialType   int = iota // line that has `/` prefix issue a "special" command (currently only /set and /unset)
	FileType                 // indented line
	DirectoryType            // ^ matched line, that is not /set
	PathStepType             // .. - keywords/options are ignored
	FullType                 // if the first word on the line has a `/` after the first character, it interpretted as a file pathname with options
)
