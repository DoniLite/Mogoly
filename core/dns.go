package core

import (
	"log"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"

	"github.com/DoniLite/Mogoly/core/events"
)

func (d *DNSServer) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true
	for _, q := range r.Question {
		name := strings.TrimSuffix(strings.ToLower(q.Name), ".")
		switch q.Qtype {
		case dns.TypeA:
			if d.isLocal(name) {
				events.Logf(events.LOG_INFO, "[DNS]: new local name resolved %s, Type A", name)
				m.Answer = append(m.Answer, &dns.A{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 5}, A: net.ParseIP("127.0.0.1")})
			} else {
				events.Logf(events.LOG_INFO, "[DNS]: new public name resolved %s, Type A", name)
				d.forward(name, q.Qtype, m)
			}
		case dns.TypeAAAA:
			if d.isLocal(name) {
				events.Logf(events.LOG_INFO, "[DNS]: new local name resolved %s, Type AAAA", name)
				m.Answer = append(m.Answer, &dns.AAAA{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: 5}, AAAA: net.ParseIP("::1")})
			} else {
				events.Logf(events.LOG_INFO, "[DNS]: new public name resolved %s, Type AAAA", name)
				d.forward(name, q.Qtype, m)
			}
		default:
			// minimal: forward others
			events.Logf(events.LOG_INFO, "[DNS]: unknown name resolved %s", name)
			d.forward(name, q.Qtype, m)
		}
	}
	_ = w.WriteMsg(m)
}

func (d *DNSServer) forward(name string, qtype uint16, m *dns.Msg) {
	var resp *dns.Msg
	var err error
	events.Logf(events.LOG_INFO, "[DNS]: Initializing forwarding")
	if d.forwardTo == "" {
		d.forwardTo = "1.1.1.1:53"
		events.Logf(events.LOG_INFO, "[DNS]: No `forwardTo` field found using fallback: %s", d.forwardTo)
	}
	c := new(dns.Client)
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(name), qtype)
	if resp, _, err = c.Exchange(msg, d.forwardTo); err == nil && resp != nil {
		events.Logf(events.LOG_INFO, "[DNS]: New answer from the dns %s", resp.String())
		m.Answer = append(m.Answer, resp.Answer...)
		return
	}
	events.Logf(events.LOG_ERROR, "[DNS]: Something went wrong during dns forwarding possible errors \nerror: %s \nresponse: %s", err.Error(), resp.String())
}

func ServeDNS(bind string, isLocal func(string) bool, forwardTo string) {
	s := &DNSServer{isLocal: isLocal, forwardTo: forwardTo}
	dns.HandleFunc(".", s.ServeDNS)
	udpServer := &dns.Server{Addr: bind, Net: "udp"}
	tcpServer := &dns.Server{Addr: bind, Net: "tcp"}
	go func() {
		events.Logf(events.LOG_INFO, "[DNS]: (upd) listening on %s", bind)
		err := udpServer.ListenAndServe()
		if err != nil {
			events.Logf(events.LOG_ERROR, "[DNS]: (udp) dns error during server starting: %s", err.Error())
			os.Exit(1)
		}
	}()
	go func() {
		events.Logf(events.LOG_INFO, "[DNS]: (tcp) listening on %s", bind)
		err := tcpServer.ListenAndServe()
		if err != nil {
			events.Logf(events.LOG_ERROR, "[DNS]: (tcp) dns error during server starting %s", err.Error())
			os.Exit(1)
		}
	}()
}

// Starts UDP+TCP DNS servers and returns a stop function.
func StartDNSServer(bind string, isLocal func(string) bool, forwardTo string) (stop func(), err error) {
	s := &DNSServer{isLocal: isLocal, forwardTo: forwardTo}
	mux := dns.NewServeMux()
	mux.HandleFunc(".", s.ServeDNS)
	udpServer := &dns.Server{Addr: bind, Net: "udp", Handler: mux}
	tcpServer := &dns.Server{Addr: bind, Net: "tcp", Handler: mux}
	go func() {
		log.Printf("DNS(udp) listening on %s", bind)
		_ = udpServer.ListenAndServe()
	}()
	go func() {
		log.Printf("DNS(tcp) listening on %s", bind)
		_ = tcpServer.ListenAndServe()
	}()
	return func() {
		_ = udpServer.Shutdown()
		_ = tcpServer.Shutdown()
	}, nil
}
