package alephzero

/*
#cgo pkg-config: alephzero
#include "rpc_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"unsafe"
)

type RpcServer struct {
	c           C.a0_rpc_server_t
	allocId     uintptr
	onrequestId uintptr
	oncancelId  uintptr
}

func NewRpcServer(shm ShmObj, onrequest func(Packet), oncancel func(string)) (rs RpcServer, err error) {
	var activePkt Packet

	rs.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		activePkt = Packet{make([]byte, int(size))}
		*out = activePkt.C()
	})

	rs.onrequestId = registerPacketCallback(func(_ C.a0_packet_t) {
		onrequest(activePkt)
	})

	rs.oncancelId = registerPacketIdCallback(func(cReqId *C.char) {
		oncancel(C.GoString(cReqId))
	})

	err = errorFrom(C.a0go_rpc_server_init_unmanaged(&rs.c, shm.c, C.uintptr_t(rs.allocId), C.uintptr_t(rs.onrequestId), C.uintptr_t(rs.oncancelId)))
	return
}

func (rs *RpcServer) Close(fn func()) error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
		unregisterPacketCallback(rs.onrequestId)
		unregisterPacketCallback(rs.oncancelId)
		if rs.allocId > 0 {
			unregisterAlloc(rs.allocId)
		}
		if fn != nil {
			fn()
		}
	})
	return errorFrom(C.a0go_rpc_server_close(&rs.c, C.uintptr_t(callbackId)))
}

func (rs *RpcServer) Reply(reqId string, resp Packet) error {
	cReqId := C.CString(reqId)
	defer C.free(unsafe.Pointer(cReqId))
	return errorFrom(C.a0_rpc_reply(&rs.c, cReqId, resp.C()))
}

type RpcClient struct {
	c       C.a0_rpc_client_t
	allocId uintptr
	// Memory must survive between the alloc and replyCb.
	activePkt Packet
}

func NewRpcClient(shm ShmObj) (rc RpcClient, err error) {
	rc.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		rc.activePkt = Packet{make([]byte, int(size))}
		*out = rc.activePkt.C()
	})

	err = errorFrom(C.a0go_rpc_client_init_unmanaged(&rc.c, shm.c, C.uintptr_t(rc.allocId)))
	return
}

func (rc *RpcClient) Close(fn func()) error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
		unregisterAlloc(rc.allocId)
		if fn != nil {
			fn()
		}
	})
	return errorFrom(C.a0go_rpc_client_close(&rc.c, C.uintptr_t(callbackId)))
}

func (rc *RpcClient) Send(pkt Packet, replyCb func(Packet)) error {
	var packetCallbackId uintptr
	packetCallbackId = registerPacketCallback(func(cPkt C.a0_packet_t) {
		// TODO: Maybe use activePkt, if using unmanaged api.
		replyCb(packetFromC(cPkt))
		unregisterPacketCallback(packetCallbackId)
	})
	return errorFrom(C.a0go_rpc_send(&rc.c, pkt.C(), C.uintptr_t(packetCallbackId)))
}

func (rc *RpcClient) Cancel(reqId string) error {
	cReqId := C.CString(reqId)
	defer C.free(unsafe.Pointer(cReqId))
	return errorFrom(C.a0_rpc_cancel(&rc.c, cReqId))
}
