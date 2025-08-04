package sync

import (
	"encoding/json"
	"log"
	"net/http"
)


// Server
func (s *Server) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	for {
		var msg Message
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Read error:", err)
			break
		}
		s.onMsg(conn, msg)
	}
}

func (s *Server) Start(addr string) error {
	http.HandleFunc("/ws", s.Handle)
	log.Println("WebSocket server listening on", addr)
	return http.ListenAndServe(addr, nil)
}

// Client
func (c *Client) Send(msg Message) error {
	return c.conn.WriteJSON(msg)
}

func (c *Client) Listen(onMsg func(msg Message)) {
	defer c.conn.Close()

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Println("Invalid JSON:", err)
			continue
		}
		onMsg(msg)
	}
}
