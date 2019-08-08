package alephzero

// #cgo pkg-config: alephzero
// #include "adapter.h"
import "C"

import (
	"syscall"
	"unsafe"
)

var (
	// TODO: make thread safe.
	allocRegistry = make(map[int]func(C.size_t, *C.a0_buf_t))
	nextAllocId   int
)

func errorFrom(err C.errno_t) error {
	if err == 0 {
		return nil
	}
	return syscall.Errno(err)
}

//export a0go_alloc
func a0go_alloc(idPtr unsafe.Pointer, size C.size_t, out *C.a0_buf_t) {
	allocRegistry[int(*(*C.int)(idPtr))](size, out)
}

func registerAlloc(fn func(C.size_t, *C.a0_buf_t)) (id int) {
	id = nextAllocId
	nextAllocId++
	allocRegistry[id] = fn
	return
}

func unregisterAlloc(id int) {
	delete(allocRegistry, id)
}

type Header struct {
	Key, Val []byte
}

type Packet struct {
	cPkt  C.a0_packet_t
	goMem []byte
}

func NewPacket(hdrs []Header, payload []byte) (pkt Packet, err error) {
	// TODO: What if payload is empty?
	var cPayload C.a0_buf_t
	cPayload.size = C.size_t(len(payload))
	cPayload.ptr = (*C.uint8_t)(&payload[0])

	// TODO: What if headers are empty?
	cHdrs := C.malloc(C.size_t(len(hdrs) * int(unsafe.Sizeof(C.a0_packet_header_t{}))))
	defer C.free(cHdrs)

	for i, hdr := range hdrs {
		cHdr := (*[1<<30]C.a0_packet_header_t)(cHdrs)[i]
		cHdr.key.size = C.size_t(len(hdr.Key))
		cHdr.key.ptr = (*C.uint8_t)(&hdr.Key[0])
		cHdr.val.size = C.size_t(len(hdr.Val))
		cHdr.val.ptr = (*C.uint8_t)(&hdr.Val[0])
	}

	allocId := registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		pkt.goMem = make([]byte, int(size))
		out.size = size
		out.ptr = (*C.uint8_t)(&pkt.goMem[0])
	})
	defer unregisterAlloc(allocId)

	err = errorFrom(C.a0go_packet_build(
		C.size_t(len(hdrs)),
		(*C.a0_packet_header_t)(cHdrs),
		cPayload,
		C.int(allocId),
		&pkt.cPkt))

	return
}

func (p *Packet) Bytes() ([]byte, error) {
	return p.goMem, nil
}

func (p *Packet) NumHeaders() (cnt int, err error) {
	var ucnt C.size_t
	err = errorFrom(C.a0_packet_num_headers(p.cPkt, &ucnt))
	if err == nil {
		return
	}
	cnt = int(ucnt)
	return
}

func (p *Packet) Header(idx int) (hdr Header, err error) {
	var cHdr C.a0_packet_header_t

	if err = errorFrom(C.a0_packet_header(p.cPkt, C.size_t(idx), &cHdr)); err != nil {
		return
	}

	hdr.Key = (*[1<<30]byte)(unsafe.Pointer(cHdr.key.ptr))[:int(cHdr.key.size):int(cHdr.key.size)]
	hdr.Val = (*[1<<30]byte)(unsafe.Pointer(cHdr.val.ptr))[:int(cHdr.val.size):int(cHdr.val.size)]

	return
}

func (p *Packet) Payload() (payload []byte, err error) {
	var out C.a0_buf_t

	if err = errorFrom(C.a0_packet_payload(p.cPkt, &out)); err != nil {
		return
	}

	payload = (*[1<<30]byte)(unsafe.Pointer(out.ptr))[:int(out.size):int(out.size)]

	return
}
