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

type Packet struct {
	id      string
	Headers map[string][]string
	Payload []byte
}

func NewPacket(headers map[string][]string, payload []byte) (pkt Packet) {
	cPkt := C.a0_packet_t{}
	C.a0_packet_init(&cPkt)

	idCStr := ([37]C.char)(cPkt.id)
	pkt.id = C.GoStringN(&idCStr[0], 36)
	pkt.Headers = headers
	pkt.Payload = payload
	return
}

func (p *Packet) ID() string {
	return p.id
}

func PacketDepKey() string {
	return C.GoString(C.a0_packet_dep_key)
}

func packetFromC(cPkt C.a0_packet_t) (pkt Packet) {
	pkt.id = string(C.GoBytes(unsafe.Pointer(&cPkt.id), 36))
	pkt.Headers = make(map[string][]string)
	cHdrs := (*[1 << 30]C.a0_packet_header_t)(unsafe.Pointer(cPkt.headers_block.headers))[:int(cPkt.headers_block.size):int(cPkt.headers_block.size)]
	for i := C.uint64_t(0); i < cPkt.headers_block.size; i++ {
		hdr := &cHdrs[i]
		hdrKey := C.GoString(hdr.key)
		hdrVal := C.GoString(hdr.val)
		pkt.Headers[hdrKey] = append(pkt.Headers[hdrKey], hdrVal)
	}
	pkt.Payload = (*[1 << 30]byte)(unsafe.Pointer(cPkt.payload.ptr))[:int(cPkt.payload.size):int(cPkt.payload.size)]
	return
}

func (p *Packet) c() (cPkt C.a0_packet_t) {
	for i := 0; i < 36; i++ {
		cPkt.id[i] = (C.char)(p.id[i])
	}
	wrapGoMem(p.Payload, &cPkt.payload)

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
	cHdrs := (*[1 << 30]C.a0_packet_header_t)(unsafe.Pointer(cPkt.headers_block.headers))[:int(cPkt.headers_block.size):int(cPkt.headers_block.size)]
	for i := C.uint64_t(0); i < cPkt.headers_block.size; i++ {
		C.free(unsafe.Pointer(cHdrs[i].key))
		C.free(unsafe.Pointer(cHdrs[i].val))
	}
	C.free(unsafe.Pointer(cPkt.headers_block.headers))
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
