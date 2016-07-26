package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/vbatts/go-mtree"
)

var (
	flCreate       = flag.Bool("c", false, "create a directory hierarchy spec")
	flFile         = flag.String("f", "", "directory hierarchy spec to validate")
	flPath         = flag.String("p", "", "root path that the hierarchy spec is relative to")
	flAddKeywords  = flag.String("K", "", "Add the specified (delimited by comma or space) keywords to the current set of keywords")
	flUseKeywords  = flag.String("k", "", "Use the specified (delimited by comma or space) keywords as the current set of keywords")
	flListKeywords = flag.Bool("list-keywords", false, "List the keywords available")
	flResultFormat = flag.String("result-format", "bsd", "output the validation results using the given format (bsd, json, path)")
	flTar          = flag.String("T", "", "use tar archive to create or validate a directory hierarchy spec")
)

var formats = map[string]func(*mtree.Result) string{
	// Outputs the errors in the BSD format.
	"bsd": func(r *mtree.Result) string {
		var buffer bytes.Buffer
		for _, fail := range r.Failures {
			fmt.Fprintln(&buffer, fail)
		}
		return buffer.String()
	},

	// Outputs the full result struct in JSON.
	"json": func(r *mtree.Result) string {
		var buffer bytes.Buffer
		if err := json.NewEncoder(&buffer).Encode(r); err != nil {
			panic(err)
		}
		return buffer.String()
	},

	// Outputs only the paths which failed to validate.
	"path": func(r *mtree.Result) string {
		var buffer bytes.Buffer
		for _, fail := range r.Failures {
			fmt.Fprintln(&buffer, fail.Path)
		}
		return buffer.String()
	},
}

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

	// --output
	formatFunc, ok := formats[*flResultFormat]
	if !ok {
		log.Printf("invalid output format: %s", *flResultFormat)
		isErr = true
		return
	}

	var currentKeywords []string
	// -k <keywords>
	if *flUseKeywords != "" {
		currentKeywords = splitKeywordsArg(*flUseKeywords)
		if !inSlice("type", currentKeywords) {
			currentKeywords = append([]string{"type"}, currentKeywords...)
		}
	} else {
		if *flTar != "" {
			currentKeywords = mtree.DefaultTarKeywords[:]
		} else {
			currentKeywords = mtree.DefaultKeywords[:]
		}
	}
	// -K <keywords>
	if *flAddKeywords != "" {
		for _, kw := range splitKeywordsArg(*flAddKeywords) {
			if !inSlice(kw, currentKeywords) {
				currentKeywords = append(currentKeywords, kw)
			}
		}
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
	var rootPath = "."
	if *flPath != "" {
		rootPath = *flPath
	}

	// -T <tar file>
	var tdh *mtree.DirectoryHierarchy
	if *flTar != "" {
		fh, err := os.Open(*flTar)
		if err != nil {
			log.Println(err)
			isErr = true
			return
		}
		ts := mtree.NewTarStreamer(fh, currentKeywords)

		if _, err := io.Copy(ioutil.Discard, ts); err != nil && err != io.EOF {
			log.Println(err)
			isErr = true
			return
		}
		if err := ts.Close(); err != nil {
			log.Println(err)
			isErr = true
			return
		}
		defer fh.Close()
		tdh, err = ts.Hierarchy()
		if err != nil {
			log.Println(err)
			isErr = true
			return
		}
	}
	// -c
	if *flCreate {
		// create a directory hierarchy
		// with a tar stream
		if tdh != nil {
			tdh.WriteTo(os.Stdout)
		} else {
			// with a root directory
			dh, err := mtree.Walk(rootPath, nil, currentKeywords)
			if err != nil {
				log.Println(err)
				isErr = true
				return
			}
			dh.WriteTo(os.Stdout)
		}
	} else if tdh != nil || dh != nil {
		var res *mtree.Result
		var err error
		// else this is a validation
		if *flTar != "" {
			res, err = mtree.TarCheck(tdh, dh, currentKeywords)
		} else {
			res, err = mtree.Check(rootPath, dh, currentKeywords)
		}
		if err != nil {
			log.Println(err)
			isErr = true
			return
		}
		if res != nil && len(res.Failures) > 0 {
			defer os.Exit(1)
			out := formatFunc(res)
			if _, err := os.Stdout.Write([]byte(out)); err != nil {
				log.Println(err)
				isErr = true
				return
			}
		}
		if res != nil {
			if len(res.Extra) > 0 {
				defer os.Exit(1)
				for _, extra := range res.Extra {
					extrapath, err := extra.Path()
					if err != nil {
						log.Println(err)
						isErr = true
						return
					}
					fmt.Printf("%s extra\n", extrapath)
				}
			}
			if len(res.Missing) > 0 {
				defer os.Exit(1)
				for _, missing := range res.Missing {
					missingpath, err := missing.Path()
					if err != nil {
						log.Println(err)
						isErr = true
						return
					}
					fmt.Printf("%s missing\n", missingpath)
				}
			}
		} else {
			log.Println("neither validating or creating a manifest. Please provide additional arguments")
			isErr = true
			defer os.Exit(1)
			return
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
