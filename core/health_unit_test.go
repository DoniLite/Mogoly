package core

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoniLite/Mogoly/core/server"
)

func TestHealthChecker_OK_and_Fail(t *testing.T) {
	okSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer okSrv.Close()

	failSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusServiceUnavailable)
	}))
	defer failSrv.Close()

	sOK := &server.Server{URL: okSrv.URL}
	sBad := &server.Server{URL: failSrv.URL}

	ok, err := server.HealthChecker(sOK)
	if !ok || err != nil {
		t.Fatalf("expected ok health, got ok=%v err=%v", ok, err)
	}
	ok, err = server.HealthChecker(sBad)
	if ok || err == nil {
		t.Fatalf("expected fail health, got ok=%v err=%v", ok, err)
	}
}
