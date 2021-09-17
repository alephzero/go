package alephzero

import (
	"sync"
	"testing"
)

func TestPrpc(t *testing.T) {
	FileRemove("alephzero/foo.prpc.a0")
	topic := PrpcTopic{"foo", nil}

	cnd := sync.NewCond(&sync.Mutex{})
	numProg := 0
	numDone := 0
	numCancel := 0

	ps, err := NewPrpcServer(
		topic,
		func(conn PrpcConnection) {
			if string(conn.Packet().Payload) == "connect" {
				conn.Send(NewPacket(nil, []byte("progress")), false)
				conn.Send(NewPacket(nil, []byte("progress")), false)
				conn.Send(NewPacket(nil, []byte("progress")), false)
				conn.Send(NewPacket(nil, []byte("progress")), false)
				conn.Send(NewPacket(nil, []byte("progress")), true)
			}
		},
		func(id string) {
			cnd.L.Lock()
			numCancel++
			cnd.Signal()
			cnd.L.Unlock()
		})
	check(t, err)
	defer ps.Close()

	pc, err := NewPrpcClient(topic)
	check(t, err)
	defer pc.Close()

	check(t, pc.Connect(NewPacket(nil, []byte("connect")), func(resp Packet, done bool) {
		cnd.L.Lock()
		numProg++
		if done {
			numDone++
		}
		cnd.Signal()
		cnd.L.Unlock()
	}))
	pkt := NewPacket(nil, []byte("cancel"))
	check(t, pc.Connect(pkt, nil))
	check(t, pc.Cancel(pkt.Id))

	cnd.L.Lock()
	for numProg != 5 && numDone != 1 && numCancel != 1 {
		cnd.Wait()
	}
	cnd.L.Unlock()
}
