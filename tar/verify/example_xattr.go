// +build ignore

package main

import (
	"fmt"
	"os"

	verify "."
)

func main() {
	for _, arg := range os.Args[1:] {
		keys, err := verify.Listxattr(arg)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if len(keys) > 0 {
			fmt.Printf("%s : %q\n", arg, keys)
			for _, key := range keys {
				buf, err := verify.Lgetxattr(arg, key)
				if err != nil {
					fmt.Printf("  ERROR: %s\n", err)
					continue
				}
				fmt.Printf("  %s = %s\n", key, string(buf))
			}
		}
	}
}
