
package sync

import (
	"testing"
	"time"
)

func TestLogger_Send(t *testing.T) {
	ch := make(chan any, 1)
	SetLogger(ch)
	logf(LOG_INFO, "hello %d", 1)
	select {
	case v := <-ch:
		lg, ok := v.(Logs)
		if !ok {
			t.Fatalf("expected Logs, got %T", v)
		}
		if lg.Message == "" {
			t.Fatalf("empty message")
		}
	default:
		t.Fatalf("no log received")
	}

	// Non-blocking behavior when channel is full
	ch2 := make(chan any)
	SetLogger(ch2)
	done := make(chan struct{})
	go func() {
		logf(LOG_INFO, "should not block")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("logf blocked on full channel")
	}
}
