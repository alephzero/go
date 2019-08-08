package alephzero

// #cgo pkg-config: alephzero
// #include "common_adapter.h"
import "C"

import (
	"syscall"
	"unsafe"
)

var (
	// TODO: make thread safe.
	allocRegistry = make(map[int]func(C.size_t, *C.a0_buf_t))
	nextAllocId   int
)

func errorFrom(err C.errno_t) error {
	if err == 0 {
		return nil
	}
	return syscall.Errno(err)
}

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
