package alephzero

/*
#cgo pkg-config: alephzero
#include <a0/arena.h>
#include <stdlib.h>  // free
*/
import "C"

import (
	"unsafe"
)

type fileOptions struct {
	CreateOptions struct {
		Size    int
		Mode    int
		DirMode int
	}

	OpenOptions struct {
		Readonly bool
	}
}

func (opts *fileOptions) toC() (c C.a0_file_options_t) {
	c.create_options.size = C.long(opts.CreateOptions.Size)
	c.create_options.mode = C.uint(opts.CreateOptions.Mode)
	c.create_options.dir_mode = C.uint(opts.CreateOptions.DirMode)
	c.open_options.readonly = C.bool(opts.OpenOptions.Readonly)
	return
}

func (opts *fileOptions) fromC(c C.a0_file_options_t) {
	opts.CreateOptions.Size = int(c.create_options.size)
	opts.CreateOptions.Mode = int(c.create_options.mode)
	opts.CreateOptions.DirMode = int(c.create_options.dir_mode)
	opts.OpenOptions.Readonly = bool(c.open_options.readonly)
	return
}

func MakeDefaultFileOptions() (opts *fileOptions) {
	opts = &fileOptions{}
	opts.fromC(C.A0_FILE_OPTIONS_DEFAULT)
	return
}

type File struct {
	c C.a0_file_t
}

func FileOpen(path string, opts *fileOptions) (file File, err error) {
	pathCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathCStr))

	if opts == nil {
		opts = MakeDefaultFileOptions()
	}
	cOpts := opts.toC()
	err = errorFrom(C.a0_file_open(pathCStr, &cOpts, &file.c))
	return
}

func (file *File) Close() error {
	return errorFrom(C.a0_file_close(&file.c))
}

func (file *File) Path() string {
	return C.GoString(file.c.path)
}

func (file *File) Fd() int {
	return int(file.c.fd)
}

func FileRemove(path string) error {
	pathCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathCStr))

	return errorFrom(C.a0_file_remove(pathCStr))
}

func FileRemoveAll(path string) error {
	pathCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathCStr))

	return errorFrom(C.a0_file_remove_all(pathCStr))
}
