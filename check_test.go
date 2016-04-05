package mtree

import "testing"

func TestCheck(t *testing.T) {
	dh, err := Walk(".", nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	res, err := Check(".", dh)
	if err != nil {
		t.Fatal(err)
	}

	if len(res.Failures) > 0 {
		t.Errorf("%#v", res)
	}
}

// TODO make a directory, walk it, check it, modify it and ensure it fails
