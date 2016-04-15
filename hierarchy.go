package mtree

import (
	"io"
	"sort"
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
