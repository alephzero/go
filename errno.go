package alephzero

/*
#cgo pkg-config: alephzero
#include <a0/errno.h>
*/
import "C"

import (
	"syscall"
)

// https://github.com/golang/go/issues/15980
var A0_OK C.errno_t = 0

func errorFrom(err C.errno_t) error {
	if err == A0_OK {
		return nil
	}
	return syscall.Errno(err)
}