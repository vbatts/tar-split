package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	flag.Parse()
	_ = mtree.DirectoryHierarchy{}
	for _, file := range flag.Args() {
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
