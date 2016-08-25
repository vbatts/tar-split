// +build cgo,!govis

package cvis

import "testing"

// The resulting string of Vis output could potentially be four times longer than
// the original. Vis must handle this possibility.
func TestVisLength(t *testing.T) {
	testString := "All work and no play makes Jack a dull boy\n"
	for i := 0; i < 20; i++ {
		Vis(testString, DefaultVisFlags)
		testString = testString + testString
	}
}
