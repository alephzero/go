package alephzero

/*
#cgo pkg-config: alephzero
#include "pubsub_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"sync"
	"unsafe"
)

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
	READ_READ_NEXT_SEQUENTIAL SubscriberReadNext = C.A0_READ_NEXT_SEQUENTIAL
	READ_READ_NEXT_RECENT                        = C.A0_READ_NEXT_RECENT
)

type SubscriberSync struct {
	c C.a0_subscriber_sync_t
}

func NewSubscriberSync(shm ShmObj, readStart SubscriberReadStart, readNext SubscriberReadNext) (ss SubscriberSync, err error) {
	err = errorFrom(C.a0_subscriber_sync_init(&ss.c, shm.c, C.a0_subscriber_read_start_t(readStart), C.a0_subscriber_read_next_t(readNext)))
	return
}

func (ss *SubscriberSync) Close() error {
	return errorFrom(C.a0_subscriber_sync_close(&ss.c))
}

func (ss *SubscriberSync) HasNext() (hasNext bool, err error) {
	err = errorFrom(C.a0_subscriber_sync_has_next(&ss.c, (*C.bool)(&hasNext)))
	return
}

func (ss *SubscriberSync) Next() (pkt Packet, err error) {
	allocId := registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		pkt.goMem = make([]byte, int(size))
		out.size = size
		out.ptr = (*C.uint8_t)(&pkt.goMem[0])
	})
	defer unregisterAlloc(allocId)

	err = errorFrom(C.a0go_subscriber_sync_next(&ss.c, C.uintptr_t(allocId), &pkt.c))
	return
}

var (
	subscriberCallbackRegistry     = make(map[uintptr]func(C.a0_packet_t))
	subscriberCallbackRegistryLock = sync.Mutex{}
	nextSubscriberCallbackId       uintptr
)

//export a0go_subscriber_callback
func a0go_subscriber_callback(id unsafe.Pointer, c C.a0_packet_t) {
	// TODO: Should this be a reader lock?
	subscriberCallbackRegistryLock.Lock()
	defer subscriberCallbackRegistryLock.Unlock()
	subscriberCallbackRegistry[uintptr(id)](c)
}

func registerSubscriberCallback(fn func(C.a0_packet_t)) (id uintptr) {
	subscriberCallbackRegistryLock.Lock()
	defer subscriberCallbackRegistryLock.Unlock()
	id = nextSubscriberCallbackId
	nextSubscriberCallbackId++
	subscriberCallbackRegistry[id] = fn
	return
}

func unregisterSubscriberCallback(id uintptr) {
	subscriberCallbackRegistryLock.Lock()
	defer subscriberCallbackRegistryLock.Unlock()
	delete(subscriberCallbackRegistry, id)
}

type Subscriber struct {
	c                    C.a0_subscriber_t
	allocId              uintptr
	subscriberCallbackId uintptr
	activePkt            Packet
}

func NewSubscriber(shm ShmObj, readStart SubscriberReadStart, readNext SubscriberReadNext, callback func(Packet)) (s Subscriber, err error) {
	s.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		s.activePkt.goMem = make([]byte, int(size))
		out.size = size
		out.ptr = (*C.uint8_t)(&s.activePkt.goMem[0])
	})

	s.subscriberCallbackId = registerSubscriberCallback(func(_ C.a0_packet_t) {
		callback(s.activePkt)
		s.activePkt.goMem = nil
	})

	err = errorFrom(C.a0go_subscriber_init(&s.c, shm.c, C.a0_subscriber_read_start_t(readStart), C.a0_subscriber_read_next_t(readNext), C.uintptr_t(s.allocId), C.uintptr_t(s.subscriberCallbackId)))
	return
}

func (s *Subscriber) Close() error {
	var callbackId uintptr
	callbackId = registerCallback(func() {
		unregisterCallback(callbackId)
	})
	return errorFrom(C.a0go_subscriber_close(&s.c, C.uintptr_t(callbackId)))
}
