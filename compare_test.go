package mtree

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// simple walk of current directory, and imediately check it.
// may not be parallelizable.
func TestCompare(t *testing.T) {
	old, err := Walk(".", nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	new, err := Walk(".", nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	diffs, err := Compare(old, new, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(diffs) > 0 {
		t.Errorf("%#v", diffs)
	}
}

func TestCompareModified(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-compare-modified")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a bunch of objects.
	tmpfile := filepath.Join(dir, "tmpfile")
	if err := ioutil.WriteFile(tmpfile, []byte("some content here"), 0666); err != nil {
		t.Fatal(err)
	}

	tmpdir := filepath.Join(dir, "testdir")
	if err := os.Mkdir(tmpdir, 0755); err != nil {
		t.Fatal(err)
	}

	tmpsubfile := filepath.Join(tmpdir, "anotherfile")
	if err := ioutil.WriteFile(tmpsubfile, []byte("some different content"), 0666); err != nil {
		t.Fatal(err)
	}

	// Walk the current state.
	old, err := Walk(dir, nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	// Overwrite the content in one of the files.
	if err := ioutil.WriteFile(tmpsubfile, []byte("modified content"), 0666); err != nil {
		t.Fatal(err)
	}

	// Walk the new state.
	new, err := Walk(dir, nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	// Compare.
	diffs, err := Compare(old, new, nil)
	if err != nil {
		t.Fatal(err)
	}

	// 1 object
	if len(diffs) != 1 {
		t.Errorf("expected the diff length to be 1, got %d", len(diffs))
		for i, diff := range diffs {
			t.Logf("diff[%d] = %#v", i, diff)
		}
	}

	// These cannot fail.
	tmpfile, _ = filepath.Rel(dir, tmpfile)
	tmpdir, _ = filepath.Rel(dir, tmpdir)
	tmpsubfile, _ = filepath.Rel(dir, tmpsubfile)

	for _, diff := range diffs {
		switch diff.Path() {
		case tmpsubfile:
			if diff.Type() != Modified {
				t.Errorf("unexpected diff type for %s: %s", diff.Path(), diff.Type())
			}

			if diff.Diff() == nil {
				t.Errorf("expect to not get nil for .Diff()")
			}

			old := diff.Old()
			new := diff.New()
			if old == nil || new == nil {
				t.Errorf("expected to get (!nil, !nil) for (.Old, .New), got (%#v, %#v)", old, new)
			}
		default:
			t.Errorf("unexpected diff found: %#v", diff)
		}
	}
}

func TestCompareMissing(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-compare-missing")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a bunch of objects.
	tmpfile := filepath.Join(dir, "tmpfile")
	if err := ioutil.WriteFile(tmpfile, []byte("some content here"), 0666); err != nil {
		t.Fatal(err)
	}

	tmpdir := filepath.Join(dir, "testdir")
	if err := os.Mkdir(tmpdir, 0755); err != nil {
		t.Fatal(err)
	}

	tmpsubfile := filepath.Join(tmpdir, "anotherfile")
	if err := ioutil.WriteFile(tmpsubfile, []byte("some different content"), 0666); err != nil {
		t.Fatal(err)
	}

	// Walk the current state.
	old, err := Walk(dir, nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	// Delete the objects.
	if err := os.RemoveAll(tmpfile); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(tmpsubfile); err != nil {
		t.Fatal(err)
	}

	if err := os.RemoveAll(tmpdir); err != nil {
		t.Fatal(err)
	}

	// Walk the new state.
	new, err := Walk(dir, nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	// Compare.
	diffs, err := Compare(old, new, nil)
	if err != nil {
		t.Fatal(err)
	}

	// 3 objects + the changes to '.'
	if len(diffs) != 4 {
		t.Errorf("expected the diff length to be 4, got %d", len(diffs))
		for i, diff := range diffs {
			t.Logf("diff[%d] = %#v", i, diff)
		}
	}

	// These cannot fail.
	tmpfile, _ = filepath.Rel(dir, tmpfile)
	tmpdir, _ = filepath.Rel(dir, tmpdir)
	tmpsubfile, _ = filepath.Rel(dir, tmpsubfile)

	for _, diff := range diffs {
		switch diff.Path() {
		case ".":
			// ignore these changes
		case tmpfile, tmpdir, tmpsubfile:
			if diff.Type() != Missing {
				t.Errorf("unexpected diff type for %s: %s", diff.Path(), diff.Type())
			}

			if diff.Diff() != nil {
				t.Errorf("expect to get nil for .Diff(), got %#v", diff.Diff())
			}

			old := diff.Old()
			new := diff.New()
			if old == nil || new != nil {
				t.Errorf("expected to get (!nil, nil) for (.Old, .New), got (%#v, %#v)", old, new)
			}
		default:
			t.Errorf("unexpected diff found: %#v", diff)
		}
	}
}

func TestCompareExtra(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-compare-extra")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Walk the current state.
	old, err := Walk(dir, nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	// Create a bunch of objects.
	tmpfile := filepath.Join(dir, "tmpfile")
	if err := ioutil.WriteFile(tmpfile, []byte("some content here"), 0666); err != nil {
		t.Fatal(err)
	}

	tmpdir := filepath.Join(dir, "testdir")
	if err := os.Mkdir(tmpdir, 0755); err != nil {
		t.Fatal(err)
	}

	tmpsubfile := filepath.Join(tmpdir, "anotherfile")
	if err := ioutil.WriteFile(tmpsubfile, []byte("some different content"), 0666); err != nil {
		t.Fatal(err)
	}

	// Walk the new state.
	new, err := Walk(dir, nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	// Compare.
	diffs, err := Compare(old, new, nil)
	if err != nil {
		t.Fatal(err)
	}

	// 3 objects + the changes to '.'
	if len(diffs) != 4 {
		t.Errorf("expected the diff length to be 4, got %d", len(diffs))
		for i, diff := range diffs {
			t.Logf("diff[%d] = %#v", i, diff)
		}
	}

	// These cannot fail.
	tmpfile, _ = filepath.Rel(dir, tmpfile)
	tmpdir, _ = filepath.Rel(dir, tmpdir)
	tmpsubfile, _ = filepath.Rel(dir, tmpsubfile)

	for _, diff := range diffs {
		switch diff.Path() {
		case ".":
			// ignore these changes
		case tmpfile, tmpdir, tmpsubfile:
			if diff.Type() != Extra {
				t.Errorf("unexpected diff type for %s: %s", diff.Path(), diff.Type())
			}

			if diff.Diff() != nil {
				t.Errorf("expect to get nil for .Diff(), got %#v", diff.Diff())
			}

			old := diff.Old()
			new := diff.New()
			if old != nil || new == nil {
				t.Errorf("expected to get (!nil, nil) for (.Old, .New), got (%#v, %#v)", old, new)
			}
		default:
			t.Errorf("unexpected diff found: %#v", diff)
		}
	}
}

func TestCompareKeys(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-compare-keys")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// Create a bunch of objects.
	tmpfile := filepath.Join(dir, "tmpfile")
	if err := ioutil.WriteFile(tmpfile, []byte("some content here"), 0666); err != nil {
		t.Fatal(err)
	}

	tmpdir := filepath.Join(dir, "testdir")
	if err := os.Mkdir(tmpdir, 0755); err != nil {
		t.Fatal(err)
	}

	tmpsubfile := filepath.Join(tmpdir, "anotherfile")
	if err := ioutil.WriteFile(tmpsubfile, []byte("aaa"), 0666); err != nil {
		t.Fatal(err)
	}

	// Walk the current state.
	old, err := Walk(dir, nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	// Overwrite the content in one of the files, but without changing the size.
	if err := ioutil.WriteFile(tmpsubfile, []byte("bbb"), 0666); err != nil {
		t.Fatal(err)
	}

	// Walk the new state.
	new, err := Walk(dir, nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	// Compare.
	diffs, err := Compare(old, new, []string{"size"})
	if err != nil {
		t.Fatal(err)
	}

	// 0 objects
	if len(diffs) != 0 {
		t.Errorf("expected the diff length to be 0, got %d", len(diffs))
		for i, diff := range diffs {
			t.Logf("diff[%d] = %#v", i, diff)
		}
	}
}

// TODO: Add test for Compare(...) between a tar and a regular dh (to make sure that tar_time is handled correctly).
