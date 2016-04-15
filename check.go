package mtree

import (
	"fmt"
	"os"
	"sort"
)

// Result of a Check
type Result struct {
	// list of any failures in the Check
	Failures []Failure `json:"failures"`
}

// Failure of a particular keyword for a path
type Failure struct {
	Path     string `json:"path"`
	Keyword  string `json:"keyword"`
	Expected string `json:"expected"`
	Got      string `json:"got"`
}

// String returns a "pretty" formatting for a Failure
func (f Failure) String() string {
	return fmt.Sprintf("%q: keyword %q: expected %s; got %s", f.Path, f.Keyword, f.Expected, f.Got)
}

// Check a root directory path against the DirectoryHierarchy, regarding only
// the available keywords from the list and each entry in the hierarchy.
// If keywords is nil, the check all present in the DirectoryHierarchy
func Check(root string, dh *DirectoryHierarchy, keywords []string) (*Result, error) {
	creator := dhCreator{DH: dh}
	curDir, err := os.Getwd()
	if err == nil {
		defer os.Chdir(curDir)
	}

	if err := os.Chdir(root); err != nil {
		return nil, err
	}
	sort.Sort(byPos(creator.DH.Entries))

	var result Result
	for i, e := range creator.DH.Entries {
		switch e.Type {
		case SpecialType:
			if e.Name == "/set" {
				creator.curSet = &creator.DH.Entries[i]
			} else if e.Name == "/unset" {
				creator.curSet = nil
			}
		case RelativeType, FullType:
			filename := e.Path()
			info, err := os.Lstat(filename)
			if err != nil {
				return nil, err
			}

			var kvs KeyVals
			if creator.curSet != nil {
				kvs = MergeSet(creator.curSet.Keywords, e.Keywords)
			} else {
				kvs = NewKeyVals(e.Keywords)
			}

			for _, kv := range kvs {
				keywordFunc, ok := KeywordFuncs[kv.Keyword()]
				if !ok {
					return nil, fmt.Errorf("Unknown keyword %q for file %q", kv.Keyword(), e.Path())
				}
				if keywords != nil && !inSlice(kv.Keyword(), keywords) {
					continue
				}
				fh, err := os.Open(filename)
				if err != nil {
					return nil, err
				}
				curKeyVal, err := keywordFunc(filename, info, fh)
				if err != nil {
					fh.Close()
					return nil, err
				}
				fh.Close()
				if string(kv) != curKeyVal {
					failure := Failure{Path: e.Path(), Keyword: kv.Keyword(), Expected: kv.Value(), Got: KeyVal(curKeyVal).Value()}
					result.Failures = append(result.Failures, failure)
				}
			}
		}
	}
	return &result, nil
}
