package sync

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Max time for any message writing
	writeWait = 10 * time.Second
	// Max time for the next peer message reading.
	pongWait = 60 * time.Second
	// Sending ping to the server after this period. Must be low than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Max message body.
	maxMessageSize = 8192 // Adjust if the body size can be consequent (ex: build spec)
)

type Connection struct {
	ws   *websocket.Conn
	send chan *Message // Channel for writing the i/o message
}

// creating a new connection struct.
func NewConnection(ws *websocket.Conn) *Connection {
	return &Connection{
		ws:   ws,
		send: make(chan *Message, 256),
	}
}

// fetching message from the channels 'send' to the WebSocket connection.
func (c *Connection) write(msgType int, payload []byte) error {
	err := c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		logf(LOG_ERROR, "write: Error setting write deadline: %v\n", err)
		return err
	}
	return c.ws.WriteMessage(msgType, payload)
}

// Handling sorting and periodical ping messages to the server
func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.ws.Close()
		if err != nil {
			logf(LOG_ERROR, "writePump: Error closing WebSocket connection: %v\n", err)
		}
		logf(LOG_INFO, "writePump: Stopped and closed WebSocket connection")
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				logf(LOG_INFO, "writePump: Send channel closed, closing connection.")
				err := c.write(websocket.CloseMessage, []byte{})
				if err != nil {
					logf(LOG_ERROR, "writePump: Error writing close message: %v\n", err)
				}
				return
			}

			err := c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				logf(LOG_ERROR, "write: Error setting write deadline: %v\n", err)
			}
			w, err := c.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				logf(LOG_ERROR, "writePump: Error getting next writer: %v\n", err)
				return
			}
			jsonBytes, err := json.Marshal(message)
			if err != nil {
				logf(LOG_ERROR,"writePump: Error marshaling message type %d: %v\n", message.Action.Type, err)
				// Don't return try to send the next message
				err := w.Close() // Close the actual writer
				if err != nil {
					logf(LOG_ERROR, "writePump: Error closing writer: %v\n", err)
				}
				continue
			}

			_, err = w.Write(jsonBytes)
			if err != nil {
				logf(LOG_ERROR, "writePump: Error writing JSON: %v\n", err)
				err := w.Close()
				if err != nil {
					logf(LOG_ERROR, "writePump: Error closing writer: %v\n", err)
				}
			}

			if err := w.Close(); err != nil {
				logf(LOG_ERROR, "writePump: Error closing writer: %v\n", err)
				return
			}
			logf(LOG_INFO, "writePump: Sent message type %d", message.Action.Type)

		case <-ticker.C:
			// Sending a periodical ping message
			logf(LOG_INFO, "writePump: Sending ping")
			if err := c.write(websocket.PingMessage, nil); err != nil {
				logf(LOG_ERROR, "writePump: Error sending ping: %v\n", err)
				return
			}
		}
	}
}

// Handling entering message
func (c *Connection) readPump(handler func(msg *Message, conn *Connection) error, disconnect func(conn *Connection)) {
	defer func() {
		disconnect(c)
		err := c.ws.Close()
		if err != nil {
			logf(LOG_ERROR, "readPump: Error closing WebSocket connection: %v\n", err)
		}
		logf(LOG_INFO, "readPump: Stopped and closed WebSocket connection")
	}()

	c.ws.SetReadLimit(maxMessageSize)
	err := c.ws.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		logf(LOG_ERROR, "readPump: Error setting read deadline: %v\n", err)
	}
	c.ws.SetPongHandler(func(string) error {
		logf(LOG_INFO, "readPump: Received pong")
		err = c.ws.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			logf(LOG_ERROR, "readPump: Error setting read deadline: %v\n", err)
		}
		return nil
	})

	for {
		msgType, messageBytes, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
				logf(LOG_ERROR, "readPump: WebSocket read error: %v\n", err)
			} else if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				logf(LOG_INFO, "readPump: WebSocket closed normally: %v\n", err)
			} else {
				logf(LOG_ERROR, "readPump: Unhandled WebSocket read error: %v\n", err)
			}
			break
		}

		// Ignore non text message
		if msgType != websocket.TextMessage {
			logf(LOG_INFO, "readPump: Received non-text message type: %d\n", msgType)
			continue
		}

		logf(LOG_INFO, "readPump: Received raw message: %s", string(messageBytes)) // Debug

		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			logf(LOG_ERROR, "readPump: Error unmarshaling message: %v --- Raw: %s\n", err, string(messageBytes))
			errMsg := NewErrorMessage("Invalid message format", err.Error())
			c.send <- errMsg
			continue
		}

		if err := handler(&msg, c); err != nil {
			logf(LOG_ERROR, "readPump: Error handling message type %d: %v\n", msg.Action.Type, err)
			errMsg := NewErrorMessage("Failed to handle request", err.Error())
			c.send <- errMsg
		}

		err = c.ws.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			logf(LOG_ERROR, "readPump: Error setting read deadline: %v\n", err)
		}
	}
}

// sending message asynchronously via the websocket.
func (c *Connection) SendMsg(msg *Message) {
	select {
	case c.send <- msg:
	default:
		logf(LOG_INFO, "Warning: Send channel full for connection %p. Message type %d dropped.\n", c.ws, msg.Action.Type)
	}
}

// closing the send channel and stopping the writePump function.
func (c *Connection) CloseSend() {
	close(c.send)
}
