package mtree

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Result struct {
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

	for _, e := range creator.DH.Entries {
		switch e.Type {
		case SpecialType:
			if e.Name == "/set" {
				creator.curSet = &e
			} else if e.Name == "/unset" {
				creator.curSet = nil
			}
		case DotDotType:
			// TODO step
		case RelativeType:
			// TODO determine path, and check keywords
			//      or maybe to Chdir when type=dir?
		case FullType:
			info, err := os.Lstat(filepath.Join(root, e.Name))
			if err != nil {
				return nil, err
			}
			// TODO check against keywords present
			_ = info
		}
	}

	return nil, nil
}
