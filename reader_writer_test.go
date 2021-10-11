package alephzero

import (
	"fmt"
	"sync"
	"testing"
)

func TestReaderWriter(t *testing.T) {
	FileRemove("foo")

	file, err := FileOpen("foo", nil)
	check(t, err)
	defer file.Close()

	w, err := NewWriter(file.Arena())
	check(t, err)
	defer w.Close()
	rs, err := NewReaderSync(file.Arena(), INIT_OLDEST, ITER_NEXT)
	check(t, err)
	defer rs.Close()

	cnd := sync.NewCond(&sync.Mutex{})

	allPayloads := [][]byte{}

	r, err := NewReader(file.Arena(), INIT_OLDEST, ITER_NEXT, func(pkt Packet) {
		cnd.L.Lock()
		allPayloads = append(allPayloads, pkt.Payload)
		cnd.Signal()
		cnd.L.Unlock()
	})
	check(t, err)
	defer r.Close()

	if hasNext, err := rs.HasNext(); err != nil || hasNext {
		t.Error("HasNext() should be false")
	}
	check(t, w.Write(NewPacket(nil, []byte("hello"))))
	if hasNext, err := rs.HasNext(); err != nil || !hasNext {
		t.Error("HasNext() should be true")
	}
	pkt, err := rs.Next()
	check(t, err)
	if string(pkt.Payload) != "hello" {
		t.Error("Payload() should be 'hello'")
	}
	if hasNext, err := rs.HasNext(); err != nil || hasNext {
		t.Error("HasNext() should be false")
	}

	check(t, w.Push(AddTransportSeqHeader()))
	w2, err := w.Wrap(AddWriterSeqHeader())
	check(t, err)

	check(t, w.Write(NewPacket(nil, []byte("aaa"))))
	check(t, w2.Write(NewPacket(nil, []byte("bbb"))))

	pkt, err = rs.Next()
	check(t, err)
	if string(pkt.Payload) != "aaa" {
		t.Error("Payload() should be 'aaa'")
	}
	if fmt.Sprint(pkt.Headers) != "map[a0_transport_seq:[1]]" {
		t.Error("Headers() should be 'map[a0_transport_seq:[1]]'")
	}

	pkt, err = rs.Next()
	check(t, err)
	if string(pkt.Payload) != "bbb" {
		t.Error("Payload() should be 'bbb'")
	}
	if fmt.Sprint(pkt.Headers) != "map[a0_transport_seq:[2] a0_writer_seq:[0]]" {
		t.Error("Headers() should be 'map[a0_transport_seq:[2] a0_writer_seq:[0]]'")
	}

	cnd.L.Lock()
	for len(allPayloads) < 3 {
		cnd.Wait()
	}
	cnd.L.Unlock()

	if len(allPayloads) != 3 {
		t.Error("should have received 3 packet")
	}
	if string(allPayloads[0]) != "hello" {
		t.Error("payload 0 should be 'hello'")
	}
	if string(allPayloads[1]) != "aaa" {
		t.Error("payload 1 should be 'aaa'")
	}
	if string(allPayloads[2]) != "bbb" {
		t.Error("payload 2 should be 'bbb'")
	}
}
