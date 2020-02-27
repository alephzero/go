package alephzero

/*
#cgo pkg-config: alephzero
#include "prpc_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"sync"
	"unsafe"
)

type PrpcConnection struct {
	c C.a0_prpc_connection_t
}

func (conn *PrpcConnection) Packet() Packet {
	return packetFromC(conn.c.pkt)
}

func (conn *PrpcConnection) Send(prog Packet, done bool) error {
	cPkt := prog.c()
	defer freeCPacket(cPkt)
	return errorFrom(C.a0_prpc_send(conn.c, cPkt, C.bool(done)))
}

type PrpcServer struct {
	c           C.a0_prpc_server_t
	allocId     uintptr
	onconnectId uintptr
	oncancelId  uintptr
}

func NewPrpcServer(shm Shm, onconnect func(PrpcConnection), oncancel func(string)) (rs *PrpcServer, err error) {
	rs = &PrpcServer{}

	var activePktSpace []byte
	rs.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		activePktSpace = make([]byte, int(size))
		wrapGoMem(activePktSpace, out)
	})

	rs.onconnectId = registerPrpcConnectionCallback(func(cConn C.a0_prpc_connection_t) {
		onconnect(PrpcConnection{cConn})
		_ = activePktSpace  // keep alive
	})

	rs.oncancelId = registerPacketIdCallback(func(cConnId *C.char) {
		oncancel(C.GoString(cConnId))
		_ = activePktSpace  // keep alive
	})

	err = errorFrom(C.a0go_prpc_server_init(&rs.c, shm.c.buf, C.uintptr_t(rs.allocId), C.uintptr_t(rs.onconnectId), C.uintptr_t(rs.oncancelId)))
	return
}

func (rs *PrpcServer) AsyncClose(fn func()) error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
		unregisterPrpcConnectionCallback(rs.onconnectId)
		unregisterPacketIdCallback(rs.oncancelId)
		if rs.allocId > 0 {
			unregisterAlloc(rs.allocId)
		}
		if fn != nil {
			fn()
		}
	})
	return errorFrom(C.a0go_prpc_server_async_close(&rs.c, C.uintptr_t(callbackId)))
}

func (rs *PrpcServer) Close() (err error) {
	err = errorFrom(C.a0_prpc_server_close(&rs.c))
	unregisterPrpcConnectionCallback(rs.onconnectId)
	unregisterPacketIdCallback(rs.oncancelId)
	if rs.allocId > 0 {
		unregisterAlloc(rs.allocId)
	}
	return
}

type PrpcClient struct {
	c       C.a0_prpc_client_t
	allocId uintptr
	// Memory must survive between the alloc and replyCb.
	activePktSpace []byte
}

func NewPrpcClient(shm Shm) (rc *PrpcClient, err error) {
	rc = &PrpcClient{}

	rc.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		rc.activePktSpace = make([]byte, int(size))
		wrapGoMem(rc.activePktSpace, out)
	})

	err = errorFrom(C.a0go_prpc_client_init(&rc.c, shm.c.buf, C.uintptr_t(rc.allocId)))
	return
}

func (rc *PrpcClient) AsyncClose(fn func()) error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
		unregisterAlloc(rc.allocId)
		if fn != nil {
			fn()
		}
	})
	return errorFrom(C.a0go_prpc_client_async_close(&rc.c, C.uintptr_t(callbackId)))
}

func (rc *PrpcClient) Close() (err error) {
	err = errorFrom(C.a0_prpc_client_close(&rc.c))
	unregisterAlloc(rc.allocId)
	return
}

func (rc *PrpcClient) Connect(pkt Packet, progCb func(Packet, bool)) error {
	var prpcCallbackId uintptr
	prpcCallbackId = registerPrpcCallback(func(cPkt C.a0_packet_t, done C.bool) {
		progCb(packetFromC(cPkt), bool(done))
		if done {
			unregisterPrpcCallback(prpcCallbackId)
		}
	})

	cPkt := pkt.c()
	defer freeCPacket(cPkt)

	return errorFrom(C.a0go_prpc_connect(&rc.c, cPkt, C.uintptr_t(prpcCallbackId)))
}

func (rc *PrpcClient) Cancel(reqId string) error {
	cReqId := C.CString(reqId)
	defer C.free(unsafe.Pointer(cReqId))
	return errorFrom(C.a0_prpc_cancel(&rc.c, cReqId))
}

var (
	prpcConnectionCallbackMutex    = sync.Mutex{}
	prpcConnectionCallbackRegistry = make(map[uintptr]func(C.a0_prpc_connection_t))
	nextPrpcConnectionCallbackId   uintptr
)

//export a0go_prpc_connection_callback
func a0go_prpc_connection_callback(id unsafe.Pointer, c C.a0_prpc_connection_t) {
	prpcConnectionCallbackMutex.Lock()
	fn := prpcConnectionCallbackRegistry[uintptr(id)]
	prpcConnectionCallbackMutex.Unlock()
	fn(c)
}

func registerPrpcConnectionCallback(fn func(C.a0_prpc_connection_t)) (id uintptr) {
	prpcConnectionCallbackMutex.Lock()
	defer prpcConnectionCallbackMutex.Unlock()
	id = nextPrpcConnectionCallbackId
	nextPrpcConnectionCallbackId++
	prpcConnectionCallbackRegistry[id] = fn
	return
}

func unregisterPrpcConnectionCallback(id uintptr) {
	prpcConnectionCallbackMutex.Lock()
	defer prpcConnectionCallbackMutex.Unlock()
	delete(prpcConnectionCallbackRegistry, id)
}

var (
	prpcCallbackMutex    = sync.Mutex{}
	prpcCallbackRegistry = make(map[uintptr]func(C.a0_packet_t, C.bool))
	nextPrpcCallbackId   uintptr
)

//export a0go_prpc_callback
func a0go_prpc_callback(id unsafe.Pointer, cPkt C.a0_packet_t, done C.bool) {
	prpcCallbackMutex.Lock()
	fn := prpcCallbackRegistry[uintptr(id)]
	prpcCallbackMutex.Unlock()
	fn(cPkt, C.bool(done))
}

func registerPrpcCallback(fn func(C.a0_packet_t, C.bool)) (id uintptr) {
	prpcCallbackMutex.Lock()
	defer prpcCallbackMutex.Unlock()
	id = nextPrpcCallbackId
	nextPrpcCallbackId++
	prpcCallbackRegistry[id] = fn
	return
}

func unregisterPrpcCallback(id uintptr) {
	prpcCallbackMutex.Lock()
	defer prpcCallbackMutex.Unlock()
	delete(prpcCallbackRegistry, id)
}
