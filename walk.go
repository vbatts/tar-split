package mtree

import (
	"os"
	"path/filepath"
	"sort"
)

// ExcludeFunc is the type of function called on each path walked to determine
// whether to be excluded from the assembled DirectoryHierarchy. If the func
// returns true, then the path is not included in the spec.
type ExcludeFunc func(path string, info os.FileInfo) bool

type dhCreator struct {
	DH     *DirectoryHierarchy
	curSet *Entry
	curDir *Entry
	curEnt *Entry
}

var defaultSetKeywords = []string{"type=file", "nlink=1", "flags=none", "mode=0664"}

//
// To be able to do a "walk" that produces an outcome with `/set ...` would
// need a more linear walk, which this can not ensure.
func Walk(root string, exlcudes []ExcludeFunc, keywords []string) (*DirectoryHierarchy, error) {
	creator := dhCreator{DH: &DirectoryHierarchy{}}
	// TODO insert signature and metadata comments first (user, machine, tree, date)
	err := startWalk(&creator, root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for _, ex := range exlcudes {
			if ex(path, info) {
				return nil
			}
		}

		if info.IsDir() {
			creator.DH.Entries = append(creator.DH.Entries, Entry{
				Type: BlankType,
				Pos:  len(creator.DH.Entries),
			})

			// TODO Insert a comment of the full path of the directory's name
			if creator.curDir != nil {
				creator.DH.Entries = append(creator.DH.Entries, Entry{
					Pos:  len(creator.DH.Entries),
					Raw:  "# " + filepath.Join(creator.curDir.Path(), filepath.Base(path)),
					Type: CommentType,
				})
			} else {
				creator.DH.Entries = append(creator.DH.Entries, Entry{
					Pos:  len(creator.DH.Entries),
					Raw:  "# " + filepath.Base(path),
					Type: CommentType,
				})
			}

			// set the initial /set keywords
			if creator.curSet == nil {
				e := Entry{
					Name:     "/set",
					Type:     SpecialType,
					Pos:      len(creator.DH.Entries),
					Keywords: keywordSelector(defaultSetKeywords, keywords),
				}
				for _, keyword := range SetKeywords {
					if str, err := KeywordFuncs[keyword](path, info); err == nil && str != "" {
						e.Keywords = append(e.Keywords, str)
					} else if err != nil {
						return err
					}
				}
				creator.curSet = &e
				creator.DH.Entries = append(creator.DH.Entries, e)
			} else if creator.curSet != nil {
				// check the attributes of the /set keywords and re-set if changed
				klist := []string{}
				for _, keyword := range SetKeywords {
					if str, err := KeywordFuncs[keyword](path, info); err == nil && str != "" {
						klist = append(klist, str)
					} else if err != nil {
						return err
					}
				}

				needNewSet := false
				for _, k := range klist {
					if !inSlice(k, creator.curSet.Keywords) {
						needNewSet = true
					}
				}
				if needNewSet {
					e := Entry{
						Name:     "/set",
						Type:     SpecialType,
						Pos:      len(creator.DH.Entries),
						Keywords: append(defaultSetKeywords, klist...),
					}
					creator.curSet = &e
					creator.DH.Entries = append(creator.DH.Entries, e)
				}
			}
		}

		e := Entry{
			Name:   filepath.Base(path),
			Pos:    len(creator.DH.Entries),
			Set:    creator.curSet,
			Parent: creator.curDir,
		}
		for _, keyword := range keywords {
			if str, err := KeywordFuncs[keyword](path, info); err == nil && str != "" {
				if !inSlice(str, creator.curSet.Keywords) {
					e.Keywords = append(e.Keywords, str)
				}
			} else if err != nil {
				return err
			}
		}
		if info.IsDir() {
			if creator.curDir != nil {
				creator.curDir.Next = &e
			}
			e.Prev = creator.curDir
			creator.curDir = &e
		} else {
			if creator.curEnt != nil {
				creator.curEnt.Next = &e
			}
			e.Prev = creator.curEnt
			creator.curEnt = &e
		}
		creator.DH.Entries = append(creator.DH.Entries, e)
		return nil
	})
	return creator.DH, err
}

func inSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// startWalk walks the file tree rooted at root, calling walkFn for each file or
// directory in the tree, including root. All errors that arise visiting files
// and directories are filtered by walkFn. The files are walked in lexical
// order, which makes the output deterministic but means that for very
// large directories Walk can be inefficient.
// Walk does not follow symbolic links.
func startWalk(c *dhCreator, root string, walkFn filepath.WalkFunc) error {
	info, err := os.Lstat(root)
	if err != nil {
		return walkFn(root, nil, err)
	}
	return walk(c, root, info, walkFn)
}

// walk recursively descends path, calling w.
func walk(c *dhCreator, path string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	err := walkFn(path, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return nil
	}

	names, err := readOrderedDirNames(path)
	if err != nil {
		return walkFn(path, info, err)
	}

	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := os.Lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walk(c, filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}
	c.DH.Entries = append(c.DH.Entries, Entry{
		Name: "..",
		Type: DotDotType,
		Pos:  len(c.DH.Entries),
	})
	if c.curDir != nil {
		c.curDir = c.curDir.Parent
	}
	return nil
}

// readOrderedDirNames reads the directory and returns a sorted list of all
// entries with non-directories first, followed by directories.
func readOrderedDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	infos, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	names := []string{}
	dirnames := []string{}
	for _, info := range infos {
		if info.IsDir() {
			dirnames = append(dirnames, info.Name())
			continue
		}
		names = append(names, info.Name())
	}
	sort.Strings(names)
	sort.Strings(dirnames)
	return append(names, dirnames...), nil
}
