package alephzero

import (
	"testing"
	"time"
)

func TestHeartbeat(t *testing.T) {
	ShmUnlink("/test_heartbeat")

	shm, err := ShmOpen("/test_heartbeat", nil)
	if err != nil {
		t.Errorf("ShmOpen(\"/test_heartbeat\") failed with %v", err)
	}

	h, err := NewHeartbeat(shm, nil)
	time.Sleep(time.Second)
	h.Close()

	ss, err := NewSubscriberSync(shm, INIT_OLDEST, ITER_NEXT)
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
