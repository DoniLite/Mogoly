// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package server

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/DoniLite/Mogoly/core/events"
)

var (
	requestsPerMinute = 5
	rateLimitWindow   = time.Minute

	// map[ip] = []timestamps
	visitors = make(map[string][]time.Time)
	mu       sync.Mutex
)

// ping returns a "pong" message consider registering this Handler for the health checking logic
func Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("pong"))
	if err != nil {
		events.Logf(events.LOG_ERROR, "[PING_HANDLER]: Error writing response: %v", err)
	}
}

func HealthChecker(server *Server) (bool, error) {
	events.Logf(events.LOG_INFO, "[HEALTH_CHECKER]: Initializing health checking for the %s server", server.Name)
	raw, err := BuildServerURL(server)
	if err != nil {
		events.Logf(events.LOG_ERROR, "[HEALTH_CHECKER]: Error during the %s server url building \nerror: %s", server.Name, err.Error())
		return false, err
	}
	req, err := http.NewRequest(http.MethodGet, raw, nil)
	if err != nil {
		events.Logf(events.LOG_ERROR, "[HEALTH_CHECKER]: Error during the http request init for the %s server \nerror: %s", server.Name, err.Error())
		return false, err
	}
	client := &http.Client{Timeout: 3 * time.Second}
	events.Logf(events.LOG_DEBUG, "[HEALTH_CHECKER]: New http request to %v", client)
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			events.Logf(events.LOG_ERROR, "[HEALTH_CHECKER]: Error while closing the body reader: %v", err)
		}
	}()
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
