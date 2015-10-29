package verify

import (
	"fmt"
	"os"
	"time"
)

// CheckType is how the on disk attributes will be verified against the
// recorded header information
type CheckType int

// Check types for customizing how fuzzy or strict on-disk verification will be
// handled
const (
	CheckDigest CheckType = iota
	CheckFileSize
	CheckFileMode
	CheckFileUser
	CheckFileGroup
	CheckFileMtime
	CheckFileDevice // major/minor
	CheckFileLink   // linkname
	CheckFileCaps   // which is just a subset of xattrs on linux
)

var (
	// DefaultChecks is the default for verfication steps against each
	// storage.VerficationEntry.
	// These may need to vary from platform to platform
	DefaultChecks = CheckDigest | CheckFileAttributes
	// CheckFileAttributes are the group of file attribute checks done
	CheckFileAttributes = CheckFileSize | CheckFileMode | CheckFileUser |
		CheckFileGroup | CheckFileMtime | CheckFileDevice | CheckFileCaps |
		CheckFileLink

	// ErrNotSupportedPlatform is when the platform does not support given features
	ErrNotSupportedPlatform = fmt.Errorf("platform does not support this feature")
)

// FileAttrer exposes the functions corresponding to file attribute checks
type FileAttrer interface {
	Sizer
	Moder
	ModTimer
	LinkName() string
}

// Sizer returns the size of the file (see also os.FileInfo)
type Sizer interface {
	Size() int64
}

// Moder returns the mode of the file (see also os.FileInfo)
type Moder interface {
	Mode() os.FileMode
}

// ModTimer returns the mtime of the file (see also os.FileInfo)
type ModTimer interface {
	ModTime() time.Time
}
