package alephzero

/*
#cgo pkg-config: alephzero
#include <a0/arena.h>
*/
import "C"

import (
	"unsafe"
)

type ArenaMode C.a0_arena_mode_t

const (
	MODE_SHARED    ArenaMode = C.A0_ARENA_MODE_SHARED
	MODE_EXCLUSIVE ArenaMode = C.A0_ARENA_MODE_EXCLUSIVE
	MODE_READONLY  ArenaMode = C.A0_ARENA_MODE_READONLY
)

type Arena struct {
	c C.a0_arena_t
}

func (arena Arena) Buf() []byte {
	return C.GoBytes(unsafe.Pointer(arena.c.buf.data), C.int(arena.c.buf.size))
}

func (arena Arena) Mode() ArenaMode {
	return ArenaMode(arena.c.mode)
}
