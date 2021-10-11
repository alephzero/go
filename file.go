package alephzero

/*
#cgo pkg-config: alephzero
#include <a0/file.h>
#include <stdlib.h>  // free
*/
import "C"

import (
	"unsafe"
)

type FileOptions struct {
	CreateOptions struct {
		Size    int
		Mode    int
		DirMode int
	}

	OpenOptions struct {
		ArenaMode ArenaMode
	}
}

func (opts FileOptions) toC() (c C.a0_file_options_t) {
	c.create_options.size = C.long(opts.CreateOptions.Size)
	c.create_options.mode = C.uint(opts.CreateOptions.Mode)
	c.create_options.dir_mode = C.uint(opts.CreateOptions.DirMode)
	c.open_options.arena_mode = C.a0_arena_mode_t(opts.OpenOptions.ArenaMode)
	return
}

func (opts *FileOptions) fromC(c C.a0_file_options_t) {
	opts.CreateOptions.Size = int(c.create_options.size)
	opts.CreateOptions.Mode = int(c.create_options.mode)
	opts.CreateOptions.DirMode = int(c.create_options.dir_mode)
	opts.OpenOptions.ArenaMode = ArenaMode(c.open_options.arena_mode)
	return
}

func MakeDefaultFileOptions() (opts FileOptions) {
	opts.fromC(C.A0_FILE_OPTIONS_DEFAULT)
	return
}

type File struct {
	c C.a0_file_t
}

func FileOpen(path string, opts *FileOptions) (file File, err error) {
	pathCStr := C.CString(path)
	defer C.free(unsafe.Pointer(pathCStr))

	if opts == nil {
		defaultOpts := MakeDefaultFileOptions()
		opts = &defaultOpts
	}
	cOpts := opts.toC()
	err = errorFrom(C.a0_file_open(pathCStr, &cOpts, &file.c))
	return
}

func (file *File) Close() error {
	return errorFrom(C.a0_file_close(&file.c))
}

func (file File) Path() string {
	return C.GoString(file.c.path)
}

func (file File) Fd() int {
	return int(file.c.fd)
}

// TODO: Stat

func (file File) Arena() Arena {
	return Arena{file.c.arena}
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
