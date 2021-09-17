package alephzero

/*
#cgo pkg-config: alephzero
#include "reader_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

type ReaderInit int

const (
	INIT_OLDEST      ReaderInit = C.A0_INIT_OLDEST
	INIT_MOST_RECENT            = C.A0_INIT_MOST_RECENT
	INIT_AWAIT_NEW              = C.A0_INIT_AWAIT_NEW
)

type ReaderIter int

const (
	ITER_NEXT   ReaderIter = C.A0_ITER_NEXT
	ITER_NEWEST            = C.A0_ITER_NEWEST
)

type ReaderSync struct {
	c       C.a0_reader_sync_t
	allocId uintptr
	// Memory must survive between the alloc and Next.
	activePktSpace []byte
}

func NewReaderSync(arena Arena, init ReaderInit, iter ReaderIter) (rs *ReaderSync, err error) {
	rs = &ReaderSync{}

	rs.allocId = registry.Register(func(size C.size_t, out *C.a0_buf_t) C.a0_err_t {
		rs.activePktSpace = make([]byte, int(size))
		out.size = size
		if size > 0 {
			out.data = (*C.uint8_t)(&rs.activePktSpace[0])
		}
		return A0_OK
	})

	err = errorFrom(C.a0go_reader_sync_init(&rs.c, arena.c, C.uintptr_t(rs.allocId), C.a0_reader_init_t(init), C.a0_reader_iter_t(iter)))
	return
}

func (rs *ReaderSync) Close() (err error) {
	err = errorFrom(C.a0_reader_sync_close(&rs.c))
	if rs.allocId > 0 {
		registry.Unregister(rs.allocId)
	}
	return
}

func (rs *ReaderSync) HasNext() (hasNext bool, err error) {
	err = errorFrom(C.a0_reader_sync_has_next(&rs.c, (*C.bool)(&hasNext)))
	return
}

func (rs *ReaderSync) Next() (pkt Packet, err error) {
	var cPkt C.a0_packet_t
	err = errorFrom(C.a0_reader_sync_next(&rs.c, &cPkt))
	if err == nil {
		pkt = packetFromC(cPkt)
	}
	return
}

type Reader struct {
	c                C.a0_reader_t
	allocId          uintptr
	packetCallbackId uintptr
}

func NewReader(arena Arena, init ReaderInit, iter ReaderIter, callback func(Packet)) (r *Reader, err error) {
	r = &Reader{}

	var activePktSpace []byte
	r.allocId = registry.Register(func(size C.size_t, out *C.a0_buf_t) C.a0_err_t {
		activePktSpace = make([]byte, int(size))
		out.size = size
		if size > 0 {
			out.data = (*C.uint8_t)(&activePktSpace[0])
		}
		return A0_OK
	})

	r.packetCallbackId = registry.Register(func(cPkt C.a0_packet_t) {
		callback(packetFromC(cPkt))
	})

	err = errorFrom(C.a0go_reader_init(&r.c, arena.c, C.uintptr_t(r.allocId), C.a0_reader_init_t(init), C.a0_reader_iter_t(iter), C.uintptr_t(r.packetCallbackId)))
	return
}

func (r *Reader) Close() (err error) {
	err = errorFrom(C.a0_reader_close(&r.c))
	registry.Unregister(r.packetCallbackId)
	if r.allocId > 0 {
		registry.Unregister(r.allocId)
	}
	return
}

func ReaderReadOne(file File, init ReaderInit, flags int) (pkt Packet, err error) {
	var pktSpace []byte
	allocId := registry.Register(func(size C.size_t, out *C.a0_buf_t) C.a0_err_t {
		pktSpace = make([]byte, int(size))
		out.size = size
		if size > 0 {
			out.data = (*C.uint8_t)(&pktSpace[0])
		}
		return A0_OK
	})
	defer registry.Unregister(allocId)

	cPkt := C.a0_packet_t{}
	err = errorFrom(C.a0go_reader_read_one(file.c.arena, C.uintptr_t(allocId), C.a0_reader_init_t(init), C.int(flags), &cPkt))
	pkt = packetFromC(cPkt)
	copy(pkt.Payload, pkt.Payload)
	return
}
