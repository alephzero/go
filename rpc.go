package alephzero

/*
#cgo pkg-config: alephzero
#include "rpc_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

type RpcServer struct {
	c           C.a0_rpc_server_t
	allocId     uintptr
	activePkt   Packet
	onrequestId uintptr
	oncancelId  uintptr
}

func NewRpcServer(shm ShmObj, onrequest func(Packet), oncancel func(Packet)) (rs RpcServer, err error) {
	rs.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		rs.activePkt.goMem = make([]byte, int(size))
		out.size = size
		out.ptr = (*C.uint8_t)(&rs.activePkt.goMem[0])
	})

	rs.onrequestId = registerPacketCallback(func(cPkt C.a0_packet_t) {
		onrequest(Packet{cPkt, nil})
		rs.activePkt.goMem = nil
	})

	rs.oncancelId = registerPacketCallback(func(cPkt C.a0_packet_t) {
		onrequest(Packet{cPkt, nil})
		rs.activePkt.goMem = nil
	})

	err = errorFrom(C.a0go_rpc_server_init(&rs.c, shm.c, C.uintptr_t(rs.allocId), C.uintptr_t(rs.onrequestId), C.uintptr_t(rs.oncancelId)))
	return
}

func (rs *RpcServer) Close() error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
		unregisterPacketCallback(rs.onrequestId)
		unregisterPacketCallback(rs.oncancelId)
		unregisterAlloc(rs.allocId)
	})
	return errorFrom(C.a0go_rpc_server_close(&rs.c, C.uintptr_t(callbackId)))
}

func (rs *RpcServer) Reply(req Packet, resp Packet) error {
	return errorFrom(C.a0_rpc_reply(&rs.c, req.c, resp.c))
}

type RpcClient struct {
	c         C.a0_rpc_client_t
	allocId   uintptr
	activePkt Packet
}

func NewRpcClient(shm ShmObj) (rc RpcClient, err error) {
	rc.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		rc.activePkt.goMem = make([]byte, int(size))
		out.size = size
		out.ptr = (*C.uint8_t)(&rc.activePkt.goMem[0])
	})

	err = errorFrom(C.a0go_rpc_client_init(&rc.c, shm.c, C.uintptr_t(rc.allocId)))
	return
}

func (rc *RpcClient) Close() error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
		unregisterAlloc(rc.allocId)
	})
	return errorFrom(C.a0go_rpc_client_close(&rc.c, C.uintptr_t(callbackId)))
}

func (rc *RpcClient) Send(pkt Packet, replyCb func(Packet)) error {
	var packetCallbackId uintptr
	packetCallbackId = registerPacketCallback(func(cPkt C.a0_packet_t) {
		replyCb(Packet{cPkt, nil})
		unregisterPacketCallback(packetCallbackId)
	})
	return errorFrom(C.a0go_rpc_send(&rc.c, pkt.c, C.uintptr_t(packetCallbackId)))
}

func (rc *RpcClient) Cancel(pkt Packet) error {
	return errorFrom(C.a0_rpc_cancel(&rc.c, pkt.c))
}
