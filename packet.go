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
	Key, Val string
}

type Packet []byte

func packetFromC(cPkt C.a0_packet_t) Packet {
	return C.GoBytes(unsafe.Pointer(cPkt.ptr), C.int(cPkt.size))
}

func NewPacket(hdrs []PacketHeader, payload []byte) (pkt Packet, err error) {
	// Package headers.
	var cHdrs unsafe.Pointer
	if len(hdrs) > 0 {
		cHdrs = C.malloc(C.size_t(len(hdrs) * int(unsafe.Sizeof(C.a0_packet_header_t{}))))
		defer C.free(cHdrs)

		for i, hdr := range hdrs {
			cHdr := &(*[1 << 30]C.a0_packet_header_t)(cHdrs)[i]

			cKey := C.CString(hdr.Key)
			defer C.free(unsafe.Pointer(cKey))
			cHdr.key = cKey

			cVal := C.CString(hdr.Val)
			defer C.free(unsafe.Pointer(cVal))
			cHdr.val = cVal
		}
	}

	// Package payload.
	cPayload := cBufFrom(payload)

	// Create allocator.
	allocId := registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		out.size = size
		if size > 0 {
			pkt = make([]byte, int(size))
			out.ptr = (*C.uint8_t)(&pkt[0])
		}
	})
	defer unregisterAlloc(allocId)

	// Compile packet.
	err = errorFrom(C.a0go_packet_build(
		C.size_t(len(hdrs)),
		(*C.a0_packet_header_t)(cHdrs),
		cPayload,
		C.uintptr_t(allocId),
		nil))

	return
}

func (p Packet) C() (cPkt C.a0_packet_t) {
	cPkt.size = C.size_t(len(p))
	if cPkt.size > 0 {
		cPkt.ptr = (*C.uint8_t)(&p[0])
	}
	return
}

func PacketIdKey() string {
	return C.GoString(C.a0_packet_id_key())
}

func PacketDepKey() string {
	return C.GoString(C.a0_packet_dep_key())
}

func (p Packet) NumHeaders() (cnt int, err error) {
	var ucnt C.size_t
	err = errorFrom(C.a0_packet_num_headers(p.C(), &ucnt))
	if err != nil {
		return
	}
	cnt = int(ucnt)
	return
}

func (p Packet) Header(idx int) (hdr PacketHeader, err error) {
	var cHdr C.a0_packet_header_t
	if err = errorFrom(C.a0_packet_header(p.C(), C.size_t(idx), &cHdr)); err != nil {
		return
	}
	hdr.Key = C.GoString(cHdr.key)
	hdr.Val = C.GoString(cHdr.val)
	return
}

func (p Packet) Headers() (hdrs []PacketHeader, err error) {
	n, err := p.NumHeaders()
	if err != nil {
		return
	}
	for i := 0; i < n; i++ {
		var hdr PacketHeader
		hdr, err = p.Header(i)
		if err != nil {
			return
		}
		hdrs = append(hdrs, hdr)
	}
	return
}

func (p Packet) Payload() (payload []byte, err error) {
	var cBuf C.a0_buf_t
	if err = errorFrom(C.a0_packet_payload(p.C(), &cBuf)); err != nil {
		return
	}
	payload = goBufFrom(cBuf)
	return
}

func (p Packet) Id() (val string, err error) {
	var cVal C.a0_packet_id_t
	if err = errorFrom(C.a0_packet_id(p.C(), &cVal)); err != nil {
		return
	}
	// TODO: There must be a better way!
	var goBytes []byte
	for i := 0; i < C.A0_PACKET_ID_SIZE - 1; i++ {
		goBytes = append(goBytes, byte(cVal[i]))
	}
	val = string(goBytes)
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

var (
	packetIdCallbackMutex    = sync.Mutex{}
	packetIdCallbackRegistry = make(map[uintptr]func(*C.char))
	nextPacketIdCallbackId   uintptr
)

//export a0go_packet_id_callback
func a0go_packet_id_callback(id unsafe.Pointer, c *C.char) {
	packetIdCallbackMutex.Lock()
	fn := packetIdCallbackRegistry[uintptr(id)]
	packetIdCallbackMutex.Unlock()
	fn(c)
}

func registerPacketIdCallback(fn func(*C.char)) (id uintptr) {
	packetIdCallbackMutex.Lock()
	defer packetIdCallbackMutex.Unlock()
	id = nextPacketIdCallbackId
	nextPacketIdCallbackId++
	packetIdCallbackRegistry[id] = fn
	return
}

func unregisterPacketIdCallback(id uintptr) {
	packetIdCallbackMutex.Lock()
	defer packetIdCallbackMutex.Unlock()
	delete(packetIdCallbackRegistry, id)
}
