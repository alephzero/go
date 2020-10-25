package alephzero

/*
#cgo pkg-config: alephzero
#include "alloc_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"sync"
	"unsafe"
)

var (
	allocMutex    = sync.Mutex{}
	allocRegistry = make(map[uintptr]func(C.size_t, *C.a0_buf_t) C.errno_t)
	nextAllocId   uintptr
)

//export a0go_alloc
func a0go_alloc(id unsafe.Pointer, size C.size_t, out *C.a0_buf_t) C.errno_t {
	allocMutex.Lock()
	fn := allocRegistry[uintptr(id)]
	allocMutex.Unlock()
	return fn(size, out)
}

func registerAlloc(fn func(C.size_t, *C.a0_buf_t) C.errno_t) (id uintptr) {
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
