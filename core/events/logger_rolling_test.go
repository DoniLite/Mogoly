package events

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestLoggerRolling verifies that the logger discards oldest messages when full,
// keeping the most recent ones (Circular Buffer behavior).
func TestLoggerRolling(t *testing.T) {
	bufferSize := 5
	// Re-init logger with small buffer
	logChannel = LogChannel{ch: make(chan any, bufferSize)}
	logger = &Logger{writer: &logChannel}

	// Write more items than buffer size
	totalItems := 10
	for i := range totalItems {
		Logf(LOG_INFO, "msg %d", i)
	}

	// We expect the channel to contain items: 5, 6, 7, 8, 9
	// (Assuming "msg %d" format)

	// Consume channel
	var received []string
	timeout := time.After(1 * time.Second)

	for i := range bufferSize {
		select { // non-blocking read
		case msg := <-logChannel.ch:
			// msg is a JSON string or formatted string?
			// Logf implementation: "[timestamp] msg %d"
			// We just check if it contains the expected number.
			s, ok := msg.(string)
			if !ok {
				t.Fatalf("Expected string, got %T", msg)
			}
			received = append(received, s)
		case <-timeout:
			t.Fatalf("Timeout waiting for message %d", i)
		}
	}

	// Verify we got the LATEST items
	// The first item in 'received' should be "msg 5" (if we dropped 0-4)
	// Actually, let's just check the content.

	// Since Logf adds a timestamp, we verify suffix
	expectedStart := totalItems - bufferSize // 10 - 5 = 5
	for i, msg := range received {
		expectedContent := fmt.Sprintf("msg %d", expectedStart+i)
		// Simple contains check
		if !strings.Contains(msg, expectedContent) {
			t.Errorf("Index %d: Expected message containing '%s', got '%s'", i, expectedContent, msg)
		}
	}

	// Ensure channel is empty now
	select {
	case <-logChannel.ch:
		t.Error("Channel should be empty after reading buffer size")
	default:
		// good
	}
}
