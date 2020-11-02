package alephzero

import (
	"sync"
	"testing"
)

func TestRpc(t *testing.T) {
	FileRemove("test_rpc")

	file, err := FileOpen("test_rpc", nil)
	if err != nil {
		t.Errorf("FileOpen(\"test_rpc\") failed with %v", err)
	}

	rs, err := NewRpcServer(file, func(req RpcRequest) {
		if string(req.Packet().Payload) != "foo" {
			t.Errorf("Server expected a request with message 'foo', got %v", req.Packet().Payload)
		}
		req.Reply(NewPacket(nil, []byte("bar")))
	}, nil)
	if err != nil {
		t.Errorf("NewRpcServer failed with %v", err)
	}
	defer rs.Close()

	rc, err := NewRpcClient(file)
	if err != nil {
		t.Errorf("NewRpcClient failed with %v", err)
	}
	defer rs.Close()

	mu := sync.Mutex{}
	mu.Lock()
	err = rc.Send(NewPacket(nil, []byte("foo")), func(resp Packet) {
		if string(resp.Payload) != "bar" {
			t.Errorf("Client expected a response with message 'bar', got %v", resp.Payload)
		}
		mu.Unlock()
	})
	if err != nil {
		t.Errorf("RpcClient.Send failed with %v", err)
	}
	mu.Lock()
}
