// +build linux

package mtree

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/vbatts/go-mtree/xattr"
)

func TestXattr(t *testing.T) {
	// a bit dirty to create/destory a directory in cwd, but often /tmp is
	// mounted tmpfs and doesn't support xattrs
	dir, err := ioutil.TempDir(".", "test.xattrs.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fh, err := os.Create(filepath.Join(dir, "file"))
	if err != nil {
		t.Fatal(err)
	}
	fh.WriteString("howdy")
	fh.Sync()
	if _, err := fh.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	if err := os.Symlink("./no/such/path", filepath.Join(dir, "symlink")); err != nil {
		t.Fatal(err)
	}

	if err := xattr.Set(dir, "user.test", []byte("directory")); err != nil {
		t.Skip(fmt.Sprintf("skipping: %q does not support xattrs", dir))
	}
	if err := xattr.Set(filepath.Join(dir, "file"), "user.test", []byte("regular file")); err != nil {
		t.Fatal(err)
	}

	dirstat, err := os.Lstat(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Check the directory
	str, err := xattrKeywordFunc(dir, dirstat, nil)
	if err != nil {
		t.Error(err)
	}
	if str == "" {
		t.Errorf("expected a keyval; got %q", str)
	}

	filestat, err := fh.Stat()
	if err != nil {
		t.Fatal(err)
	}
	// Check the regular file
	str, err = xattrKeywordFunc(filepath.Join(dir, "file"), filestat, fh)
	if err != nil {
		t.Error(err)
	}
	if str == "" {
		t.Errorf("expected a keyval; got %q", str)
	}

	linkstat, err := os.Lstat(filepath.Join(dir, "symlink"))
	if err != nil {
		t.Fatal(err)
	}
	// Check a broken symlink
	_, err = xattrKeywordFunc(filepath.Join(dir, "symlink"), linkstat, nil)
	if err != nil {
		t.Error(err)
	}
}
