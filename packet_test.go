package alephzero

import "testing"

func TestNewPacket(t *testing.T) {
	pkt := NewPacket(nil, nil)
	if 36 != len(pkt.ID()) {
		t.Errorf("want: %v, got: %v", 36, len(pkt.ID()))
	}
}
