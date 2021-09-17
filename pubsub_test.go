package alephzero

import (
	"fmt"
	"sort"
	"sync"
	"testing"
)

func TestPubSub(t *testing.T) {
	FileRemove("alephzero/foo.pubsub.a0")
	topic := PubSubTopic{"foo", nil}

	p, err := NewPublisher(topic)
	check(t, err)
	defer p.Close()
	ss, err := NewSubscriberSync(topic, INIT_OLDEST, ITER_NEXT)
	check(t, err)
	defer ss.Close()

	cnd := sync.NewCond(&sync.Mutex{})
	allPayloads := [][]byte{}

	s, err := NewSubscriber(topic, INIT_OLDEST, ITER_NEXT, func(pkt Packet) {
		cnd.L.Lock()
		allPayloads = append(allPayloads, pkt.Payload)
		cnd.Signal()
		cnd.L.Unlock()
	})
	check(t, err)
	defer s.Close()

	if hasNext, err := ss.HasNext(); err != nil || hasNext {
		t.Error("HasNext() should be false")
	}
	p.Pub(NewPacket(nil, []byte("hello")))
	if hasNext, err := ss.HasNext(); err != nil || !hasNext {
		t.Error("HasNext() should be true")
	}
	pkt, err := ss.Next()
	check(t, err)
	if string(pkt.Payload) != "hello" {
		t.Error("Payload() should be 'hello'")
	}

	hdrKeys := make([]string, 0, len(pkt.Headers))
	for k := range pkt.Headers {
		hdrKeys = append(hdrKeys, k)
	}
	sort.Strings(hdrKeys)
	if fmt.Sprint(hdrKeys) != "[a0_time_mono a0_time_wall a0_transport_seq a0_writer_id a0_writer_seq]" {
		t.Error("Headers() should be [a0_time_mono a0_time_wall a0_transport_seq a0_writer_id a0_writer_seq]")
	}

	if hasNext, err := ss.HasNext(); err != nil || hasNext {
		t.Error("HasNext() should be false")
	}

	cnd.L.Lock()
	for len(allPayloads) < 1 {
		cnd.Wait()
	}
	cnd.L.Unlock()

	if len(allPayloads) != 1 {
		t.Error("should have received 1 packet")
	}
	if string(allPayloads[0]) != "hello" {
		t.Error("payload 0 should be 'hello'")
	}
}
