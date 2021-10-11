package alephzero

/*
#cgo pkg-config: alephzero
#include "callback_adapter.h"
*/
import "C"

import (
	"unsafe"
)

//export a0go_callback
func a0go_callback(id unsafe.Pointer) {
	registry.Get(uintptr(id)).(func())()
}
