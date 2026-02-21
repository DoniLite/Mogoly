package router

import (
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/DoniLite/Mogoly/core/domain"
	"github.com/DoniLite/Mogoly/core/events"
)

func routeHandler(w http.ResponseWriter, r *http.Request) {
	rs, err := GetRouter()
	if err != nil {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Router not ready yet")
		http.Error(w, "router not ready", http.StatusServiceUnavailable)
		return
	}
	b, ok := rs.httpServerMap[strings.ToLower(r.Host)]
	if !ok {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Not found route for %s", strings.ToLower(r.Host))
		http.NotFound(w, r)
		return
	}
	b.ServeHTTP(w, r)
}

func httpEntry(w http.ResponseWriter, r *http.Request) {
	rs, err := GetRouter()
	if err != nil {
		events.Logf(events.LOG_ERROR, "[ROUTER]: Router not ready yet")
		http.Error(w, "router not ready", http.StatusServiceUnavailable)
		return
	}
	if b, ok := rs.serverMap[strings.ToLower(r.Host)]; ok && b.ForceTLS {
		url := *r.URL
		url.Scheme = "https"
		url.Host = r.Host
		events.Logf(events.LOG_INFO, "[ROUTER]: Redirecting new incoming request from host: %s to https url: %s", strings.ToLower(r.Host), url.String())
		http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
		return
	}
	routeHandler(w, r)
}

func ServeHTTP(addr string) *http.Server {
	hs := &http.Server{Addr: addr, Handler: http.HandlerFunc(httpEntry)}
	go func() {
		events.Logf(events.LOG_INFO, "[HTTP_SERVER]: HTTP listening on %s", addr)
		if err := hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			events.Logf(events.LOG_ERROR, "[HTTP_SERVER]: http server error: %v", err)
			os.Exit(1)
		}
	}()
	return hs
}

func ServeHTTPS(addr string, cm *domain.Manager) *http.Server {
	ts := &http.Server{Addr: addr, Handler: http.HandlerFunc(routeHandler)}
	ts.TLSConfig = &tls.Config{GetCertificate: cm.GetCertificate, MinVersion: tls.VersionTLS12}
	// Create listener *first* so we can expose the effective addr (when :0 was requested).
	ln, err := tls.Listen("tcp", addr, ts.TLSConfig)
	if err != nil {
		events.Logf(events.LOG_ERROR, "[HTTPS_SERVER]: https server error: %v", err)
		os.Exit(1)
	}
	// Publish the effective addr (e.g., 127.0.0.1:51327) for tests and callers.
	ts.Addr = ln.Addr().String()
	events.Logf(events.LOG_INFO, "[HTTPS_SERVER]: HTTPS listening on %s", ts.Addr)
	go func() {
		if err := ts.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			events.Logf(events.LOG_ERROR, "[HTTPS_SERVER]: https server error: %v", err)
			os.Exit(1)
		}
	}()
	return ts
}
