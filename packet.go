package alephzero

// #cgo pkg-config: alephzero
// #include "packet_adapter.h"
import "C"

import (
	"unsafe"
)

type PacketHeader struct {
	Key, Val []byte
}

type Packet struct {
	cPkt  C.a0_packet_t
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

			cHdr.key.size = C.size_t(len(hdr.Key))
			if cHdr.key.size > 0 {
				cHdr.key.ptr = (*C.uint8_t)(&hdr.Key[0])
			}

			cHdr.val.size = C.size_t(len(hdr.Val))
			if cHdr.val.size > 0 {
				cHdr.val.ptr = (*C.uint8_t)(&hdr.Val[0])
			}
		}
	}


	// Package payload.
	var cPayload C.a0_buf_t
	cPayload.size = C.size_t(len(payload))
	if cPayload.size > 0 {
		cPayload.ptr = (*C.uint8_t)(&payload[0])
	}

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
		&pkt.cPkt))

	return
}

func (p *Packet) Bytes() ([]byte, error) {
	return p.goMem, nil
}

func (p *Packet) NumHeaders() (cnt int, err error) {
	var ucnt C.size_t
	err = errorFrom(C.a0_packet_num_headers(p.cPkt, &ucnt))
	if err != nil {
		return
	}
	cnt = int(ucnt)
	return
}

func (p *Packet) Header(idx int) (hdr PacketHeader, err error) {
	var cHdr C.a0_packet_header_t

	if err = errorFrom(C.a0_packet_header(p.cPkt, C.size_t(idx), &cHdr)); err != nil {
		return
	}

	hdr.Key = (*[1 << 30]byte)(unsafe.Pointer(cHdr.key.ptr))[:int(cHdr.key.size):int(cHdr.key.size)]
	hdr.Val = (*[1 << 30]byte)(unsafe.Pointer(cHdr.val.ptr))[:int(cHdr.val.size):int(cHdr.val.size)]

	return
}

func (p *Packet) Payload() (payload []byte, err error) {
	var out C.a0_buf_t

	if err = errorFrom(C.a0_packet_payload(p.cPkt, &out)); err != nil {
		return
	}

	payload = (*[1 << 30]byte)(unsafe.Pointer(out.ptr))[:int(out.size):int(out.size)]

	return
}
