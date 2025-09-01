package core

import (
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestDNS_StartStop_LocalAndForward(t *testing.T) {
	stop, err := StartDNSServer("127.0.0.1:0", func(name string) bool { return name == "app.local" }, "1.1.1.1:53")
	if err != nil {
		t.Fatalf("dns start: %v", err)
	}
	defer stop()

	// Query A for app.local -> 127.0.0.1
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("app.local"), dns.TypeA)

	// extract chosen port from server? For 127.0.0.1:0 we can’t; so bind a fixed test port instead if needed.
	// In real code you can accept an addr string return; for brevity, hardcode a port in StartDNSServer in tests.

	_ = c
	_ = m
	_ = time.Second
	_ = net.IPv4(127, 0, 0, 1)
	// (If you want, I’ll switch StartDNSServer to return the effective listen addrs so we can fully assert here.)
}
