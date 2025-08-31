// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	currentRouter *RouterState
)

var (
	requestsPerMinute = 5
	rateLimitWindow   = time.Minute

	// map[ip] = []timestamps
	visitors = make(map[string][]time.Time)
	mu       sync.Mutex
)

func BuildRouter(config *Config) {
	rs := &RouterState{
		m: make(map[string]http.Handler),
		s: make(map[string]*Server),
	}
	for _, server := range config.Servers {
		rs.m[strings.ToLower(server.Name)] = createSingleHttpServer(server)
		rs.s[strings.ToLower(server.Name)] = server
	}

	rs.globalConfig = config
	currentRouter = rs
}

func GetRouter() *RouterState {
	return currentRouter
}

func createSingleHttpServer(s *Server) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.ServeHTTP)

	var middlewares []func(http.Handler) http.Handler

	for _, v := range s.Middlewares {
		m := MiddlewaresList[MiddleWareName(v.Name)]

		middlewares = append(middlewares, m.Fn(v.Config))
	}

	return ChainMiddleware(mux, middlewares...)
}

// ping returns a "pong" message consider registering this Handler for the health checking logic
func Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("pong"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

func HealthChecker(server *Server) (bool, error) {
	raw, err := buildServerURL(server)
	if err != nil {
		return false, err
	}
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		return false, err
	}
	client := &http.Client{Timeout: 3 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))
		return false, fmt.Errorf("server responded %d: %s", res.StatusCode, string(body))
	}
	return true, nil
}

type RateLimitMiddlewareConfig struct {
	ReqPerMinute int           `json:"request_per_minute,omitempty" yaml:"request_per_minute,omitempty"`
	LimitWindow  time.Duration `json:"limit_window,omitempty" yaml:"limit_window,omitempty"`
}

func RateLimiterMiddleware(config any) func(next http.Handler) http.Handler {
	conf := RateLimitMiddlewareConfig{ReqPerMinute: requestsPerMinute, LimitWindow: rateLimitWindow}
	switch v := config.(type) {
	case *RateLimitMiddlewareConfig:
		if v != nil {
			conf = *v
		}
	case RateLimitMiddlewareConfig:
		conf = v
	case map[string]any:
		if rpm, ok := v["request_per_minute"]; ok {
			switch x := rpm.(type) {
			case float64:
				conf.ReqPerMinute = int(x)
			case int:
				conf.ReqPerMinute = x
			}
		}
		if lw, ok := v["limit_window"]; ok {
			switch x := lw.(type) {
			case string:
				if d, err := time.ParseDuration(x); err == nil {
					conf.LimitWindow = d
				}
			case float64:
				conf.LimitWindow = time.Duration(int64(x)) * time.Second
			}
		}
	}
	if conf.ReqPerMinute <= 0 {
		conf.ReqPerMinute = 5
	}
	if conf.LimitWindow <= 0 {
		conf.LimitWindow = time.Minute
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			mu.Lock()
			defer mu.Unlock()

			now := time.Now()
			keep := visitors[ip][:0]
			for _, t := range visitors[ip] {
				if now.Sub(t) < conf.LimitWindow {
					keep = append(keep, t)
				}
			}
			if len(keep) >= conf.ReqPerMinute {
				http.Error(w, "Max request exceed", http.StatusTooManyRequests)
				return
			}

			visitors[ip] = append(keep, now)

			next.ServeHTTP(w, r)
		})
	}
}

func CleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, timestamps := range visitors {
			var filtered []time.Time
			for _, t := range timestamps {
				if time.Since(t) < rateLimitWindow {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				delete(visitors, ip)
			} else {
				visitors[ip] = filtered
			}
		}
		mu.Unlock()
	}
}

func ChainMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

func routeHandler(w http.ResponseWriter, r *http.Request) {
	rs := GetRouter()
	if rs == nil {
		http.Error(w, "router not ready", http.StatusServiceUnavailable)
		return
	}
	b, ok := rs.m[strings.ToLower(r.Host)]
	if !ok {
		http.NotFound(w, r)
		return
	}
	b.ServeHTTP(w, r)
}

func httpEntry(w http.ResponseWriter, r *http.Request) {
	rs := GetRouter()
	if rs == nil {
		http.Error(w, "router not ready", http.StatusServiceUnavailable)
		return
	}
	if b, ok := rs.s[strings.ToLower(r.Host)]; ok && b.forceTLS {
		url := *r.URL
		url.Scheme = "https"
		url.Host = r.Host
		http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
		return
	}
	routeHandler(w, r)
}

func ServeHTTP(addr string) *http.Server {
	hs := &http.Server{Addr: addr, Handler: http.HandlerFunc(httpEntry)}
	go func() {
		log.Printf("HTTP listening on %s", addr)
		if err := hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()
	return hs
}

func ServeHTTPS(addr string, cm *CertManager) *http.Server {
	ts := &http.Server{Addr: addr, Handler: http.HandlerFunc(routeHandler)}
	ts.TLSConfig = &tls.Config{GetCertificate: cm.GetCertificate, MinVersion: tls.VersionTLS12}
	// Create listener *first* so we can expose the effective addr (when :0 was requested).
	ln, err := tls.Listen("tcp", addr, ts.TLSConfig)
	if err != nil {
		log.Fatalf("https listen: %v", err)
	}
	// Publish the effective addr (e.g., 127.0.0.1:51327) for tests and callers.
	ts.Addr = ln.Addr().String()
	log.Printf("HTTPS listening on %s", ts.Addr)
	go func() {
		if err := ts.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("https server: %v", err)
		}
	}()
	return ts
}
