package mtree

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Result struct {
	// XXX perhaps this is a list of the failed files and keywords?
}

var ErrNotAllClear = fmt.Errorf("some keyword check failed validation")

func Check(root string, dh *DirectoryHierarchy) (*Result, error) {
	creator := dhCreator{DH: dh}
	curDir, err := os.Getwd()
	if err == nil {
		defer os.Chdir(curDir)
	}

	if err := os.Chdir(root); err != nil {
		return nil, err
	}
	sort.Sort(byPos(creator.DH.Entries))

	var failed bool
	for _, e := range creator.DH.Entries {
		switch e.Type {
		case SpecialType:
			if e.Name == "/set" {
				creator.curSet = &e
			} else if e.Name == "/unset" {
				creator.curSet = nil
			}
		case RelativeType, FullType:
			info, err := os.Lstat(filepath.Join(root, e.Path()))
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
				curKeyVal, err := keywordFunc(filepath.Join(root, e.Path()), info)
				if err != nil {
					return nil, err
				}
				if string(kv) != curKeyVal {
					failed = true
					fmt.Printf("%q: keyword %q: expected %s; got %s", e.Path(), kv.Keyword(), kv.Value(), KeyVal(curKeyVal).Value())
				}
			}
		}
	}

	if failed {
		return nil, ErrNotAllClear
	}

	return nil, nil
}
