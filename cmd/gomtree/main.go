package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	mtree "../.."
)

var (
	flCreate       = flag.Bool("c", false, "create a directory hierarchy spec")
	flFile         = flag.String("f", "", "directory hierarchy spec to validate")
	flPath         = flag.String("p", "", "root path that the hierarchy spec is relative to")
	flAddKeywords  = flag.String("K", "", "Add the specified (delimited by comma or space) keywords to the current set of keywords")
	flUseKeywords  = flag.String("k", "", "Use the specified (delimited by comma or space) keywords as the current set of keywords")
	flListKeywords = flag.Bool("l", false, "List the keywords available")
)

func main() {
	flag.Parse()

	// so that defers cleanly exec
	var isErr bool
	defer func() {
		if isErr {
			os.Exit(1)
		}
	}()

	// -l
	if *flListKeywords {
		fmt.Println("Available keywords:")
		for k := range mtree.KeywordFuncs {
			if inSlice(k, mtree.DefaultKeywords) {
				fmt.Println(" ", k, " (default)")
			} else {
				fmt.Println(" ", k)
			}
		}
		return
	}

	var currentKeywords []string
	// -k <keywords>
	if *flUseKeywords != "" {
		currentKeywords = splitKeywordsArg(*flUseKeywords)
	} else {
		currentKeywords = mtree.DefaultKeywords[:]
	}
	// -K <keywords>
	if *flAddKeywords != "" {
		currentKeywords = append(currentKeywords, splitKeywordsArg(*flAddKeywords)...)
	}

	// -f <file>
	var dh *mtree.DirectoryHierarchy
	if *flFile != "" && !*flCreate {
		// load the hierarchy, if we're not creating a new spec
		fh, err := os.Open(*flFile)
		if err != nil {
			log.Println(err)
			isErr = true
			return
		}
		dh, err = mtree.ParseSpec(fh)
		fh.Close()
		if err != nil {
			log.Println(err)
			isErr = true
			return
		}
	}

	// -p <path>
	var rootPath string = "."
	if *flPath != "" {
		rootPath = *flPath
	}

	// -c
	if *flCreate {
		// create a directory hierarchy
		dh, err := mtree.Walk(rootPath, nil, currentKeywords)
		if err != nil {
			log.Println(err)
			isErr = true
			return
		}
		dh.WriteTo(os.Stdout)
	} else {
		// else this is a validation
		res, err := mtree.Check(rootPath, dh)
		if err != nil {
			log.Println(err)
			isErr = true
			return
		}
		if res != nil {
			fmt.Printf("%#v\n", res)
		}
	}
}

func splitKeywordsArg(str string) []string {
	return strings.Fields(strings.Replace(str, ",", " ", -1))
}

func inSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
