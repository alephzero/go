package alephzero

/*
#cgo pkg-config: alephzero
#include "pubsub_adapter.h"
#include <stdlib.h>  // free
*/
import "C"

import (
	"unsafe"
)

type PubSubTopic struct {
	Name        string
	FileOptions *FileOptions
}

func (t *PubSubTopic) c() (cTopic C.a0_pubsub_topic_t) {
	cTopic.name = C.CString(t.Name)
	if t.FileOptions != nil {
		localOpts := t.FileOptions.toC()
		cTopic.file_opts = &localOpts
	}
	return
}

func freeCPubSubTopic(cTopic C.a0_pubsub_topic_t) {
	C.free(unsafe.Pointer(cTopic.name))
}

type Publisher struct {
	c C.a0_publisher_t
}

func NewPublisher(topic PubSubTopic) (p *Publisher, err error) {
	p = &Publisher{}
	cTopic := topic.c()
	defer freeCPubSubTopic(cTopic)
	err = errorFrom(C.a0_publisher_init(&p.c, cTopic))
	return
}

func (p *Publisher) Close() error {
	return errorFrom(C.a0_publisher_close(&p.c))
}

func (p *Publisher) Pub(pkt Packet) error {
	cPkt := pkt.c()
	defer freeCPacket(cPkt)
	return errorFrom(C.a0_publisher_pub(&p.c, cPkt))
}

type SubscriberSync struct {
	c       C.a0_subscriber_sync_t
	allocId uintptr
	// Memory must survive between the alloc and Next.
	activePktSpace []byte
}

func NewSubscriberSync(topic PubSubTopic, init ReaderInit, iter ReaderIter) (ss *SubscriberSync, err error) {
	ss = &SubscriberSync{}

	cTopic := topic.c()
	defer freeCPubSubTopic(cTopic)

	ss.allocId = registry.Register(func(size C.size_t, out *C.a0_buf_t) C.a0_err_t {
		ss.activePktSpace = make([]byte, int(size))
		out.size = size
		if size > 0 {
			out.data = (*C.uint8_t)(&ss.activePktSpace[0])
		}
		return A0_OK
	})

	err = errorFrom(C.a0go_subscriber_sync_init(&ss.c, cTopic, C.uintptr_t(ss.allocId), C.a0_reader_init_t(init), C.a0_reader_iter_t(iter)))
	return
}

func (ss *SubscriberSync) Close() (err error) {
	err = errorFrom(C.a0_subscriber_sync_close(&ss.c))
	if ss.allocId > 0 {
		registry.Unregister(ss.allocId)
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

func NewSubscriber(topic PubSubTopic, init ReaderInit, iter ReaderIter, callback func(Packet)) (s *Subscriber, err error) {
	s = &Subscriber{}

	cTopic := topic.c()
	defer freeCPubSubTopic(cTopic)

	var activePktSpace []byte
	s.allocId = registry.Register(func(size C.size_t, out *C.a0_buf_t) C.a0_err_t {
		activePktSpace = make([]byte, int(size))
		out.size = size
		if size > 0 {
			out.data = (*C.uint8_t)(&activePktSpace[0])
		}
		return A0_OK
	})

	s.packetCallbackId = registry.Register(func(cPkt C.a0_packet_t) {
		callback(packetFromC(cPkt))
	})

	err = errorFrom(C.a0go_subscriber_init(&s.c, cTopic, C.uintptr_t(s.allocId), C.a0_reader_init_t(init), C.a0_reader_iter_t(iter), C.uintptr_t(s.packetCallbackId)))
	return
}

func (s *Subscriber) Close() (err error) {
	err = errorFrom(C.a0_subscriber_close(&s.c))
	registry.Unregister(s.packetCallbackId)
	if s.allocId > 0 {
		registry.Unregister(s.allocId)
	}
	return
}

func SubscriberReadOne(topic PubSubTopic, init ReaderInit, flags int) (pkt Packet, err error) {
	cTopic := topic.c()
	defer freeCPubSubTopic(cTopic)

	var pktSpace []byte
	allocId := registry.Register(func(size C.size_t, out *C.a0_buf_t) C.a0_err_t {
		pktSpace = make([]byte, int(size))
		out.size = size
		if size > 0 {
			out.data = (*C.uint8_t)(&pktSpace[0])
		}
		return A0_OK
	})
	defer registry.Unregister(allocId)

	cPkt := C.a0_packet_t{}
	err = errorFrom(C.a0go_subscriber_read_one(cTopic, C.uintptr_t(allocId), C.a0_reader_init_t(init), C.int(flags), &cPkt))
	pkt = packetFromC(cPkt)
	copy(pkt.Payload, pkt.Payload)
	return
}
