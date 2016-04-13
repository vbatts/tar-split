package mtree

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// simple walk of current directory, and imediately check it.
// may not be parallelizable.
func TestCheck(t *testing.T) {
	dh, err := Walk(".", nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	res, err := Check(".", dh, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Failures) > 0 {
		t.Errorf("%#v", res)
	}
}

// make a directory, walk it, check it, modify the timestamp and ensure it fails.
// only check again for size and sha1, and ignore time, and ensure it passes
func TestCheckKeywords(t *testing.T) {
	content := []byte("I know half of you half as well as I ought to")
	dir, err := ioutil.TempDir("", "test-check-keywords")
	if err != nil {
		t.Fatal(err)
	}
	//defer os.RemoveAll(dir) // clean up

	tmpfn := filepath.Join(dir, "tmpfile")
	if err := ioutil.WriteFile(tmpfn, content, 0666); err != nil {
		t.Fatal(err)
	}

	// Walk this tempdir
	dh, err := Walk(dir, nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	// Check for sanity. This ought to pass.
	res, err := Check(dir, dh, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Failures) > 0 {
		t.Errorf("%#v", res)
	}

	// Touch a file, so the mtime changes.
	now := time.Now()
	if err := os.Chtimes(tmpfn, now, now); err != nil {
		t.Fatal(err)
	}

	// Check again. This ought to fail.
	res, err = Check(dir, dh, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Failures) == 0 {
		t.Errorf("expected to fail on changed mtimes, but did not")
	}

	// Check again, but only sha1 and mode. This ought to pass.
	res, err = Check(dir, dh, []string{"sha1", "mode"})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Failures) > 0 {
		t.Errorf("%#v", res)
	}
}

func ExampleCheck() {
	dh, err := Walk(".", nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		// handle error ...
	}

	res, err := Check(".", dh, nil)
	if err != nil {
		// handle error ...
	}
	if len(res.Failures) > 0 {
		// handle failed validity ...
	}
}
