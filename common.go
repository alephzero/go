package alephzero

/*
#cgo pkg-config: alephzero
#include "common_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"sync"
	"syscall"
	"unsafe"
)

func errorFrom(err C.errno_t) error {
	if err == 0 {
		return nil
	}
	return syscall.Errno(err)
}

func goBufFrom(b C.a0_buf_t) []byte {
	return (*[1 << 30]byte)(unsafe.Pointer(b.ptr))[:int(b.size):int(b.size)]
}

func cBufFrom(b []byte) (out C.a0_buf_t) {
	out.size = C.size_t(len(b))
	if out.size > 0 {
		out.ptr = (*C.uint8_t)(&b[0])
	}
	return
}

///////////
// Alloc //
///////////

var (
	allocMutex    = sync.Mutex{}
	allocRegistry = make(map[uintptr]func(C.size_t, *C.a0_buf_t))
	nextAllocId   uintptr
)

//export a0go_alloc
func a0go_alloc(id unsafe.Pointer, size C.size_t, out *C.a0_buf_t) {
	allocMutex.Lock()
	fn := allocRegistry[uintptr(id)]
	allocMutex.Unlock()
	fn(size, out)
}

func registerAlloc(fn func(C.size_t, *C.a0_buf_t)) (id uintptr) {
	allocMutex.Lock()
	defer allocMutex.Unlock()
	id = nextAllocId
	nextAllocId++
	allocRegistry[id] = fn
	return
}

func unregisterAlloc(id uintptr) {
	allocMutex.Lock()
	defer allocMutex.Unlock()
	delete(allocRegistry, id)
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
