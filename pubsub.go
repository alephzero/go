package alephzero

// #cgo pkg-config: alephzero
// #include "pubsub_adapter.h"
import "C"

import (
	"unsafe"
)

type Publisher struct {
	cPub C.a0_publisher_t
}

func NewPublisherUnmapped(container, topic string) (p Publisher, err error) {
	err = errorFrom(C.a0_publisher_init_unmapped(&p.cPub, C.CString(container), C.CString(topic)))
	return
}

func (p *Publisher) Close() error {
	return errorFrom(C.a0_publisher_close(&p.cPub))
}

func (p *Publisher) Pub(pkt Packet) error {
	return errorFrom(C.a0_pub(&p.cPub, pkt.cPkt))
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
	cSubSync C.a0_subscriber_sync_t
}

func NewSubscriberSyncUnmapped(container, topic string, readStart SubscriberReadStart, readNext SubscriberReadNext) (ss SubscriberSync, err error) {
	containerCStr := C.CString(container)
	defer C.free(unsafe.Pointer(containerCStr))

	topicCStr := C.CString(topic)
	defer C.free(unsafe.Pointer(topicCStr))

	err = errorFrom(C.a0_subscriber_sync_init_unmapped(&ss.cSubSync, containerCStr, topicCStr, readStart, readNext))
	return
}

func (ss *SubscriberSync) Close() error {
	return errorFrom(C.a0_subscriber_sync_close(&ss.cSubSync))
}

func (ss *SubscriberSync) HasNext() (hasNext bool, err error) {
	err = errorFrom(C.a0_subscriber_sync_has_next(&ss.cSubSync, &hasNext))
	return
}

func (ss *SubscriberSync) Next() (pkt Packet, err error) {
	allocId := registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		pkt.goMem = make([]byte, int(size))
		out.size = size
		out.ptr = (*C.uint8_t)(&pkt.goMem[0])
	})
	defer unregisterAlloc(allocId)

	err = errorFrom(C.a0go_subscriber_sync_next(&ss.cSubSync, allocId, &pkt.cPkt))
	return
}

var (
	// TODO: make thread safe.
	subscriberCallbackRegistry = make(map[int]func(C.a0_packet_t))
	nextSubscriberCallbackId   int
)

//export a0go_subscriber_callback
func a0go_subscriber_callback(idPtr unsafe.Pointer, cPkt C.a0_packet_t) {
	allocRegistry[int(*(*C.int)(idPtr))](size, cPkt)
}

func registerSubscriberCallback(fn func(C.a0_packet_t)) (id int) {
	id = nextSubscriberCallbackId
	nextSubscriberCallbackId++
	subscriberCallbackRegistry[id] = fn
	return
}

func unregisterSubscriberCallback(id int) {
	delete(subscriberCallbackRegistry, id)
}

type Subscriber struct {
	cSub                 C.a0_subscriber_t
	allocId              int
	subscriberCallbackId int
	activePkt            Packet
}

func NewSubscriberUnmapped(container, topic string, readStart SubscriberReadStart, readNext SubscriberReadNext, callback func(Packet)) (s Subscriber, err error) {
	s.allocId = registerAlloc(func(size C.size_t, out *C.a0_buf_t) {
		s.activePkt.goMem = make([]byte, int(size))
		out.size = size
		out.ptr = (*C.uint8_t)(&s.activePkt.goMem[0])
	})

	s.subscriberCallbackId = registerSubscriberCallback(func(_ C.a0_packet_t) {
		callback(s.activePkt)
		s.activePkt.goMem = nil
	})

	containerCStr := C.CString(container)
	defer C.free(unsafe.Pointer(containerCStr))

	topicCStr := C.CString(topic)
	defer C.free(unsafe.Pointer(topicCStr))

	err = errorFrom(C.a0go_subscriber_init_unmapped(&s.cSub, containerCStr, topicCStr, readStart, readNext, s.allocId, s.subscriberCallbackId))
	return
}

func (ss *Subscriber) Close() error {
	callbackId = registerCallback(func() {})
	return errorFrom(C.a0go_subscriber_close(&s.cSub, callbackId))
}
