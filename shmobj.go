package alephzero

/*
#cgo pkg-config: alephzero
#include <a0/shmobj.h>
#include <stdlib.h>  // free
*/
import "C"

import (
	"unsafe"
)

type ShmObjOptions struct {
	Size int
}

type ShmObj struct {
	c C.a0_shmobj_t
}

func ShmOpen(path string, opts *ShmObjOptions) (so ShmObj, err error) {
	pathCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathCStr))

	var cOpts C.a0_shmobj_options_t
	if opts != nil {
		cOpts.size = C.off_t(opts.Size)
	}
	err = errorFrom(C.a0_shmobj_open(pathCStr, &cOpts, &so.c))
	return
}

func ShmUnlink(path string) error {
	pathCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathCStr))

	return errorFrom(C.a0_shmobj_unlink(pathCStr))
}

func (so *ShmObj) Close() error {
	return errorFrom(C.a0_shmobj_close(&so.c))
}
