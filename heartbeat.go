package alephzero

/*
#cgo pkg-config: alephzero
#include <a0/heartbeat.h>
*/
import "C"

type HeartbeatOptions struct {
	Freq float64
}

type Heartbeat struct {
	c C.a0_heartbeat_t
}

func NewHeartbeat(file File, opts *HeartbeatOptions) (h Heartbeat, err error) {
	var cOpts C.a0_heartbeat_options_t = C.A0_HEARTBEAT_OPTIONS_DEFAULT
	if opts != nil {
		cOpts.freq = C.double(opts.Freq)
	}
	err = errorFrom(C.a0_heartbeat_init(&h.c, file.c.arena, &cOpts))
	return
}

func (h *Heartbeat) Close() error {
	return errorFrom(C.a0_heartbeat_close(&h.c))
}
