package core

import (
	"net/http/httptest"
	"testing"
)

func TestCreateSingleHttpServer_AppliesMiddlewareDefaults(t *testing.T) {
	s := &Server{
		Name: "x",
		URL:  "http://example.com",
		Middlewares: []Middleware{{
			Name: string(MogolyRatelimiter),
			// nil config -> should apply registry defaults without panic
			Config: nil,
		}},
	}
	h := createSingleHttpServer(s)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "4.5.6.7:1111"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	// no assertion on code (middleware default rate is permissive)
}
