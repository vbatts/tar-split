package verify

import "testing"

func TestHeader(t *testing.T) {
	hdr := Header{}
	t.Fatalf("%#v", hdr)
}
