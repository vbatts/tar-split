package verify

import "fmt"

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
	CheckFileDevice
	CheckFileLink
	CheckFileCaps
)

var (
	// DefaultChecks is the default for verfication steps against each
	// storage.VerficationEntry
	DefaultChecks = CheckDigest | CheckFileAttributes
	// CheckFileAttributes are the group of file attribute checks done
	CheckFileAttributes = CheckFileSize | CheckFileMode | CheckFileUser |
		CheckFileGroup | CheckFileMtime | CheckFileDevice | CheckFileCaps |
		CheckFileLink

	// ErrNotSupportedPlatform is when the platform does not support given features
	ErrNotSupportedPlatform = fmt.Errorf("platform does not support this feature")
)
