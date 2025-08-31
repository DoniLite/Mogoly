package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
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

	sOK := &Server{URL: okSrv.URL}
	sBad := &Server{URL: failSrv.URL}

	ok, err := HealthChecker(sOK)
	if !ok || err != nil {
		t.Fatalf("expected ok health, got ok=%v err=%v", ok, err)
	}
	ok, err = HealthChecker(sBad)
	if ok || err == nil {
		t.Fatalf("expected fail health, got ok=%v err=%v", ok, err)
	}
}
