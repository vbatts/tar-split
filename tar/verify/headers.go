package verify

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PosixHeader is the structure from a POSIX tar header, to be marshalled from
// the tar stream, and available for on-disk comparison and verification
type PosixHeader struct {
	Name     string      `json:"name,omitempty"`
	Mode     os.FileMode `json:"mode,omitempty"`
	UID      uint32      `json:"uid,omitempty"`
	GID      uint32      `json:"gid,omitempty"`
	Size     int64       `json:"size,omitempty"`
	Mtime    time.Time   `json:"mtime,omitempty"`
	Checksum []byte      `json:"chksum,omitempty"`
	LinkName string      `json:"linkname,omitempty"`
	Magic    []byte      `json:"magic,omitempty"`
	Version  string      `json:"version,omitempty"`
	Uname    string      `json:"uname,omitempty"`
	Gname    string      `json:"gname,omitempty"`
	DevMajor int         `json:"devmajor,omitempty"`
	DevMinor int         `json:"devminor,omitempty"`
	Prefix   string      `json:"prefix,omitempty"`
}

type PaxHeader struct {
	Atime  time.Time
	Ctime  time.Time
	Xattrs map[string]string
}

type Header struct {
	// maybe I do not want these grouped like this. Maybe this should be an interface instead.
	Posix PosixHeader
	Pax   PaxHeader
}

// Size returns file size (implements Sizer)
func (hdr Header) Size() int64 {
	return int64(hdr.Posix.Size)
}

// ModTime returns file mtime (implements ModTimer)
func (hdr Header) ModTime() time.Time {
	return hdr.Posix.Mtime
}

// Mode returns file mode (implements Moder)
func (hdr Header) Mode() os.FileMode {
	return hdr.Posix.Mode
}

func (hdr Header) LinkName() string {
	return hdr.Posix.LinkName
}

// HeaderFromFile takes a relative root and the filename of the file to collect
// information on.
func HeaderFromFile(rel, filename string) (*Header, error) {
	absRel, err := filepath.Abs(rel)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(filename, "/") {
		var err error
		filename, err = filepath.Abs(filename)
		if err != nil {
			return nil, err
		}
	}

	stat, err := os.Lstat(filename)
	if err != nil {
		return nil, err
	}

	name := filename
	if strings.HasPrefix(filename, absRel) {
		name = strings.TrimPrefix(filename, absRel)
	}

	hdr := Header{
		Posix: PosixHeader{
			Name:  name,
			Size:  stat.Size(),
			Mode:  stat.Mode(),
			Mtime: stat.ModTime(),
		},
	}
	if stat.Mode()&os.ModeSymlink != 0 {
		l, _ := os.Readlink(filename) // if this errors, the empty string is OK
		hdr.Posix.LinkName = l
	}

	return &hdr, nil
}
