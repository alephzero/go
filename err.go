package alephzero

/*
#cgo pkg-config: alephzero
#include <a0/err.h>
*/
import "C"

import (
	"errors"
)

const (
	A0_OK              = C.A0_OK
	A0_ERR_SYS         = C.A0_ERR_SYS
	A0_ERR_CUSTOM_MSG  = C.A0_ERR_CUSTOM_MSG
	A0_ERR_INVALID_ARG = C.A0_ERR_INVALID_ARG
	A0_ERR_RANGE       = C.A0_ERR_RANGE
	A0_ERR_AGAIN       = C.A0_ERR_AGAIN
	A0_ERR_ITER_DONE   = C.A0_ERR_ITER_DONE
	A0_ERR_NOT_FOUND   = C.A0_ERR_NOT_FOUND
	A0_ERR_FRAME_LARGE = C.A0_ERR_FRAME_LARGE
	A0_ERR_BAD_TOPIC   = C.A0_ERR_BAD_TOPIC
)

func errorFrom(err C.a0_err_t) error {
	if err == C.A0_OK {
		return nil
	}
	return errors.New(C.GoString(C.a0_strerror(err)))
}
