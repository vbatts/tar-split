package cvis

// #include "vis.h"
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"
)

// Unvis decodes the Vis() string encoding
func Unvis(src string) (string, error) {
	cDst, cSrc := C.CString(string(make([]byte, len(src)+1))), C.CString(src)
	defer C.free(unsafe.Pointer(cDst))
	defer C.free(unsafe.Pointer(cSrc))
	ret := C.strunvis(cDst, cSrc)
	// TODO(vbatts) this needs to be confirmed against UnvisError
	if ret == -1 {
		return "", fmt.Errorf("failed to decode: %q", src)
	}
	return C.GoString(cDst), nil
}
