package events

func (ch *LogChannel) Write(p []byte) (int, error) {
	msg := string(p)
	for {
		select {
		case ch.ch <- msg:
			// Success
			return len(p), nil
		default:
			// Channel full, try to remove oldest
			select {
			case <-ch.ch:
				// Removed oldest
			default:
				// Channel empty? Loop back to try send
			}
		}
	}
}
