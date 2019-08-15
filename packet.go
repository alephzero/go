package alephzero

/*
#cgo pkg-config: alephzero
#include "packet_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"sync"
	"unsafe"
)

type PacketHeader struct {
	Key, Val []byte
}

type Packet struct {
	c     C.a0_packet_t
	goMem []byte
}

func NewPacket(hdrs []PacketHeader, payload []byte) (pkt Packet, err error) {
	// Package headers.
	var cHdrs unsafe.Pointer
	if len(hdrs) > 0 {
		cHdrs = C.malloc(C.size_t(len(hdrs) * int(unsafe.Sizeof(C.a0_packet_header_t{}))))
		defer C.free(cHdrs)

		for i, hdr := range hdrs {
			cHdr := &(*[1 << 30]C.a0_packet_header_t)(cHdrs)[i]

			cHdr.key = cBufFrom(hdr.Key)
			cHdr.val = cBufFrom(hdr.Val)
		}
	}

	// Package payload.
	cPayload := cBufFrom(payload)

	// Create allocator.
	allocId := registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		out.size = size
		if size > 0 {
			pkt.goMem = make([]byte, int(size))
			out.ptr = (*C.uint8_t)(&pkt.goMem[0])
		}
	})
	defer unregisterAlloc(allocId)

	// Compile packet.
	err = errorFrom(C.a0go_packet_build(
		C.size_t(len(hdrs)),
		(*C.a0_packet_header_t)(cHdrs),
		cPayload,
		C.uintptr_t(allocId),
		&pkt.c))

	return
}

func PacketIdKey() []byte {
	return goBufFrom(C.a0_packet_id_key())
}

func PacketDepKey() []byte {
	return goBufFrom(C.a0_packet_dep_key())
}

func (p *Packet) Bytes() ([]byte, error) {
	return p.goMem, nil
}

func (p *Packet) NumHeaders() (cnt int, err error) {
	var ucnt C.size_t
	err = errorFrom(C.a0_packet_num_headers(p.c, &ucnt))
	if err != nil {
		return
	}
	cnt = int(ucnt)
	return
}

func (p *Packet) Header(idx int) (hdr PacketHeader, err error) {
	var cHdr C.a0_packet_header_t
	if err = errorFrom(C.a0_packet_header(p.c, C.size_t(idx), &cHdr)); err != nil {
		return
	}
	hdr.Key = goBufFrom(cHdr.key)
	hdr.Key = goBufFrom(cHdr.val)
	return
}

func (p *Packet) FindHeader(key []byte) (val []byte, err error) {
	var cVal C.a0_buf_t
	if err = errorFrom(C.a0_packet_find_header(p.c, cBufFrom(key), &cVal)); err != nil {
		return
	}
	val = goBufFrom(cVal)
	return
}

func (p *Packet) Payload() (payload []byte, err error) {
	var cBuf C.a0_buf_t
	if err = errorFrom(C.a0_packet_payload(p.c, &cBuf)); err != nil {
		return
	}
	payload = goBufFrom(cBuf)
	return
}

func (p *Packet) Id() (val []byte, err error) {
	var cVal C.a0_buf_t
	if err = errorFrom(C.a0_packet_id(p.c, &cVal)); err != nil {
		return
	}
	val = goBufFrom(cVal)
	return
}

var (
	packetCallbackMutex    = sync.Mutex{}
	packetCallbackRegistry = make(map[uintptr]func(C.a0_packet_t))
	nextPacketCallbackId   uintptr
)

//export a0go_packet_callback
func a0go_packet_callback(id unsafe.Pointer, c C.a0_packet_t) {
	packetCallbackMutex.Lock()
	fn := packetCallbackRegistry[uintptr(id)]
	packetCallbackMutex.Unlock()
	fn(c)
}

func registerPacketCallback(fn func(C.a0_packet_t)) (id uintptr) {
	packetCallbackMutex.Lock()
	defer packetCallbackMutex.Unlock()
	id = nextPacketCallbackId
	nextPacketCallbackId++
	packetCallbackRegistry[id] = fn
	return
}

func unregisterPacketCallback(id uintptr) {
	packetCallbackMutex.Lock()
	defer packetCallbackMutex.Unlock()
	delete(packetCallbackRegistry, id)
}
