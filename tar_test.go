package mtree

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func ExampleStreamer() {
	fh, err := os.Open("./testdata/test.tar")
	if err != nil {
		// handle error ...
	}
	str := NewTarStreamer(fh, nil)
	if err := extractTar("/tmp/dir", str); err != nil {
		// handle error ...
	}

	dh, err := str.Hierarchy()
	if err != nil {
		// handle error ...
	}

	res, err := Check("/tmp/dir/", dh, nil)
	if err != nil {
		// handle error ...
	}
	if len(res.Failures) > 0 {
		// handle validation issue ...
	}
}
func extractTar(root string, tr io.Reader) error {
	return nil
}

func TestTar(t *testing.T) {
	/*
		data, err := makeTarStream()
		if err != nil {
			t.Fatal(err)
		}
		buf := bytes.NewBuffer(data)
		str := NewTarStreamer(buf, append(DefaultKeywords, "sha1"))
	*/
	/*
		// open empty folder and check size.
		fh, err := os.Open("./testdata/empty")
		if err != nil {
			t.Fatal(err)
		}
		log.Println(fh.Stat())
		fh.Close() */
	fh, err := os.Open("./testdata/test.tar")
	if err != nil {
		t.Fatal(err)
	}
	str := NewTarStreamer(fh, append(DefaultKeywords, "sha1"))

	if _, err := io.Copy(ioutil.Discard, str); err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if err := str.Close(); err != nil {
		t.Fatal(err)
	}
	defer fh.Close()

	// get DirectoryHierarcy struct from walking the tar archive
	tdh, err := str.Hierarchy()
	if err != nil {
		t.Fatal(err)
	}
	if tdh == nil {
		t.Fatal("expected a DirectoryHierarchy struct, but got nil")
	}

	fh, err = os.Create("./testdata/test.mtree")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("./testdata/test.mtree")

	// put output of tar walk into test.mtree
	_, err = tdh.WriteTo(fh)
	if err != nil {
		t.Fatal(err)
	}
	fh.Close()

	// now simulate gomtree -T testdata/test.tar -f testdata/test.mtree
	fh, err = os.Open("./testdata/test.mtree")
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()

	dh, err := ParseSpec(fh)
	if err != nil {
		t.Fatal(err)
	}

	res, err := TarCheck(tdh, dh, append(DefaultKeywords, "sha1"))

	if err != nil {
		t.Fatal(err)
	}

	// print any failures, and then call t.Fatal once all failures/extra/missing
	// are outputted
	if res != nil {
		errors := ""
		switch {
		case len(res.Failures) > 0:
			for _, f := range res.Failures {
				fmt.Printf("%s\n", f)
			}
			errors += "Keyword validation errors\n"
		case len(res.Missing) > 0:
			for _, m := range res.Missing {
				fmt.Printf("Missing file: %s\n", m.Path())
			}
			errors += "Missing files not expected for this test\n"
		case len(res.Extra) > 0:
			for _, e := range res.Extra {
				fmt.Printf("Extra file: %s\n", e.Path())
			}
			errors += "Extra files not expected for this test\n"
		}
		if errors != "" {
			t.Fatal(errors)
		}
	}
}

// minimal tar archive stream that mimics what is in ./testdata/test.tar
func makeTarStream() ([]byte, error) {
	buf := new(bytes.Buffer)

	// Create a new tar archive.
	tw := tar.NewWriter(buf)

	// Add some files to the archive.
	var files = []struct {
		Name, Body string
		Mode       int64
		Type       byte
		Xattrs     map[string]string
	}{
		{"x/", "", 0755, '5', nil},
		{"x/files", "howdy\n", 0644, '0', nil},
	}
	for _, file := range files {
		hdr := &tar.Header{
			Name:   file.Name,
			Mode:   file.Mode,
			Size:   int64(len(file.Body)),
			Xattrs: file.Xattrs,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}
		if len(file.Body) > 0 {
			if _, err := tw.Write([]byte(file.Body)); err != nil {
				return nil, err
			}
		}
	}
	// Make sure to check the error on Close.
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
