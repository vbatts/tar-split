package mtree

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// Result of a Check
type Result struct {
	// list of any failures in the Check
	Failures []Failure `json:"failures"`
	Missing  []Entry   `json:"missing,omitempty"`
	Extra    []Entry   `json:"extra,omitempty"`
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
	curDir, err := os.Getwd()
	if err == nil {
		defer os.Chdir(curDir)
	}

	if err := os.Chdir(root); err != nil {
		return nil, err
	}
	sort.Sort(byPos(dh.Entries))

	var result Result
	for _, e := range dh.Entries {
		switch e.Type {
		case RelativeType, FullType:
			pathname, err := e.Path()
			if err != nil {
				return nil, err
			}
			info, err := os.Lstat(pathname)
			if err != nil {
				return nil, err
			}

			kvs := e.AllKeys()

			for _, kv := range kvs {
				kw := kv.Keyword()
				// 'tar_time' keyword evaluation wins against 'time' keyword evaluation
				if kv.Keyword() == "time" && inSlice("tar_time", keywords) {
					kw = "tar_time"
					tartime := fmt.Sprintf("%s.%s", (strings.Split(kv.Value(), ".")[0]), "000000000")
					kv = KeyVal(KeyVal(kw).ChangeValue(tartime))
				}

				keywordFunc, ok := KeywordFuncs[kw]
				if !ok {
					return nil, fmt.Errorf("Unknown keyword %q for file %q", kv.Keyword(), pathname)
				}
				if keywords != nil && !inSlice(kv.Keyword(), keywords) {
					continue
				}

				var curKeyVal string
				if info.Mode().IsRegular() {
					fh, err := os.Open(pathname)
					if err != nil {
						return nil, err
					}
					curKeyVal, err = keywordFunc(pathname, info, fh)
					if err != nil {
						fh.Close()
						return nil, err
					}
					fh.Close()
				} else {
					curKeyVal, err = keywordFunc(pathname, info, nil)
					if err != nil {
						return nil, err
					}
				}
				if string(kv) != curKeyVal {
					failure := Failure{Path: pathname, Keyword: kv.Keyword(), Expected: kv.Value(), Got: KeyVal(curKeyVal).Value()}
					result.Failures = append(result.Failures, failure)
				}
			}
		}
	}
	return &result, nil
}

// TarCheck is the tar equivalent of checking a file hierarchy spec against a tar stream to
// determine if files have been changed.
func TarCheck(tarDH, dh *DirectoryHierarchy, keywords []string) (*Result, error) {
	var result Result
	var err error
	var tarRoot *Entry

	for _, e := range tarDH.Entries {
		if e.Name == "." {
			tarRoot = &e
			break
		}
	}
	if tarRoot == nil {
		return nil, fmt.Errorf("root of archive could not be found")
	}
	tarRoot.Next = &Entry{
		Name: "seen",
		Type: CommentType,
	}
	curDir := tarRoot
	sort.Sort(byPos(dh.Entries))

	var outOfTree bool
	for _, e := range dh.Entries {
		switch e.Type {
		case RelativeType, FullType:
			pathname, err := e.Path()
			if err != nil {
				return nil, err
			}
			if outOfTree {
				return &result, fmt.Errorf("No parent node from %s", pathname)
			}
			// TODO: handle the case where "." is not the first Entry to be found
			tarEntry := curDir.Descend(e.Name)
			if tarEntry == nil {
				result.Missing = append(result.Missing, e)
				continue
			}

			tarEntry.Next = &Entry{
				Type: CommentType,
				Name: "seen",
			}

			// expected values from file hierarchy spec
			kvs := e.AllKeys()

			// actual
			tarkvs := tarEntry.AllKeys()

			for _, kv := range kvs {
				// TODO: keep track of symlinks
				if _, ok := KeywordFuncs[kv.Keyword()]; !ok {
					return nil, fmt.Errorf("Unknown keyword %q for file %q", kv.Keyword(), pathname)
				}
				if keywords != nil && !inSlice(kv.Keyword(), keywords) {
					continue
				}
				tarpath, err := tarEntry.Path()
				if err != nil {
					return nil, err
				}
				if tarkv := tarkvs.Has(kv.Keyword()); tarkv != emptyKV {
					if string(tarkv) != string(kv) {
						failure := Failure{Path: tarpath, Keyword: kv.Keyword(), Expected: kv.Value(), Got: tarkv.Value()}
						result.Failures = append(result.Failures, failure)
					}
				}
			}
			// Step into a directory
			if tarEntry.Prev != nil {
				curDir = tarEntry
			}
		case DotDotType:
			if outOfTree {
				return &result, fmt.Errorf("No parent node.")
			}
			curDir = curDir.Ascend()
			if curDir == nil {
				outOfTree = true
			}
		}
	}
	result.Extra = filter(tarRoot, func(e *Entry) bool {
		return e.Next == nil
	})
	return &result, err
}
