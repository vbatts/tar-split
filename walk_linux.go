// +build linux

package mtree

import (
	"fmt"
	"os"
	"os/user"
	"syscall"
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
)
