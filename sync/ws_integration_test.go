package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// simple echo handler: respond with same payload and request id
func echoHandler(msg *Message, conn *Connection) error {
	var payload map[string]any
	_ = json.Unmarshal(msg.Action.Payload, &payload)
	resp := &Message{
		RequestID: msg.RequestID,
		Action: Action{
			Type:    msg.Action.Type,
			Payload: msg.Action.Payload,
		},
	}
	conn.SendMsg(resp)
	return nil
}

func TestClientServer_RequestResponse(t *testing.T) {
	// server
	srv := NewServer(echoHandler, func(r *http.Request) bool { return true })
	srv.Run()
	httpSrv := httptest.NewServer(http.HandlerFunc(srv.ServeHTTP))
	defer httpSrv.Close()

	// client
	cl := NewClient()
	defer cl.Close()
	if err := cl.Connect("ws"+httpSrv.URL[4:], nil); err != nil {
		t.Fatalf("connect: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	payload := map[string]any{"k": "v"}
	resp, err := cl.SendRequest(ctx, ADD_SERVER, payload, nil)
	if err != nil {
		t.Fatalf("send request: %v", err)
	}
	var back map[string]any
	if err := resp.DecodePayload(&back); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if back["k"] != "v" {
		t.Fatalf("unexpected payload: %#v", back)
	}
}
