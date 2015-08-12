package verify

import "time"

// PosixHeader is the structure from a POSIX tar header, to be marshalled from
// the tar stream, and available for on-disk comparison and verification
type PosixHeader struct {
	Name     string    `json:"name,omitempty"`
	Mode     uint32    `json:"mode,omitempty"`
	UID      uint32    `json:"uid,omitempty"`
	GID      uint32    `json:"gid,omitempty"`
	Size     int       `json:"size,omitempty"`
	Mtime    time.Time `json:"mtime,omitempty"`
	Checksum []byte    `json:"chksum,omitempty"`
	LinkName string    `json:"linkname,omitempty"`
	Magic    []byte    `json:"magic,omitempty"`
	Version  string    `json:"version,omitempty"`
	Uname    string    `json:"uname,omitempty"`
	Gname    string    `json:"gname,omitempty"`
	DevMajor int       `json:"devmajor,omitempty"`
	DevMinor int       `json:"devminor,omitempty"`
	Prefix   string    `json:"prefix,omitempty"`
}
