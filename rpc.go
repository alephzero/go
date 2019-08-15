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

type RpcRequest struct {
	c C.a0_rpc_request_t
}

func (req *RpcRequest) Packet() (p Packet) {
	p.c = req.c.pkt
	return
}

var (
	// TODO: Thread safety.
	rpcServerOnRequestRegistry     = make(map[uintptr]func(C.a0_rpc_request_t))
	nextRpcServerOnRequestId       uintptr
)

//export a0go_rpc_server_onrequest
func a0go_rpc_server_onrequest(id unsafe.Pointer, req C.a0_rpc_request_t) {
	rpcServerOnRequestRegistry[uintptr(id)](req)
}

func registerRpcServerOnRequest(fn func(C.a0_rpc_request_t)) (id uintptr) {
	id = nextRpcServerOnRequestId
	nextRpcServerOnRequestId++
	rpcServerOnRequestRegistry[id] = fn
	return
}

func unregisterRpcServerOnRequest(id uintptr) {
	delete(rpcServerOnRequestRegistry, id)
}

type RpcServer struct {
	c                    C.a0_rpc_server_t
	allocId              uintptr
	rpcServerOnRequestId uintptr
	activePkt            Packet
}

func NewRpcServer(requestShm, responseShm ShmObj, onrequest func(RpcRequest)) (rs RpcServer, err error) {
	rs.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		rs.activePkt.goMem = make([]byte, int(size))
		out.size = size
		out.ptr = (*C.uint8_t)(&rs.activePkt.goMem[0])
	})

	rs.rpcServerOnRequestId = registerRpcServerOnRequest(func(req C.a0_rpc_request_t) {
		onrequest(RpcRequest{req})
		rs.activePkt.goMem = nil
	})

	err = errorFrom(C.a0go_rpc_server_init(&rs.c, requestShm.c, responseShm.c, C.uintptr_t(rs.allocId), C.uintptr_t(rs.rpcServerOnRequestId)))
	return
}

func (rs *RpcServer) Close() error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
		unregisterRpcServerOnRequest(rs.rpcServerOnRequestId)
		unregisterAlloc(rs.allocId)
	})
	return errorFrom(C.a0go_rpc_server_close(&rs.c, C.uintptr_t(callbackId)))
}

func (rs *RpcServer) Reply(req RpcRequest, pkt Packet) error {
	return errorFrom(C.a0_rpc_reply(&rs.c, req.c, pkt.c))
}

type RpcClient struct {
	c         C.a0_rpc_client_t
	allocId   uintptr
	activePkt Packet
}

func NewRpcClient(requestShm, responseShm ShmObj) (rc RpcClient, err error) {
	rc.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		rc.activePkt.goMem = make([]byte, int(size))
		out.size = size
		out.ptr = (*C.uint8_t)(&rc.activePkt.goMem[0])
	})

	err = errorFrom(C.a0go_rpc_client_init(&rc.c, requestShm.c, responseShm.c, C.uintptr_t(rc.allocId)))
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
