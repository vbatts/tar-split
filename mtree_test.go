package mtree

import (
	"fmt"
	"os"
	"testing"
)

var testFiles = []string{
	"testdata/source.mtree",
}

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
			fmt.Printf("%q", dh)

			_, err = dh.WriteTo(os.Stdout)
			if err != nil {
				t.Error(err)
			}

		}()
	}
}
