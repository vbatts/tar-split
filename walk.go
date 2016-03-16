package mtree

import (
	"os"
	"path/filepath"
)

// ExcludeFunc is the type of function called on each path walked to determine
// whether to be excluded from the assembled DirectoryHierarchy. If the func
// returns true, then the path is not included in the spec.
type ExcludeFunc func(path string, info os.FileInfo) bool

//
// To be able to do a "walk" that produces an outcome with `/set ...` would
// need a more linear walk, which this can not ensure.
func Walk(root string, exlcudes []ExcludeFunc, keywords []string) (*DirectoryHierarchy, error) {
	dh := DirectoryHierarchy{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, ex := range exlcudes {
			if ex(path, info) {
				return nil
			}
		}
		e := Entry{}
		//e.Name = filepath.Base(path)
		e.Name = path
		for _, keyword := range keywords {
			if str, err := KeywordFuncs[keyword](path, info); err == nil && str != "" {
				e.Keywords = append(e.Keywords, str)
			} else if err != nil {
				return err
			}
		}
		// XXX
		dh.Entries = append(dh.Entries, e)
		return nil
	})
	return &dh, err
}
