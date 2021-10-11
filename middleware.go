package alephzero

/*
#cgo pkg-config: alephzero
#include <a0/middleware.h>
*/
import "C"

type Middleware struct {
	c C.a0_middleware_t
}

func AddTimeMonoHeader() Middleware {
	return Middleware{C.a0_add_time_mono_header()}
}

func AddTimeWallHeader() Middleware {
	return Middleware{C.a0_add_time_wall_header()}
}

func AddWriterIdHeader() Middleware {
	return Middleware{C.a0_add_writer_id_header()}
}

func AddWriterSeqHeader() Middleware {
	return Middleware{C.a0_add_writer_seq_header()}
}

func AddTransportSeqHeader() Middleware {
	return Middleware{C.a0_add_transport_seq_header()}
}

func AddStandardHeaders() Middleware {
	return Middleware{C.a0_add_standard_headers()}
}
