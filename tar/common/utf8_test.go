package common

import "testing"

func TestStringValidation(t *testing.T) {
	cases := []struct {
		value  string
		result bool
		offset int
	}{
		{"aä\uFFFD本☺", false, 3},
		{"aä本☺", true, -1},
	}

	for _, c := range cases {
		if i := InvalidUtf8Index([]byte(c.value)); i != c.offset {
			t.Errorf("string %q - offset expected %d, got %d", c.value, c.offset, i)
		}
		if got := IsValidUtf8String(c.value); got != c.result {
			t.Errorf("string %q - expected %v, got %v", c.value, c.result, got)
		}
	}
}

func TestBytesValidation(t *testing.T) {
	cases := []struct {
		value  []byte
		result bool
		offset int
	}{
		{[]byte{0xE4}, false, 0},
		{[]byte("aä本☺"), true, -1},
	}

	for _, c := range cases {
		if i := InvalidUtf8Index(c.value); i != c.offset {
			t.Errorf("bytes %q - offset expected %d, got %d", c.value, c.offset, i)
		}
		if got := IsValidUtf8Btyes(c.value); got != c.result {
			t.Errorf("bytes %q - expected %v, got %v", c.value, c.result, got)
		}
	}
}
