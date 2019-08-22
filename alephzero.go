package alephzero

/*
#cgo pkg-config: alephzero
#include "alephzero_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"unsafe"
)

type AlephZero struct {
	c C.a0_alephzero_t
}

func NewAlephZero() (a0 AlephZero, err error) {
	err = errorFrom(C.a0_alephzero_init(&a0.c))
	return
}

func (a0 *AlephZero) Close() error {
	return errorFrom(C.a0_alephzero_close(&a0.c))
}

func (a0 *AlephZero) NewConfigReaderSync() (ss SubscriberSync, err error) {
	err = errorFrom(C.a0_config_reader_sync_init(&ss.c, a0.c))
	return
}

func (a0 *AlephZero) NewConfigReader(callback func(Packet)) (s Subscriber, err error) {
	s.packetCallbackId = registerPacketCallback(func(cPkt C.a0_packet_t) {
		callback(packetFromC(cPkt))
	})

	err = errorFrom(C.a0go_config_reader_init(&s.c, a0.c, C.uintptr_t(s.packetCallbackId)))
	return
}

func (a0 *AlephZero) NewPublisher(name string) (p Publisher, err error) {
	nameCStr := C.CString(name)
	defer C.free(unsafe.Pointer(nameCStr))
	err = errorFrom(C.a0_publisher_init(&p.c, a0.c, nameCStr))
	return
}

func (a0 *AlephZero) NewSubscriberSync(name string, readStart SubscriberReadStart, readNext SubscriberReadNext) (ss SubscriberSync, err error) {
	nameCStr := C.CString(name)
	defer C.free(unsafe.Pointer(nameCStr))
	err = errorFrom(C.a0_subscriber_sync_init(&ss.c, a0.c, nameCStr, C.a0_subscriber_read_start_t(readStart), C.a0_subscriber_read_next_t(readNext)))
	return
}

func (a0 *AlephZero) NewSubscriber(name string, readStart SubscriberReadStart, readNext SubscriberReadNext, callback func(Packet)) (s Subscriber, err error) {
	nameCStr := C.CString(name)
	defer C.free(unsafe.Pointer(nameCStr))

	s.packetCallbackId = registerPacketCallback(func(cPkt C.a0_packet_t) {
		callback(packetFromC(cPkt))
	})

	err = errorFrom(C.a0go_subscriber_init(&s.c, a0.c, nameCStr, C.a0_subscriber_read_start_t(readStart), C.a0_subscriber_read_next_t(readNext), C.uintptr_t(s.packetCallbackId)))
	return
}

func (a0 *AlephZero) NewRpcServer(name string, onrequest func(Packet), oncancel func(string)) (rs RpcServer, err error) {
	nameCStr := C.CString(name)
	defer C.free(unsafe.Pointer(nameCStr))

	rs.onrequestId = registerPacketCallback(func(cPkt C.a0_packet_t) {
		onrequest(packetFromC(cPkt))
	})

	rs.oncancelId = registerPacketIdCallback(func(cPktId *C.char) {
		oncancel(C.GoStringN(cPktId, C.A0_PACKET_ID_SIZE))
	})

	err = errorFrom(C.a0go_rpc_server_init(&rs.c, a0.c, nameCStr, C.uintptr_t(rs.onrequestId), C.uintptr_t(rs.oncancelId)))
	return
}

func (a0 *AlephZero) NewRpcClient(name string) (rc RpcClient, err error) {
	nameCStr := C.CString(name)
	defer C.free(unsafe.Pointer(nameCStr))

	err = errorFrom(C.a0_rpc_client_init(&rc.c, a0.c, nameCStr))
	return
}
