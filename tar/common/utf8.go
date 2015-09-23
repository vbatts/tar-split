package common

// IsValidUtf8String checks for in valid UTF-8 characters
func IsValidUtf8String(s string) bool {
	return InvalidUtf8Index([]byte(s)) == -1
}

// IsValidUtf8Btyes checks for in valid UTF-8 characters
func IsValidUtf8Btyes(b []byte) bool {
	return InvalidUtf8Index(b) == -1
}

// InvalidUtf8Index returns the offset of the first invalid UTF-8 character.
// Default is to return -1 for a wholly valid sequence.
func InvalidUtf8Index(b []byte) int {
	for i, r := range string(b) {
		if int(r) == 0xfffd {
			return i
		}
	}
	return -1
}
