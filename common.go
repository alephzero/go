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
	allocRegistry     = make(map[uintptr]func(C.size_t, *C.a0_buf_t))
	allocRegistryLock = sync.Mutex{}
	nextAllocId       uintptr
)

//export a0go_alloc
func a0go_alloc(id unsafe.Pointer, size C.size_t, out *C.a0_buf_t) {
	// TODO: Should this be a reader lock?
	allocRegistryLock.Lock()
	defer allocRegistryLock.Unlock()
	allocRegistry[uintptr(id)](size, out)
}

func registerAlloc(fn func(C.size_t, *C.a0_buf_t)) (id uintptr) {
	allocRegistryLock.Lock()
	defer allocRegistryLock.Unlock()
	id = nextAllocId
	nextAllocId++
	allocRegistry[id] = fn
	return
}

func unregisterAlloc(id uintptr) {
	allocRegistryLock.Lock()
	defer allocRegistryLock.Unlock()
	delete(allocRegistry, id)
}

//////////////
// Callback //
//////////////

var (
	callbackRegistry     = make(map[uintptr]func())
	callbackRegistryLock = sync.Mutex{}
	nextCallbackId       uintptr
)

//export a0go_callback
func a0go_callback(id unsafe.Pointer) {
	// TODO: Should this be a reader lock?
	callbackRegistryLock.Lock()
	defer callbackRegistryLock.Unlock()
	callbackRegistry[uintptr(id)]()
}

func registerCallback(fn func()) (id uintptr) {
	callbackRegistryLock.Lock()
	defer callbackRegistryLock.Unlock()
	id = nextCallbackId
	nextCallbackId++
	callbackRegistry[id] = fn
	return
}

func unregisterCallback(id uintptr) {
	callbackRegistryLock.Lock()
	defer callbackRegistryLock.Unlock()
	delete(callbackRegistry, id)
}
