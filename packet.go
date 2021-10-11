package alephzero

/*
#cgo pkg-config: alephzero
#include "packet_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"unsafe"
)

type Packet struct {
	Id      string
	Headers map[string][]string
	Payload []byte
}

func NewPacket(headers map[string][]string, payload []byte) (pkt Packet) {
	cPkt := C.a0_packet_t{}
	C.a0_packet_init(&cPkt)

	idCStr := ([C.A0_UUID_SIZE]C.char)(cPkt.id)
	pkt.Id = C.GoStringN(&idCStr[0], C.A0_UUID_SIZE-1)
	pkt.Headers = headers
	pkt.Payload = payload
	return
}

func PacketDepKey() string {
	return C.GoString(C.A0_PACKET_DEP_KEY)
}

func packetFromC(cPkt C.a0_packet_t) (pkt Packet) {
	pkt.Id = string(C.GoBytes(unsafe.Pointer(&cPkt.id), 36))
	pkt.Headers = make(map[string][]string)

	hdrIter := &C.a0_packet_header_iterator_t{}
	C.a0_packet_header_iterator_init(hdrIter, &cPkt)

	hdr := &C.a0_packet_header_t{}
	for C.a0_packet_header_iterator_next(hdrIter, hdr) == 0 {
		hdrKey := C.GoString(hdr.key)
		hdrVal := C.GoString(hdr.val)
		pkt.Headers[hdrKey] = append(pkt.Headers[hdrKey], hdrVal)
	}

	pkt.Payload = (*[1 << 30]byte)(unsafe.Pointer(cPkt.payload.data))[:int(cPkt.payload.size):int(cPkt.payload.size)]
	return
}

func (p Packet) c() (cPkt C.a0_packet_t) {
	for i := 0; i < 36; i++ {
		cPkt.id[i] = (C.char)(p.Id[i])
	}
	cPkt.payload.size = C.size_t(len(p.Payload))
	if cPkt.payload.size > 0 {
		cPkt.payload.data = (*C.uint8_t)(&p.Payload[0])
	}

	numHeaders := 0
	for _, v := range p.Headers {
		numHeaders += len(v)
	}

	cPkt.headers_block.size = C.size_t(numHeaders)
	cPkt.headers_block.headers = (*C.a0_packet_header_t)(C.malloc(C.size_t(numHeaders) * C.size_t(unsafe.Sizeof(C.a0_packet_header_t{}))))

	cHdrs := (*[1 << 30]C.a0_packet_header_t)(unsafe.Pointer(cPkt.headers_block.headers))[:int(cPkt.headers_block.size):int(cPkt.headers_block.size)]

	i := 0
	for k, vs := range p.Headers {
		cHdrs[i].key = C.CString(k)
		for _, v := range vs {
			cHdrs[i].val = C.CString(v)
			i++
		}
	}

	return
}

func freeCPacket(cPkt C.a0_packet_t) {
	numHdrs := int(cPkt.headers_block.size)
	cHdrs := (*[1 << 30]C.a0_packet_header_t)(unsafe.Pointer(cPkt.headers_block.headers))[:numHdrs:numHdrs]
	for i := 0; i < numHdrs; i++ {
		C.free(unsafe.Pointer(cHdrs[i].key))
		C.free(unsafe.Pointer(cHdrs[i].val))
	}
	C.free(unsafe.Pointer(cPkt.headers_block.headers))
}

//export a0go_packet_callback
func a0go_packet_callback(id unsafe.Pointer, c C.a0_packet_t) {
	registry.Get(uintptr(id)).(func(C.a0_packet_t))(c)
}

//export a0go_packet_id_callback
func a0go_packet_id_callback(id unsafe.Pointer, c *C.char) {
	registry.Get(uintptr(id)).(func(*C.char))(c)
}
