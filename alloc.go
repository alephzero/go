package alephzero

/*
#cgo pkg-config: alephzero
#include "alloc_adapter.h"
*/
import "C"

import (
	"unsafe"
)

//export a0go_alloc
func a0go_alloc(id unsafe.Pointer, size C.size_t, out *C.a0_buf_t) C.a0_err_t {
	return registry.Get(uintptr(id)).(func(C.size_t, *C.a0_buf_t) C.a0_err_t)(size, out)
}
