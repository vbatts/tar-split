package mtree

// #include "vis.h"
import "C"
import "fmt"

func Vis(str string) (string, error) {
	dst := new(C.char)
	ret := C.strvis(dst, C.CString(str), C.VIS_WHITE|C.VIS_OCTAL|C.VIS_GLOB)
	if ret == 0 {
		return "", fmt.Errorf("failed to encode string")
	}
	return C.GoString(dst), nil
}
