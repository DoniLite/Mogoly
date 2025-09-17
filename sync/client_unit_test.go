
package sync

import (
	"context"
	"testing"
	"time"
)

func TestClient_Send_WhenNotConnected(t *testing.T) {
	cl := NewClient()
	msg, _ := NewMessage(PING, nil, nil)
	if err := cl.Send(msg); err == nil {
		t.Fatalf("expected error when not connected")
	}
	_, err := cl.SendRequest(context.Background(), PING, nil, nil)
	if err == nil {
		t.Fatalf("expected error on SendRequest when not connected")
	}
	cl.Close() // should be safe when never connected
	// Close should be idempotent
	cl.Close()
	_ = time.Second
}
