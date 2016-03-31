// +build linux

package mtree

import (
	"crypto/sha1"
	"fmt"
	"os"
	"os/user"
	"strings"
	"syscall"

	"github.com/vbatts/go-mtree/xattr"
)

var (
	unameKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		stat := info.Sys().(*syscall.Stat_t)
		u, err := user.LookupId(fmt.Sprintf("%d", stat.Uid))
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("uname=%s", u.Username), nil
	}
	uidKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		stat := info.Sys().(*syscall.Stat_t)
		return fmt.Sprintf("uid=%d", stat.Uid), nil
	}
	gidKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		stat := info.Sys().(*syscall.Stat_t)
		return fmt.Sprintf("gid=%d", stat.Gid), nil
	}
	nlinkKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		stat := info.Sys().(*syscall.Stat_t)
		return fmt.Sprintf("nlink=%d", stat.Nlink), nil
	}
	xattrKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		xlist, err := xattr.List(path)
		if err != nil {
			return "", err
		}
		klist := make([]string, len(xlist))
		for i := range xlist {
			data, err := xattr.Get(path, xlist[i])
			if err != nil {
				return "", err
			}
			klist[i] = fmt.Sprintf("xattr.%s=%x", xlist[i], sha1.Sum(data))
		}
		return strings.Join(klist, " "), nil
	}
)
