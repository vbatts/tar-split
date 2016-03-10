package mtree

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
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
				log.Println(err)
				return
			}
			defer fh.Close()

			s := bufio.NewScanner(fh)
			for s.Scan() {
				str := s.Text()
				switch {
				case strings.HasPrefix(str, "#"):
					continue
				default:
				}
				fmt.Printf("%q\n", str)
			}
			if err := s.Err(); err != nil {
				log.Println("ERROR:", err)
			}
		}()
	}
}
