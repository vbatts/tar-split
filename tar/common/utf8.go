package common

// IsValidUtf8String checks for in valid UTF-8 characters
func IsValidUtf8String(s string) bool {
	for _, r := range s {
		if int(r) == 0xfffd {
			return false
		}
	}
	return true
}

// IsValidUtf8Btyes checks for in valid UTF-8 characters
func IsValidUtf8Btyes(b []byte) bool {
	for _, r := range string(b) {
		if int(r) == 0xfffd {
			return false
		}
	}
	return true
}
