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

type RpcTopic struct {
	Name        string
	FileOptions *FileOptions
}

func (t *RpcTopic) c() (cTopic C.a0_rpc_topic_t) {
	cTopic.name = C.CString(t.Name)
	if t.FileOptions != nil {
		localOpts := t.FileOptions.toC()
		cTopic.file_opts = &localOpts
	}
	return
}

func freeCRpcTopic(cTopic C.a0_rpc_topic_t) {
	C.free(unsafe.Pointer(cTopic.name))
}

type RpcRequest struct {
	c C.a0_rpc_request_t
}

func (req *RpcRequest) Packet() Packet {
	return packetFromC(req.c.pkt)
}

func (req *RpcRequest) Reply(resp Packet) error {
	cPkt := resp.c()
	defer freeCPacket(cPkt)
	return errorFrom(C.a0_rpc_server_reply(req.c, cPkt))
}

type RpcServer struct {
	c           C.a0_rpc_server_t
	allocId     uintptr
	onrequestId uintptr
	oncancelId  uintptr
}

func NewRpcServer(topic RpcTopic, onrequest func(RpcRequest), oncancel func(string)) (rs *RpcServer, err error) {
	rs = &RpcServer{}

	cTopic := topic.c()
	defer freeCRpcTopic(cTopic)

	var activePktSpace []byte
	rs.allocId = registry.Register(func(size C.size_t, out *C.a0_buf_t) C.a0_err_t {
		activePktSpace = make([]byte, int(size))
		out.size = size
		if size > 0 {
			out.data = (*C.uint8_t)(&activePktSpace[0])
		}
		return A0_OK
	})

	rs.onrequestId = registry.Register(func(cReq C.a0_rpc_request_t) {
		onrequest(RpcRequest{cReq})
		_ = activePktSpace // keep alive
	})

	rs.oncancelId = registry.Register(func(cReqId *C.char) {
		oncancel(C.GoString(cReqId))
		_ = activePktSpace // keep alive
	})

	err = errorFrom(C.a0go_rpc_server_init(&rs.c, cTopic, C.uintptr_t(rs.allocId), C.uintptr_t(rs.onrequestId), C.uintptr_t(rs.oncancelId)))
	return
}

func (rs *RpcServer) Close() (err error) {
	err = errorFrom(C.a0_rpc_server_close(&rs.c))
	registry.Unregister(rs.onrequestId)
	registry.Unregister(rs.oncancelId)
	if rs.allocId > 0 {
		registry.Unregister(rs.allocId)
	}
	return
}

type RpcClient struct {
	c       C.a0_rpc_client_t
	allocId uintptr
	// Memory must survive between the alloc and replyCb.
	activePktSpace []byte
}

func NewRpcClient(topic RpcTopic) (rc *RpcClient, err error) {
	rc = &RpcClient{}

	cTopic := topic.c()
	defer freeCRpcTopic(cTopic)

	rc.allocId = registry.Register(func(size C.size_t, out *C.a0_buf_t) C.a0_err_t {
		rc.activePktSpace = make([]byte, int(size))
		out.size = size
		if size > 0 {
			out.data = (*C.uint8_t)(&rc.activePktSpace[0])
		}
		return A0_OK
	})

	err = errorFrom(C.a0go_rpc_client_init(&rc.c, cTopic, C.uintptr_t(rc.allocId)))
	return
}

func (rc *RpcClient) Close() (err error) {
	err = errorFrom(C.a0_rpc_client_close(&rc.c))
	registry.Unregister(rc.allocId)
	return
}

func (rc *RpcClient) Send(pkt Packet, replyCb func(Packet)) error {
	var packetCallbackId uintptr
	packetCallbackId = registry.Register(func(cPkt C.a0_packet_t) {
		replyCb(packetFromC(cPkt))
		registry.Unregister(packetCallbackId)
	})

	cPkt := pkt.c()
	defer freeCPacket(cPkt)

	return errorFrom(C.a0go_rpc_send(&rc.c, cPkt, C.uintptr_t(packetCallbackId)))
}

func (rc *RpcClient) Cancel(reqId string) error {
	cReqId := C.CString(reqId)
	defer C.free(unsafe.Pointer(cReqId))
	return errorFrom(C.a0_rpc_client_cancel(&rc.c, cReqId))
}

//export a0go_rpc_request_callback
func a0go_rpc_request_callback(id unsafe.Pointer, c C.a0_rpc_request_t) {
	registry.Get(uintptr(id)).(func(C.a0_rpc_request_t))(c)
}
