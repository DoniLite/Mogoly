package core

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
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
