package sync

import "github.com/gorilla/websocket"

type Client struct {
	conn *websocket.Conn
}

type HandlerFunc func(conn *websocket.Conn, msg Message)

type Server struct {
	upgrader websocket.Upgrader
	onMsg    HandlerFunc
}

type Message struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

type RunTaskPayload struct {
	TaskName string `json:"task_name"`
}

type TaskStatusPayload struct {
	TaskName string `json:"task_name"`
	Status   string `json:"status"`
}