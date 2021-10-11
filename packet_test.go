package alephzero

import "testing"

func TestNewPacket(t *testing.T) {
	pkt := NewPacket(nil, nil)
	if len(pkt.Id) != 36 {
		t.Errorf("Id failed: %s", pkt.Id)
	}

	pkt = NewPacket(nil, []byte("Hello, World!"))
	if string(pkt.Payload) != "Hello, World!" {
		t.Errorf("Payload failed: %s", pkt.Payload)
	}
}
