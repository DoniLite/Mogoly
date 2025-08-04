package sync

import (
	"net/http"

	"github.com/gorilla/websocket"
)

func NewClient(url string) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn}, nil
}


func NewServer(onMsg HandlerFunc) *Server {
	return &Server{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		onMsg: onMsg,
	}
}