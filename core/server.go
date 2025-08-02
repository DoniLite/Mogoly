package core

import "net/http"

// ping returns a "pong" message consider registering this Handler for the health checking logic
func Ping(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("pong"))
}