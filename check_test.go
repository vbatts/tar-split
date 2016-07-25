package mtree

import (
	"bytes"
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

// https://github.com/vbatts/go-mtree/issues/8
func TestTimeComparison(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-time.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// This is the format of time from FreeBSD
	spec := `
/set type=file time=5.000000000
.               type=dir
    file       time=5.000000000
..
`

	fh, err := os.Create(filepath.Join(dir, "file"))
	if err != nil {
		t.Fatal(err)
	}
	// This is what mode we're checking for. Round integer of epoch seconds
	epoch := time.Unix(5, 0)
	if err := os.Chtimes(fh.Name(), epoch, epoch); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(dir, epoch, epoch); err != nil {
		t.Fatal(err)
	}
	if err := fh.Close(); err != nil {
		t.Error(err)
	}

	dh, err := ParseSpec(bytes.NewBufferString(spec))
	if err != nil {
		t.Fatal(err)
	}

	res, err := Check(dir, dh, nil)
	if err != nil {
		t.Error(err)
	}
	if len(res.Failures) > 0 {
		t.Fatal(res.Failures)
	}
}

func TestTarTime(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-tar-time.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// This is the format of time from FreeBSD
	spec := `
/set type=file time=5.454353132
.               type=dir time=5.123456789
    file       time=5.911134111
..
`

	fh, err := os.Create(filepath.Join(dir, "file"))
	if err != nil {
		t.Fatal(err)
	}
	// This is what mode we're checking for. Round integer of epoch seconds
	epoch := time.Unix(5, 0)
	if err := os.Chtimes(fh.Name(), epoch, epoch); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(dir, epoch, epoch); err != nil {
		t.Fatal(err)
	}
	if err := fh.Close(); err != nil {
		t.Error(err)
	}

	dh, err := ParseSpec(bytes.NewBufferString(spec))
	if err != nil {
		t.Fatal(err)
	}

	// make sure "time" keyword works
	_, err = Check(dir, dh, DefaultKeywords)
	if err != nil {
		t.Error(err)
	}

	// make sure tar_time wins
	res, err := Check(dir, dh, append(DefaultKeywords, "tar_time"))
	if err != nil {
		t.Error(err)
	}
	if len(res.Failures) > 0 {
		t.Fatal(res.Failures)
	}
}

func TestIgnoreComments(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-comments.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// This is the format of time from FreeBSD
	spec := `
/set type=file time=5.000000000
.               type=dir
    file1       time=5.000000000
..
`

	fh, err := os.Create(filepath.Join(dir, "file1"))
	if err != nil {
		t.Fatal(err)
	}
	// This is what mode we're checking for. Round integer of epoch seconds
	epoch := time.Unix(5, 0)
	if err := os.Chtimes(fh.Name(), epoch, epoch); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(dir, epoch, epoch); err != nil {
		t.Fatal(err)
	}
	if err := fh.Close(); err != nil {
		t.Error(err)
	}

	dh, err := ParseSpec(bytes.NewBufferString(spec))
	if err != nil {
		t.Fatal(err)
	}

	res, err := Check(dir, dh, nil)
	if err != nil {
		t.Error(err)
	}

	if len(res.Failures) > 0 {
		t.Fatal(res.Failures)
	}

	// now change the spec to a comment that looks like an actual Entry but has
	// whitespace in front of it
	spec = `
/set type=file time=5.000000000
.               type=dir
    file1       time=5.000000000
	#file2 		time=5.000000000
..
`
	dh, err = ParseSpec(bytes.NewBufferString(spec))

	res, err = Check(dir, dh, nil)
	if err != nil {
		t.Error(err)
	}

	if len(res.Failures) > 0 {
		t.Fatal(res.Failures)
	}
}
