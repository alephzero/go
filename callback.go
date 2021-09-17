package alephzero

/*
#cgo pkg-config: alephzero
#include "callback_adapter.h"
*/
import "C"

import (
	"sync"
	"unsafe"
)

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
