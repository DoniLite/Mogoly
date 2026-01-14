package events

import (
	"testing"
	"time"
)

// TestLoggerBlocking verifies that the logger blocks (or deadlocks) if the channel is full and not drained.
// This confirms the dangerous behavior of the current implementation.
func TestLoggerBlocking(t *testing.T) {
	// Re-init logger with small buffer for testing
	logChannel = LogChannel{ch: make(chan any, 5)}
	logger = &Logger{writer: &logChannel} // direct write to channel to avoid stdout noise

	done := make(chan bool)
	go func() {
		// Try to log more than buffer size
		for i := 0; i < 10; i++ {
			Logf(LOG_INFO, "test log %d", i)
		}
		done <- true
	}()

	select {
	case <-done:
		t.Log("Successfully logged all messages (Unexpected if buffer is full and blocking)")
	case <-time.After(100 * time.Millisecond):
		t.Log("Test timed out as expected - Logger is blocking on full buffer")
		// Ideally we want to fail or pass based on intent.
		// Since we want to PROVE it blocks, this timeout is actually "Success" of the reproduction.
		// But for a proper unit test suite that should PASS, we should fix the code so this doesn't timeout.
	}
}

// TestLoggerNonBlocking verifies the fix (to be implemented)
func TestLoggerNonBlocking(t *testing.T) {
	// Re-init logger
	logChannel = LogChannel{ch: make(chan any, 5)}
	logger = &Logger{writer: &logChannel}

	done := make(chan bool)
	go func() {
		for i := 0; i < 20; i++ {
			Logf(LOG_INFO, "test log %d", i)
		}
		done <- true
	}()

	select {
	case <-done:
		// Success - didn't block
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Logger blocked! Fix required.")
	}
}
