package mtree

import (
	"io/ioutil"
	"os"
	"testing"
)

var (
	testFiles  = []string{"testdata/source.mtree"}
	numEntries = map[EntryType]int{
		FullType:     0,
		RelativeType: 45,
		CommentType:  37,
		SpecialType:  7,
		DotDotType:   17,
		BlankType:    34,
	}
	expectedLength = int64(7887)
)

func TestParser(t *testing.T) {
	for _, file := range testFiles {
		func() {
			fh, err := os.Open(file)
			if err != nil {
				t.Error(err)
				return
			}
			defer fh.Close()

			dh, err := ParseSpec(fh)
			if err != nil {
				t.Error(err)
			}
			gotNums := countTypes(dh)
			for typ, num := range numEntries {
				if gNum, ok := gotNums[typ]; ok {
					if num != gNum {
						t.Errorf("for type %s: expected %d, got %d", typ, num, gNum)
					}
				}
			}

			i, err := dh.WriteTo(ioutil.Discard)
			if err != nil {
				t.Error(err)
			}
			if i != expectedLength {
				t.Errorf("expected to write %d, but wrote %d", expectedLength, i)
			}

		}()
	}
}

func countTypes(dh *DirectoryHierarchy) map[EntryType]int {
	nT := map[EntryType]int{}
	for i := range dh.Entries {
		typ := dh.Entries[i].Type
		if _, ok := nT[typ]; !ok {
			nT[typ] = 1
		} else {
			nT[typ]++
		}
	}
	return nT
}
