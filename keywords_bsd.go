// +build darwin freebsd netbsd openbsd

package mtree

import (
	"io"
	"os"
)

var (
	flagsKeywordFunc = func(path string, info os.FileInfo, r io.Reader) (string, error) {
		// ideally this will pull in from here https://www.freebsd.org/cgi/man.cgi?query=chflags&sektion=2
		return "", nil
	}
)
