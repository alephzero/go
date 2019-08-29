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
	err = errorFrom(C.a0_publisher_init(&p.c, shm.c))
	return
}

func (p *Publisher) Close() error {
	return errorFrom(C.a0_publisher_close(&p.c))
}

func (p *Publisher) Pub(pkt Packet) error {
	return errorFrom(C.a0_pub(&p.c, pkt.C()))
}

type SubscriberInit int

const (
	INIT_OLDEST      SubscriberInit = C.A0_INIT_OLDEST
	INIT_MOST_RECENT                = C.A0_INIT_MOST_RECENT
	INIT_AWAIT_NEW                  = C.A0_INIT_AWAIT_NEW
)

type SubscriberIter int

const (
	ITER_NEXT   SubscriberIter = C.A0_ITER_NEXT
	ITER_NEWEST                = C.A0_ITER_NEWEST
)

type SubscriberSync struct {
	c       C.a0_subscriber_sync_t
	allocId uintptr
	// Memory must survive between the alloc and Next.
	activePkt Packet
}

func NewSubscriberSync(shm ShmObj, subInit SubscriberInit, subIter SubscriberIter) (ss SubscriberSync, err error) {
	ss.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		ss.activePkt = make([]byte, int(size))
		*out = ss.activePkt.C()
	})

	err = errorFrom(C.a0go_subscriber_sync_init(&ss.c, shm.c, C.uintptr_t(ss.allocId), C.a0_subscriber_init_t(subInit), C.a0_subscriber_iter_t(subIter)))
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
	var cPkt C.a0_packet_t
	err = errorFrom(C.a0_subscriber_sync_next(&ss.c, &cPkt))
	if err == nil {
		pkt = packetFromC(cPkt)
	}
	return
}

type Subscriber struct {
	c                C.a0_subscriber_t
	allocId          uintptr
	packetCallbackId uintptr
}

func NewSubscriber(shm ShmObj, subInit SubscriberInit, subIter SubscriberIter, callback func(Packet)) (s Subscriber, err error) {
	var activePkt Packet

	s.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		activePkt = make([]byte, int(size))
		*out = activePkt.C()
	})

	s.packetCallbackId = registerPacketCallback(func(_ C.a0_packet_t) {
		callback(activePkt)
	})

	err = errorFrom(C.a0go_subscriber_init(&s.c, shm.c, C.uintptr_t(s.allocId), C.a0_subscriber_init_t(subInit), C.a0_subscriber_iter_t(subIter), C.uintptr_t(s.packetCallbackId)))
	return
}

func (s *Subscriber) AsyncClose(fn func()) error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
		unregisterPacketCallback(s.packetCallbackId)
		if s.allocId > 0 {
			unregisterAlloc(s.allocId)
		}
		if fn != nil {
			fn()
		}
	})
	return errorFrom(C.a0go_subscriber_async_close(&s.c, C.uintptr_t(callbackId)))
}

func (s *Subscriber) Close() (err error) {
	err = errorFrom(C.a0_subscriber_close(&s.c))
	unregisterPacketCallback(s.packetCallbackId)
	if s.allocId > 0 {
		unregisterAlloc(s.allocId)
	}
	return
}
