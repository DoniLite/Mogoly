
package sync

import (
	"testing"
	"time"
)

func TestHub_RegisterBroadcastUnregister(t *testing.T) {
	h := newHub(nil)
	go h.run()

	// fake connection with only send channel
	c := &Connection{send: make(chan *Message, 1)}
	h.register <- c

	// broadcast one message
	msg := &Message{Action: Action{Type: ADD_SERVER}}
	h.broadcast <- msg

	select {
	case got := <-c.send:
		if got.Action.Type != ADD_SERVER {
			t.Fatalf("unexpected message: %#v", got)
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting broadcast")
	}

	// unregister
	h.unregister <- c
}
