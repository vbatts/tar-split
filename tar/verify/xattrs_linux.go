package verify

/*
Lgetxattr and Lsetxattr are copied directly from https://github.com/docker/docker
  ./pkg/system/xattr_linux.go commit 7e420ad8502089e66ce0ade92bf70574f894f287
  Apache License Version 2.0, January 2004 https://www.apache.org/licenses/
  Copyright 2013-2015 Docker, Inc.
*/

import (
	"bytes"
	"syscall"
	"unsafe"
)

// Listxattr is a helper around the syscall.Listxattr
func Listxattr(path string) ([]string, error) {
	buf := make([]byte, 1024)
	sz, err := syscall.Listxattr(path, buf)
	if err == syscall.ENODATA {
		return nil, nil
	}
	if err == syscall.ERANGE && sz > 0 {
		buf = make([]byte, sz)
		sz, err = syscall.Listxattr(path, buf)
	}
	keys := []string{}
	for _, key := range bytes.Split(bytes.Trim(buf, "\x00"), []byte{0x0}) {
		if string(key) != "" {
			keys = append(keys, string(key))
		}
	}
	return keys, nil
}

// Lgetxattr retrieves the value of the extended attribute identified by attr
// and associated with the given path in the file system.
// It will returns a nil slice and nil error if the xattr is not set.
func Lgetxattr(path string, attr string) ([]byte, error) {
	pathBytes, err := syscall.BytePtrFromString(path)
	if err != nil {
		return nil, err
	}
	attrBytes, err := syscall.BytePtrFromString(attr)
	if err != nil {
		return nil, err
	}

	dest := make([]byte, 128)
	destBytes := unsafe.Pointer(&dest[0])
	sz, _, errno := syscall.Syscall6(syscall.SYS_LGETXATTR, uintptr(unsafe.Pointer(pathBytes)), uintptr(unsafe.Pointer(attrBytes)), uintptr(destBytes), uintptr(len(dest)), 0, 0)
	if errno == syscall.ENODATA {
		return nil, nil
	}
	if errno == syscall.ERANGE {
		dest = make([]byte, sz)
		destBytes := unsafe.Pointer(&dest[0])
		sz, _, errno = syscall.Syscall6(syscall.SYS_LGETXATTR, uintptr(unsafe.Pointer(pathBytes)), uintptr(unsafe.Pointer(attrBytes)), uintptr(destBytes), uintptr(len(dest)), 0, 0)
	}
	if errno != 0 {
		return nil, errno
	}

	return dest[:sz], nil
}

var _zero uintptr

// Lsetxattr sets the value of the extended attribute identified by attr
// and associated with the given path in the file system.
func Lsetxattr(path string, attr string, data []byte, flags int) error {
	pathBytes, err := syscall.BytePtrFromString(path)
	if err != nil {
		return err
	}
	attrBytes, err := syscall.BytePtrFromString(attr)
	if err != nil {
		return err
	}
	var dataBytes unsafe.Pointer
	if len(data) > 0 {
		dataBytes = unsafe.Pointer(&data[0])
	} else {
		dataBytes = unsafe.Pointer(&_zero)
	}
	_, _, errno := syscall.Syscall6(syscall.SYS_LSETXATTR, uintptr(unsafe.Pointer(pathBytes)), uintptr(unsafe.Pointer(attrBytes)), uintptr(dataBytes), uintptr(len(data)), uintptr(flags), 0)
	if errno != 0 {
		return errno
	}
	return nil
}