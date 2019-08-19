package alephzero

/*
#cgo pkg-config: alephzero
#include "pubsub_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

type Publisher struct {
	c C.a0_publisher_t
}

func NewPublisher(shm ShmObj) (p Publisher, err error) {
	err = errorFrom(C.a0_publisher_init_unmanaged(&p.c, shm.c))
	return
}

func (p *Publisher) Close() error {
	return errorFrom(C.a0_publisher_close(&p.c))
}

func (p *Publisher) Pub(pkt Packet) error {
	return errorFrom(C.a0_pub(&p.c, pkt.c))
}

type SubscriberReadStart int

const (
	READ_START_EARLIEST SubscriberReadStart = C.A0_READ_START_EARLIEST
	READ_START_LATEST                       = C.A0_READ_START_LATEST
	READ_START_NEW                          = C.A0_READ_START_NEW
)

type SubscriberReadNext int

const (
	READ_NEXT_SEQUENTIAL SubscriberReadNext = C.A0_READ_NEXT_SEQUENTIAL
	READ_NEXT_RECENT                        = C.A0_READ_NEXT_RECENT
)

type SubscriberSync struct {
	c         C.a0_subscriber_sync_t
	allocId   uintptr
	activePkt Packet
}

func NewSubscriberSync(shm ShmObj, readStart SubscriberReadStart, readNext SubscriberReadNext) (ss SubscriberSync, err error) {
	ss.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		ss.activePkt.goMem = make([]byte, int(size))
		out.size = size
		out.ptr = (*C.uint8_t)(&ss.activePkt.goMem[0])
		ss.activePkt.c = *out
	})

	err = errorFrom(C.a0go_subscriber_sync_init_unmanaged(&ss.c, shm.c, C.uintptr_t(ss.allocId), C.a0_subscriber_read_start_t(readStart), C.a0_subscriber_read_next_t(readNext)))
	return
}

func (ss *SubscriberSync) Close() (err error) {
	err = errorFrom(C.a0_subscriber_sync_close(&ss.c))
	if ss.allocId > 0 {
		unregisterAlloc(ss.allocId)
	}
	return
}

func (ss *SubscriberSync) HasNext() (hasNext bool, err error) {
	err = errorFrom(C.a0_subscriber_sync_has_next(&ss.c, (*C.bool)(&hasNext)))
	return
}

func (ss *SubscriberSync) Next() (pkt Packet, err error) {
	err = errorFrom(C.a0_subscriber_sync_next(&ss.c, &pkt.c))
	return
}

type Subscriber struct {
	c                C.a0_subscriber_t
	allocId          uintptr
	activePkt        Packet
	packetCallbackId uintptr
}

func NewSubscriber(shm ShmObj, readStart SubscriberReadStart, readNext SubscriberReadNext, callback func(Packet)) (s Subscriber, err error) {
	s.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		s.activePkt.goMem = make([]byte, int(size))
		out.size = size
		out.ptr = (*C.uint8_t)(&s.activePkt.goMem[0])
		s.activePkt.c = *out
	})

	s.packetCallbackId = registerPacketCallback(func(_ C.a0_packet_t) {
		callback(s.activePkt)
		s.activePkt.goMem = nil
	})

	err = errorFrom(C.a0go_subscriber_init_unmanaged(&s.c, shm.c, C.uintptr_t(s.allocId), C.a0_subscriber_read_start_t(readStart), C.a0_subscriber_read_next_t(readNext), C.uintptr_t(s.packetCallbackId)))
	return
}

func (s *Subscriber) Close() error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
		unregisterPacketCallback(s.packetCallbackId)
		if s.allocId > 0 {
			unregisterAlloc(s.allocId)
		}
	})
	return errorFrom(C.a0go_subscriber_close(&s.c, C.uintptr_t(callbackId)))
}
