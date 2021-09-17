package alephzero

/*
#cgo pkg-config: alephzero
#include "prpc_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"unsafe"
)

type PrpcTopic struct {
	Name        string
	FileOptions *FileOptions
}

func (t *PrpcTopic) c() (cTopic C.a0_prpc_topic_t) {
	cTopic.name = C.CString(t.Name)
	if t.FileOptions != nil {
		localOpts := t.FileOptions.toC()
		cTopic.file_opts = &localOpts
	}
	return
}

func freeCPrpcTopic(cTopic C.a0_prpc_topic_t) {
	C.free(unsafe.Pointer(cTopic.name))
}

type PrpcConnection struct {
	c C.a0_prpc_connection_t
}

func (conn *PrpcConnection) Packet() Packet {
	return packetFromC(conn.c.pkt)
}

func (conn *PrpcConnection) Send(resp Packet, done bool) error {
	cPkt := resp.c()
	defer freeCPacket(cPkt)
	return errorFrom(C.a0_prpc_server_send(conn.c, cPkt, C.bool(done)))
}

type PrpcServer struct {
	c           C.a0_prpc_server_t
	allocId     uintptr
	onconnectId uintptr
	oncancelId  uintptr
}

func NewPrpcServer(topic PrpcTopic, onconnect func(PrpcConnection), oncancel func(string)) (ps *PrpcServer, err error) {
	ps = &PrpcServer{}

	cTopic := topic.c()
	defer freeCPrpcTopic(cTopic)

	var activePktSpace []byte
	ps.allocId = registry.Register(func(size C.size_t, out *C.a0_buf_t) C.a0_err_t {
		activePktSpace = make([]byte, int(size))
		out.size = size
		if size > 0 {
			out.data = (*C.uint8_t)(&activePktSpace[0])
		}
		return A0_OK
	})

	ps.onconnectId = registry.Register(func(cReq C.a0_prpc_connection_t) {
		onconnect(PrpcConnection{cReq})
		_ = activePktSpace // keep alive
	})

	ps.oncancelId = registry.Register(func(cReqId *C.char) {
		oncancel(C.GoString(cReqId))
		_ = activePktSpace // keep alive
	})

	err = errorFrom(C.a0go_prpc_server_init(&ps.c, cTopic, C.uintptr_t(ps.allocId), C.uintptr_t(ps.onconnectId), C.uintptr_t(ps.oncancelId)))
	return
}

func (ps *PrpcServer) Close() (err error) {
	err = errorFrom(C.a0_prpc_server_close(&ps.c))
	registry.Unregister(ps.onconnectId)
	registry.Unregister(ps.oncancelId)
	if ps.allocId > 0 {
		registry.Unregister(ps.allocId)
	}
	return
}

type PrpcClient struct {
	c       C.a0_prpc_client_t
	allocId uintptr
	// Memory must survive between the alloc and progressCb.
	activePktSpace []byte
}

func NewPrpcClient(topic PrpcTopic) (rc *PrpcClient, err error) {
	rc = &PrpcClient{}

	cTopic := topic.c()
	defer freeCPrpcTopic(cTopic)

	rc.allocId = registry.Register(func(size C.size_t, out *C.a0_buf_t) C.a0_err_t {
		rc.activePktSpace = make([]byte, int(size))
		out.size = size
		if size > 0 {
			out.data = (*C.uint8_t)(&rc.activePktSpace[0])
		}
		return A0_OK
	})

	err = errorFrom(C.a0go_prpc_client_init(&rc.c, cTopic, C.uintptr_t(rc.allocId)))
	return
}

func (rc *PrpcClient) Close() (err error) {
	err = errorFrom(C.a0_prpc_client_close(&rc.c))
	registry.Unregister(rc.allocId)
	return
}

func (rc *PrpcClient) Connect(pkt Packet, progressCb func(Packet, bool)) error {
	var packetCallbackId uintptr
	packetCallbackId = registry.Register(func(cPkt C.a0_packet_t, done C.bool) {
		progressCb(packetFromC(cPkt), bool(done))
		if done {
			registry.Unregister(packetCallbackId)
		}
	})

	cPkt := pkt.c()
	defer freeCPacket(cPkt)

	return errorFrom(C.a0go_prpc_client_connect(&rc.c, cPkt, C.uintptr_t(packetCallbackId)))
}

func (rc *PrpcClient) Cancel(reqId string) error {
	cReqId := C.CString(reqId)
	defer C.free(unsafe.Pointer(cReqId))
	return errorFrom(C.a0_prpc_client_cancel(&rc.c, cReqId))
}

//export a0go_prpc_connection_callback
func a0go_prpc_connection_callback(id unsafe.Pointer, c C.a0_prpc_connection_t) {
	registry.Get(uintptr(id)).(func(C.a0_prpc_connection_t))(c)
}

//export a0go_prpc_progress_callback
func a0go_prpc_progress_callback(id unsafe.Pointer, pkt C.a0_packet_t, done C.bool) {
	registry.Get(uintptr(id)).(func(C.a0_packet_t, C.bool))(pkt, done)
}
