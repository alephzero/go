package alephzero

import (
	"testing"
	"time"
)

func TestHeartbeat(t *testing.T) {
	FileRemove("test_heartbeat")

	file, err := FileOpen("test_heartbeat", nil)
	if err != nil {
		t.Errorf("FileOpen(\"test_heartbeat\") failed with %v", err)
	}

	h, err := NewHeartbeat(file, nil)
	time.Sleep(time.Second)
	h.Close()

	ss, err := NewSubscriberSync(file, INIT_OLDEST, ITER_NEXT)
	if err != nil {
		t.Errorf("NewSubscriberSync failed with %v", err)
	}

	cnt := 0
	for {
		if hasNext, err := ss.HasNext(); err != nil {
			t.Errorf("SubscriberSync.HasNext() failed with %v", err)
		} else if hasNext {
			ss.Next()
			cnt++
		} else {
			break
		}
	}

	if cnt <= 8 {
		t.Errorf("Expected 10 heartbeats, only recieved %v", cnt)
	}
}
