// +build !linux

package mtree

import "os"

var (
	unameKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		return "", nil
	}
	uidKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		return "", nil
	}
	gidKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		return "", nil
	}
	nlinkKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		return "", nil
	}
)
