package mtree

import (
	"fmt"
	"os"
	"testing"
	"time"
)

type fakeFileInfo struct {
	mtime time.Time
}

func (ffi fakeFileInfo) Name() string {
	// noop
	return ""
}

func (ffi fakeFileInfo) Size() int64 {
	// noop
	return -1
}

func (ffi fakeFileInfo) Mode() os.FileMode {
	// noop
	return 0
}

func (ffi fakeFileInfo) ModTime() time.Time {
	return ffi.mtime
}

func (ffi fakeFileInfo) IsDir() bool {
	return ffi.Mode().IsDir()
}

func (ffi fakeFileInfo) Sys() interface{} {
	// noop
	return nil
}

func TestKeywordsTimeNano(t *testing.T) {
	// We have to make sure that timeKeywordFunc always returns the correct
	// formatting with regards to the nanotime.

	for _, test := range []struct {
		sec, nsec int64
	}{
		{1234, 123456789},
		{5555, 987654321},
		{1337, 100000000},
		{8888, 999999999},
		{144123582122, 1},
		{857125628319, 0},
	} {
		mtime := time.Unix(test.sec, test.nsec)
		expected := fmt.Sprintf("time=%d.%9.9d", test.sec, test.nsec)
		got, err := timeKeywordFunc("", fakeFileInfo{
			mtime: mtime,
		}, nil)
		if err != nil {
			t.Errorf("unexpected error while parsing '%q': %q", mtime, err)
		}
		if expected != got {
			t.Errorf("keyword didn't match, expected '%s' got '%s'", expected, got)
		}
	}
}

func TestKeywordsTimeTar(t *testing.T) {
	// tartimeKeywordFunc always has nsec = 0.

	for _, test := range []struct {
		sec, nsec int64
	}{
		{1234, 123456789},
		{5555, 987654321},
		{1337, 100000000},
		{8888, 999999999},
		{144123582122, 1},
		{857125628319, 0},
	} {
		mtime := time.Unix(test.sec, test.nsec)
		expected := fmt.Sprintf("tar_time=%d.%9.9d", test.sec, 0)
		got, err := tartimeKeywordFunc("", fakeFileInfo{
			mtime: mtime,
		}, nil)
		if err != nil {
			t.Errorf("unexpected error while parsing '%q': %q", mtime, err)
		}
		if expected != got {
			t.Errorf("keyword didn't match, expected '%s' got '%s'", expected, got)
		}
	}
}
