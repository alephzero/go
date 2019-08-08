package alephzero

// #cgo pkg-config: alephzero
// #include "common_adapter.h"
import "C"

import (
	"syscall"
	"unsafe"
)

func errorFrom(err C.errno_t) error {
	if err == 0 {
		return nil
	}
	return syscall.Errno(err)
}

///////////
// Alloc //
///////////

var (
	// TODO: make thread safe.
	allocRegistry = make(map[int]func(C.size_t, *C.a0_buf_t))
	nextAllocId   int
)

//export a0go_alloc
func a0go_alloc(idPtr unsafe.Pointer, size C.size_t, out *C.a0_buf_t) {
	allocRegistry[int(*(*C.int)(idPtr))](size, out)
}

func registerAlloc(fn func(C.size_t, *C.a0_buf_t)) (id int) {
	id = nextAllocId
	nextAllocId++
	allocRegistry[id] = fn
	return
}

func unregisterAlloc(id int) {
	delete(allocRegistry, id)
}

//////////////
// Callback //
//////////////

var (
	// TODO: make thread safe.
	callbackRegistry = make(map[int]func())
	nextCallbackId   int
)

//export a0go_callback
func a0go_callback(idPtr unsafe.Pointer) {
	callbackRegistry[int(*(*C.int)(idPtr))]()
}

func registerCallback(fn func()) (id int) {
	id = nextCallbackId
	nextCallbackId++
	callbackRegistry[id] = fn
	return
}

func unregisterCallback(id int) {
	delete(callbackRegistry, id)
}
