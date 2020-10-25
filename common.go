package alephzero

/*
#cgo pkg-config: alephzero
#include "common_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"sync"
	"unsafe"
)

// dst is a void**. src is a void*;
// This function is *dst = src.
func cpPtr(dst unsafe.Pointer, src unsafe.Pointer) {
	C.a0go_copy_ptr(C.uintptr_t(uintptr(dst)), C.uintptr_t(uintptr(src)))
}

func wrapGoMem(goMem []byte, out *C.a0_buf_t) {
	out.size = C.size_t(len(goMem))
	if out.size > 0 {
		cpPtr(unsafe.Pointer(&out.ptr), unsafe.Pointer(&goMem[0]))
	}
	return
}

//////////////
// Callback //
//////////////

var (
	callbackMutex    = sync.Mutex{}
	callbackRegistry = make(map[uintptr]func())
	nextCallbackId   uintptr
)

//export a0go_callback
func a0go_callback(id unsafe.Pointer) {
	callbackMutex.Lock()
	fn := callbackRegistry[uintptr(id)]
	callbackMutex.Unlock()
	fn()
}

func registerCallback(fn func()) (id uintptr) {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	id = nextCallbackId
	nextCallbackId++
	callbackRegistry[id] = fn
	return
}

func unregisterCallback(id uintptr) {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	delete(callbackRegistry, id)
}
