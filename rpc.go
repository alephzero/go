package alephzero

/*
#cgo pkg-config: alephzero
#include "rpc_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"sync"
	"unsafe"
)

type RpcRequest struct {
	c C.a0_rpc_request_t
}

func (req *RpcRequest) Packet() Packet {
	return packetFromC(req.c.pkt)
}

func (req *RpcRequest) Reply(resp Packet) error {
	cPkt := resp.c()
	defer freeCPacket(cPkt)
	return errorFrom(C.a0_rpc_reply(req.c, cPkt))
}

type RpcServer struct {
	c           C.a0_rpc_server_t
	allocId     uintptr
	onrequestId uintptr
	oncancelId  uintptr
}

func NewRpcServer(shm Shm, onrequest func(RpcRequest), oncancel func(string)) (rs *RpcServer, err error) {
	rs = &RpcServer{}

	var activePktSpace []byte
	rs.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		activePktSpace = make([]byte, int(size))
		wrapGoMem(activePktSpace, out)
	})

	rs.onrequestId = registerRpcRequestCallback(func(cReq C.a0_rpc_request_t) {
		onrequest(RpcRequest{cReq})
		_ = activePktSpace // keep alive
	})

	rs.oncancelId = registerPacketIdCallback(func(cReqId *C.char) {
		oncancel(C.GoString(cReqId))
		_ = activePktSpace // keep alive
	})

	err = errorFrom(C.a0go_rpc_server_init(&rs.c, shm.c.buf, C.uintptr_t(rs.allocId), C.uintptr_t(rs.onrequestId), C.uintptr_t(rs.oncancelId)))
	return
}

func (rs *RpcServer) AsyncClose(fn func()) error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
		unregisterRpcRequestCallback(rs.onrequestId)
		unregisterPacketIdCallback(rs.oncancelId)
		if rs.allocId > 0 {
			unregisterAlloc(rs.allocId)
		}
		if fn != nil {
			fn()
		}
	})
	return errorFrom(C.a0go_rpc_server_async_close(&rs.c, C.uintptr_t(callbackId)))
}

func (rs *RpcServer) Close() (err error) {
	err = errorFrom(C.a0_rpc_server_close(&rs.c))
	unregisterRpcRequestCallback(rs.onrequestId)
	unregisterPacketIdCallback(rs.oncancelId)
	if rs.allocId > 0 {
		unregisterAlloc(rs.allocId)
	}
	return
}

type RpcClient struct {
	c       C.a0_rpc_client_t
	allocId uintptr
	// Memory must survive between the alloc and replyCb.
	activePktSpace []byte
}

func NewRpcClient(shm Shm) (rc *RpcClient, err error) {
	rc = &RpcClient{}

	rc.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		rc.activePktSpace = make([]byte, int(size))
		wrapGoMem(rc.activePktSpace, out)
	})

	err = errorFrom(C.a0go_rpc_client_init(&rc.c, shm.c.buf, C.uintptr_t(rc.allocId)))
	return
}

func (rc *RpcClient) AsyncClose(fn func()) error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
		unregisterAlloc(rc.allocId)
		if fn != nil {
			fn()
		}
	})
	return errorFrom(C.a0go_rpc_client_async_close(&rc.c, C.uintptr_t(callbackId)))
}

func (rc *RpcClient) Close() (err error) {
	err = errorFrom(C.a0_rpc_client_close(&rc.c))
	unregisterAlloc(rc.allocId)
	return
}

func (rc *RpcClient) Send(pkt Packet, replyCb func(Packet)) error {
	var packetCallbackId uintptr
	packetCallbackId = registerPacketCallback(func(cPkt C.a0_packet_t) {
		replyCb(packetFromC(cPkt))
		unregisterPacketCallback(packetCallbackId)
	})

	cPkt := pkt.c()
	defer freeCPacket(cPkt)

	return errorFrom(C.a0go_rpc_send(&rc.c, cPkt, C.uintptr_t(packetCallbackId)))
}

func (rc *RpcClient) Cancel(reqId string) error {
	cReqId := C.CString(reqId)
	defer C.free(unsafe.Pointer(cReqId))
	return errorFrom(C.a0_rpc_cancel(&rc.c, cReqId))
}

var (
	rpcRequestCallbackMutex    = sync.Mutex{}
	rpcRequestCallbackRegistry = make(map[uintptr]func(C.a0_rpc_request_t))
	nextRpcRequestCallbackId   uintptr
)

//export a0go_rpc_request_callback
func a0go_rpc_request_callback(id unsafe.Pointer, c C.a0_rpc_request_t) {
	rpcRequestCallbackMutex.Lock()
	fn := rpcRequestCallbackRegistry[uintptr(id)]
	rpcRequestCallbackMutex.Unlock()
	fn(c)
}

func registerRpcRequestCallback(fn func(C.a0_rpc_request_t)) (id uintptr) {
	rpcRequestCallbackMutex.Lock()
	defer rpcRequestCallbackMutex.Unlock()
	id = nextRpcRequestCallbackId
	nextRpcRequestCallbackId++
	rpcRequestCallbackRegistry[id] = fn
	return
}

func unregisterRpcRequestCallback(id uintptr) {
	rpcRequestCallbackMutex.Lock()
	defer rpcRequestCallbackMutex.Unlock()
	delete(rpcRequestCallbackRegistry, id)
}
