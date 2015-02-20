package main

import (
	"flag"
	"fmt"
	"io"
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

			output, err := os.Create(fmt.Sprintf("%s.out", arg))
			if err != nil {
				log.Fatal(err)
			}
			defer output.Close()
			log.Printf("writing %q to %q", fh.Name(), output.Name())

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
				output.Write(pre)
				sum += int64(len(pre))

				var i int64
				if i, err = io.Copy(output, tr); err != nil {
					log.Println(err)
					break
				}
				sum += i

				post := tr.RawBytes()
				output.Write(post)
				sum += int64(len(post))

				fmt.Println(hdr.Name, "pre:", len(pre), "read:", i, "post:", len(post))
			}

			if size != sum {
				fmt.Printf("Size: %d; Sum: %d; Diff: %d\n", size, sum, size-sum)
				fmt.Printf("Compare like `cmp -bl %s %s | less`\n", fh.Name(), output.Name())
			} else {
				fmt.Printf("Size: %d; Sum: %d\n", size, sum)
			}
		}()
	}
}
