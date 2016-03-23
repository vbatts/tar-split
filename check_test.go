package mtree

import (
	"log"
	"testing"
)

func TestCheck(t *testing.T) {
	dh, err := Walk(".", nil, append(DefaultKeywords, "sha1"))
	if err != nil {
		t.Fatal(err)
	}

	res, err := Check(".", dh)
	if err != nil {
		t.Fatal(err)
	}
	//log.Fatalf("%#v", dh)
	log.Fatalf("%#v", res)
}
