package sync

import (
	"sync"
)

type Hub struct {
	clients    map[*Connection]bool // List of connection registered
	register   chan *Connection     // Channel for connection registration
	unregister chan *Connection     // Channel for connection removing
	broadcast  chan *Message        // Diffusing message for all registered instance

	mu sync.RWMutex

	// Handler for incoming message
	messageHandler func(msg *Message, client *Connection) error
}

func newHub(handler func(msg *Message, client *Connection) error) *Hub {
	return &Hub{
		clients:        make(map[*Connection]bool),
		register:       make(chan *Connection),
		unregister:     make(chan *Connection),
		broadcast:      make(chan *Message),
		messageHandler: handler,
	}
}

// handling registration/removing clients connection asynchronously
func (h *Hub) run() {
	logf(LOG_INFO, "Hub: Starting run loop")
	for {
		select {
		case conn := <-h.register:
			h.mu.Lock()
			h.clients[conn] = true
			h.mu.Unlock()
			logf(LOG_INFO, "Hub: Client registered (%p). Total clients: %d\n", conn.ws, len(h.clients))

		case conn := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.CloseSend()
				logf(LOG_INFO, "Hub: Client unregistered (%p). Total clients: %d\n", conn.ws, len(h.clients))
			} else {
				logf(LOG_INFO, "Hub: Unregister request for non-existent client (%p)\n", conn.ws)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for conn := range h.clients {
				select {
				case conn.send <- message:
				default:
					logf(LOG_ERROR, "Hub: Broadcast failed for client %p, closing its send channel.\n", conn.ws)
					close(conn.send)
					delete(h.clients, conn)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Calling this handler if a connection is disconnected
func (h *Hub) handleDisconnect(conn *Connection) {
	h.unregister <- conn
}

// handler passed to readPump for incoming message.
func (h *Hub) handleIncomingMessage(msg *Message, conn *Connection) error {
	if h.messageHandler != nil {
		return h.messageHandler(msg, conn)
	}
	logf(LOG_ERROR, "Hub: No message handler configured, dropping message type %d from %p\n", msg.Action.Type, conn.ws)
	return nil
}
