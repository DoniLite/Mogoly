package server

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func parseServerURL(s *Server) (*url.URL, error) {
	if s.URL != "" {
		u, err := url.Parse(s.URL)
		if err == nil {
			return u, nil
		}
	}
	built, err := BuildServerURL(s)
	if err != nil {
		return nil, err
	}
	return url.Parse(built)
}

func singleSlashJoin(a, b string) string {
	slashA := strings.HasSuffix(a, "/")
	prefixB := strings.HasPrefix(b, "/")
	switch {
	case slashA && prefixB:
		return a + strings.TrimPrefix(b, "/")
	case !slashA && !prefixB:
		return a + "/" + b
	default:
		return a + b
	}
}

func appendForwardHeaders(h http.Header, r *http.Request, scheme string) {
	ip := clientIP(r.RemoteAddr)
	if prior := h.Get("X-Forwarded-For"); prior != "" {
		h.Set("X-Forwarded-For", prior+", "+ip)
	} else {
		h.Set("X-Forwarded-For", ip)
	}
	h.Set("X-Forwarded-Proto", scheme)
	if r.Host != "" {
		h.Set("X-Forwarded-Host", r.Host)
	}
}

func clientIP(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}

func BuildServerURL(server *Server) (string, error) {
	if server == nil {
		return "", fmt.Errorf("nil server")
	}
	if server.URL != "" {
		if _, err := url.Parse(server.URL); err == nil {
			return server.URL, nil
		}
	}
	if server.Protocol == "" || server.Host == "" || server.Port == 0 {
		return "", fmt.Errorf("incomplete server fields for URL (need protocol, host, port)")
	}
	return fmt.Sprintf("%s://%s:%d", server.Protocol, server.Host, server.Port), nil
}
