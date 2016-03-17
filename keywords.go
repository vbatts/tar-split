package mtree

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"math"
	"os"

	"golang.org/x/crypto/ripemd160"
)

// KeywordFunc is the type of a function called on each file to be included in
// a DirectoryHierarchy, that will produce the string output of the keyword to
// be included for the file entry. Otherwise, empty string.
type KeywordFunc func(path string, info os.FileInfo) (string, error)

var (
	// DefaultKeywords has the several default keyword producers (uid, gid,
	// mode, nlink, type, size, mtime)
	DefaultKeywords = []string{
		"size",
		"type",
		"uid",
		"gid",
		"mode",
		"link",
		"nlink",
		"time",
	}
	SetKeywords = []string{
		"uid",
		"gid",
	}
	// KeywordFuncs is the map of all keywords (and the functions to produce them)
	KeywordFuncs = map[string]KeywordFunc{
		"size":            sizeKeywordFunc,                                     // The size, in bytes, of the file
		"type":            typeKeywordFunc,                                     // The type of the file
		"time":            timeKeywordFunc,                                     // The last modification time of the file
		"link":            linkKeywordFunc,                                     // The target of the symbolic link when type=link
		"uid":             uidKeywordFunc,                                      // The file owner as a numeric value
		"gid":             gidKeywordFunc,                                      // The file group as a numeric value
		"nlink":           nlinkKeywordFunc,                                    // The number of hard links the file is expected to have
		"uname":           unameKeywordFunc,                                    // The file owner as a symbolic name
		"mode":            modeKeywordFunc,                                     // The current file's permissions as a numeric (octal) or symbolic value
		"cksum":           cksumKeywordFunc,                                    // The checksum of the file using the default algorithm specified by the cksum(1) utility
		"md5":             hasherKeywordFunc("md5", md5.New),                   // The MD5 message digest of the file
		"md5digest":       hasherKeywordFunc("md5digest", md5.New),             // A synonym for `md5`
		"rmd160":          hasherKeywordFunc("rmd160", ripemd160.New),          // The RIPEMD160 message digest of the file
		"rmd160digest":    hasherKeywordFunc("rmd160digest", ripemd160.New),    // A synonym for `rmd160`
		"ripemd160digest": hasherKeywordFunc("ripemd160digest", ripemd160.New), // A synonym for `rmd160`
		"sha1":            hasherKeywordFunc("sha1", sha1.New),                 // The SHA1 message digest of the file
		"sha1digest":      hasherKeywordFunc("sha1digest", sha1.New),           // A synonym for `sha1`
		"sha256":          hasherKeywordFunc("sha256", sha256.New),             // The SHA256 message digest of the file
		"sha256digest":    hasherKeywordFunc("sha256digest", sha256.New),       // A synonym for `sha256`
		"sha384":          hasherKeywordFunc("sha384", sha512.New384),          // The SHA384 message digest of the file
		"sha384digest":    hasherKeywordFunc("sha384digest", sha512.New384),    // A synonym for `sha384`
		"sha512":          hasherKeywordFunc("sha512", sha512.New),             // The SHA512 message digest of the file
		"sha512digest":    hasherKeywordFunc("sha512digest", sha512.New),       // A synonym for `sha512`

		// This is not an upstreamed keyword, but a needed attribute for file validation.
		// The pattern for this keyword key is prefixed by "xattr." followed by the extended attribute "namespace.key".
		// The keyword value is the SHA1 digest of the extended attribute's value.
		// In this way, the order of the keys does not matter, and the contents of the value is not revealed.
		"xattr": xattrKeywordFunc,
	}
)

var (
	modeKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		return fmt.Sprintf("mode=%#o", info.Mode().Perm()), nil
	}
	sizeKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		return fmt.Sprintf("size=%d", info.Size()), nil
	}
	cksumKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		if !info.Mode().IsRegular() {
			return "", nil
		}

		fh, err := os.Open(path)
		if err != nil {
			return "", err
		}
		defer fh.Close()
		sum, _, err := cksum(fh)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("cksum=%d", sum), nil
	}
	hasherKeywordFunc = func(name string, newHash func() hash.Hash) KeywordFunc {
		return func(path string, info os.FileInfo) (string, error) {
			if !info.Mode().IsRegular() {
				return "", nil
			}

			fh, err := os.Open(path)
			if err != nil {
				return "", err
			}
			defer fh.Close()

			h := newHash()
			if _, err := io.Copy(h, fh); err != nil {
				return "", err
			}
			return fmt.Sprintf("%s=%x", name, h.Sum(nil)), nil
		}
	}
	timeKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		t := info.ModTime()
		n := float64(t.UnixNano()) / math.Pow10(9)
		return fmt.Sprintf("time=%0.9f", n), nil
	}
	linkKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		if info.Mode()&os.ModeSymlink != 0 {
			str, err := os.Readlink(path)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("link=%s", str), nil
		}
		return "", nil
	}
	typeKeywordFunc = func(path string, info os.FileInfo) (string, error) {
		if info.Mode().IsDir() {
			return "type=dir", nil
		}
		if info.Mode().IsRegular() {
			return "type=file", nil
		}
		if info.Mode()&os.ModeSocket != 0 {
			return "type=socket", nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "type=link", nil
		}
		if info.Mode()&os.ModeNamedPipe != 0 {
			return "type=fifo", nil
		}
		if info.Mode()&os.ModeDevice != 0 {
			if info.Mode()&os.ModeCharDevice != 0 {
				return "type=char", nil
			}
			return "type=device", nil
		}
		return "", nil
	}
)
