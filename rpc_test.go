package alephzero

import (
	"sync"
	"testing"
)

func TestRpc(t *testing.T) {
	FileRemove("test_rpc")
	topic := RpcTopic{"foo", nil}

	cnd := sync.NewCond(&sync.Mutex{})
	gotReply := false
	gotCancel := false

	rs, err := NewRpcServer(
		topic,
		func(req RpcRequest) {
			if string(req.Packet().Payload) == "reply" {
				req.Reply(NewPacket(nil, []byte("echo reply")))
			}
		},
		func(id string) {
			cnd.L.Lock()
			gotCancel = true
			cnd.Signal()
			cnd.L.Unlock()
		})
	check(t, err)
	defer rs.Close()

	rc, err := NewRpcClient(topic)
	check(t, err)
	defer rc.Close()

	check(t, rc.Send(NewPacket(nil, []byte("reply")), func(resp Packet) {
		cnd.L.Lock()
		gotReply = true
		cnd.Signal()
		cnd.L.Unlock()
	}))
	pkt := NewPacket(nil, []byte("cancel"))
	check(t, rc.Send(pkt, nil))
	check(t, rc.Cancel(pkt.Id))

	cnd.L.Lock()
	for !gotReply || !gotCancel {
		cnd.Wait()
	}
	cnd.L.Unlock()
}
