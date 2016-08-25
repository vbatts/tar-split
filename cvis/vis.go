package cvis

// #include "vis.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"math"
	"unsafe"
)

// Vis is a wrapper around the C implementation
func Vis(src string, flags int) (string, error) {
	// dst needs to be 4 times the length of str, must check appropriate size
	if uint32(len(src)*4+1) >= math.MaxUint32/4 {
		return "", fmt.Errorf("failed to encode: %q", src)
	}
	dst := string(make([]byte, 4*len(src)+1))
	cDst, cSrc := C.CString(dst), C.CString(src)
	defer C.free(unsafe.Pointer(cDst))
	defer C.free(unsafe.Pointer(cSrc))
	C.strvis(cDst, cSrc, C.int(flags))

	return C.GoString(cDst), nil
}

// DefaultVisFlags are the common flags used in mtree string encoding
var DefaultVisFlags = C.VIS_WHITE | C.VIS_OCTAL | C.VIS_GLOB
