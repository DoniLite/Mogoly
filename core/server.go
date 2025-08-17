// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
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
	w.Write([]byte("pong"))
}

func HealthChecker(server *Server) (bool, error) {
	serverURL, err := url.Parse(server.URL)
	var responseBody []byte

	if server.URL == "" || err != nil {
		serverURL, err = url.Parse(fmt.Sprintf("%s://%s:%d", server.Protocol, server.Host, server.Port))
		if err != nil {
			return false, err
		}
	}

	req, err := http.NewRequest("GET", serverURL.String(), &io.LimitedReader{})

	if err != nil {
		return false, err
	}

	client := &http.Client{}

	res, err := client.Do(req)

	if err != nil {
		return false, err
	}

	_, err = res.Body.Read(responseBody)

	if err != nil {
		return false, err
	}

	if res.StatusCode >= 400 {
		return false, fmt.Errorf("server respond with error code: %s body: %s", fmt.Sprint(res.StatusCode), string(responseBody))
	}

	return true, nil
}

type RateLimitMiddlewareConfig struct {
	ReqPerMinute int
	LimitWindow time.Duration
}

func RateLimiterMiddleware(config *RateLimitMiddlewareConfig) func(next http.Handler) http.Handler {
	if config.ReqPerMinute != 0 {
		requestsPerMinute = config.ReqPerMinute
	}

	if config.LimitWindow != 0 {
		rateLimitWindow = config.LimitWindow
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr // or r.Header["X-Forwarded-For"][0] when the request comes from proxy

			mu.Lock()
			defer mu.Unlock()

			now := time.Now()
			requestTimes := visitors[ip]

			var filtered []time.Time
			for _, t := range requestTimes {
				if now.Sub(t) < rateLimitWindow {
					filtered = append(filtered, t)
				}
			}

			if len(filtered) >= requestsPerMinute {
				http.Error(w, "Trop de requêtes, réessaie plus tard", http.StatusTooManyRequests)
				return
			}

			filtered = append(filtered, now)
			visitors[ip] = filtered

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
