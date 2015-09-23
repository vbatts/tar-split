package common

import "testing"

func TestStringValidation(t *testing.T) {
	cases := []struct {
		value  string
		result bool
	}{
		{"aä\uFFFD本☺", false},
		{"aä本☺", true},
	}

	for _, c := range cases {
		if got := IsValidUtf8String(c.value); got != c.result {
			t.Errorf("string %q - expected %v, got %v", c.value, c.result, got)
		}
	}
}
func TestBytesValidation(t *testing.T) {
	cases := []struct {
		value  []byte
		result bool
	}{
		{[]byte{0xE4}, false},
		{[]byte("aä本☺"), true},
	}

	for _, c := range cases {
		if got := IsValidUtf8Btyes(c.value); got != c.result {
			t.Errorf("bytes %q - expected %v, got %v", c.value, c.result, got)
		}
	}
}
