
package sync

import (
	"encoding/json"
	"testing"
)

func TestAction_AddPayload_ThenDeserialize(t *testing.T) {
	type P struct{ X string `json:"x"` }
	a := &Action{Type: CREATE_SERVER}
	if err := a.AddPayload(P{X: "ok"}); err != nil {
		t.Fatalf("AddPayload error: %v", err)
	}
	var back P
	if err := a.Deserialize(&back); err != nil {
		t.Fatalf("Deserialize error: %v", err)
	}
	if back.X != "ok" {
		t.Fatalf("roundtrip mismatch: %#v", back)
	}
}

func TestMessage_DecodePayload_ErrOnEmpty(t *testing.T) {
	m := &Message{Action: Action{Type: ADD_SERVER}}
	var out map[string]any
	if err := m.DecodePayload(&out); err == nil {
		t.Fatalf("expected error for empty payload")
	}
}

func TestNewMessage_WithPayloadAndMeta(t *testing.T) {
	type P struct{ A int `json:"a"` }
	type M struct{ Z string `json:"z"` }

	msg, err := NewMessage(ADD_SERVER, P{A: 3}, M{Z: "m"})
	if err != nil {
		t.Fatalf("NewMessage error: %v", err)
	}
	if msg.Action.Type != ADD_SERVER {
		t.Fatalf("wrong type: %v", msg.Action.Type)
	}
	var p P
	if err := json.Unmarshal(msg.Action.Payload, &p); err != nil {
		t.Fatalf("payload unmarshal: %v", err)
	}
	if p.A != 3 {
		t.Fatalf("want 3 got %d", p.A)
	}
	var md M
	if err := json.Unmarshal(msg.Meta, &md); err != nil {
		t.Fatalf("meta unmarshal: %v", err)
	}
	if md.Z != "m" {
		t.Fatalf("meta roundtrip mismatch: %#v", md)
	}
}
