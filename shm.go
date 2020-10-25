package alephzero

/*
#cgo pkg-config: alephzero
#include <a0/legacy_arena.h>
#include <stdlib.h>  // free
*/
import "C"

import (
	"unsafe"
)

type ShmOptions struct {
	Size   int
	Resize bool
}

type Shm struct {
	c C.a0_shm_t
}

func ShmOpen(path string, opts *ShmOptions) (shm Shm, err error) {
	pathCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathCStr))

	var cOpts C.a0_shm_options_t = C.A0_SHM_OPTIONS_DEFAULT
	if opts != nil {
		cOpts.size = C.off_t(opts.Size)
		cOpts.resize = C.bool(opts.Resize)
	}
	err = errorFrom(C.a0_shm_open(pathCStr, &cOpts, &shm.c))
	return
}

func ShmUnlink(path string) error {
	pathCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathCStr))

	return errorFrom(C.a0_shm_unlink(pathCStr))
}

func (shm *Shm) Close() error {
	return errorFrom(C.a0_shm_close(&shm.c))
}

func (shm *Shm) Path() string {
	return C.GoString(shm.c.path)
}
