package core

import "testing"

func TestRoundRobinPrefersHealthy(t *testing.T) {
	a := &Server{Name: "a", URL: "http://127.0.0.1:1", IsHealthy: false}
	b := &Server{Name: "b", URL: "http://127.0.0.1:2", IsHealthy: true}
	c := &Server{Name: "c", URL: "http://127.0.0.1:3", IsHealthy: false}

	lb := &Server{BalancingServers: []*Server{a, b, c}}
	got, err := lb.GetNextServer()
	if err != nil || got.Name != "b" {
		t.Fatalf("want b, got %v err %v", got, err)
	}
	// Walk around the ring; b should keep being returned while others unhealthy
	for range 5 {
		got, _ = lb.GetNextServer()
		if got.Name != "b" {
			t.Fatalf("want b, got %s", got.Name)
		}
	}
	// Mark a healthy; next should eventually return a
	a.IsHealthy = true
	_, _ = lb.GetNextServer() // b -> next
	got, _ = lb.GetNextServer()
	if got.Name != "a" && got.Name != "b" { // either b then a depending on idx
		t.Fatalf("unexpected rr pick: %s", got.Name)
	}
}
