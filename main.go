package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"./archive/tar"
)

func main() {
	flag.Parse()
	log.SetOutput(os.Stderr)
	for _, arg := range flag.Args() {
		func() {
			// Open the tar archive
			fh, err := os.Open(arg)
			if err != nil {
				log.Fatal(err)
			}
			defer fh.Close()

			fi, err := fh.Stat()
			if err != nil {
				log.Fatal(err)
			}
			size := fi.Size()
			var sum int64
			tr := tar.NewReader(fh)
			tr.RawAccounting = true
			for {
				hdr, err := tr.Next()
				if err != nil {
					if err != io.EOF {
						log.Println(err)
					}
					break
				}
				pre := tr.RawBytes()
				var i int64
				if i, err = io.Copy(ioutil.Discard, tr); err != nil {
					log.Println(err)
					break
				}
				post := tr.RawBytes()
				fmt.Println(hdr.Name, "pre:", len(pre), "read:", i, "post:", len(post))
				sum += int64(len(pre))
				sum += i
				sum += int64(len(post))
			}

			if size != sum {
				fmt.Printf("Size: %d; Sum: %d; Diff: %d\n", size, sum, size-sum)
			} else {
				fmt.Printf("Size: %d; Sum: %d\n", size, sum)
			}
		}()
	}
}
