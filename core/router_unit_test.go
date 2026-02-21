package core

import (
	"net/http/httptest"
	"testing"

	"github.com/DoniLite/Mogoly/core/router"
	"github.com/DoniLite/Mogoly/core/server"
)

func TestCreateSingleHttpServer_AppliesMiddlewareDefaults(t *testing.T) {
	s := &server.Server{
		Name: "x",
		URL:  "http://example.com",
		Middlewares: []server.Middleware{{
			Name: string(server.MogolyRatelimiter),
			// nil config -> should apply registry defaults without panic
			Config: nil,
		}},
	}
	h := router.CreateSingleHttpServer(s)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "4.5.6.7:1111"
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	// no assertion on code (middleware default rate is permissive)
}
