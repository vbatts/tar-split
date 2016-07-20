package mtree

// #include "vis.h"
import "C"
import "fmt"

func Unvis(str string) (string, error) {
	dst := new(C.char)
	ret := C.strunvis(dst, C.CString(str))
	if ret == 0 {
		return "", fmt.Errorf("failed to encode string")
	}

	return C.GoString(dst), nil
}
