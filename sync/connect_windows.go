//go:build windows

package sync

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
	"github.com/gorilla/websocket"
)

func (c *Client) Connect(serverUrl string, headers http.Header) error {
	c.mu.Lock()
	if c.isConnected {
		c.mu.Unlock()
		return fmt.Errorf("client already connected")
	}
	c.connUrl = serverUrl
	c.headers = headers
	c.mu.Unlock()

	logf(LOG_INFO, "Client: Attempting to connect to %s...\n", serverUrl)

	var ws *websocket.Conn
	var resp *http.Response
	var err error

	switch {
	case strings.HasPrefix(serverUrl, "np://"):
		parts := strings.SplitN(strings.TrimPrefix(serverUrl, "np://"), ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid named pipe url: %s", serverUrl)
		}
		pipeName, requestPath := parts[0], parts[1]
		if requestPath == "" {
			requestPath = "/"
		}
		pipePath := `\\.\pipe\` + pipeName

		dialer := websocket.Dialer{
			NetDial: func(network, addr string) (net.Conn, error) {
				return winio.DialPipe(pipePath, nil)
			},
			HandshakeTimeout: 5 * time.Second,
		}

		u := url.URL{Scheme: "ws", Host: "pipe", Path: requestPath}
		ws, resp, err = dialer.Dial(u.String(), headers)

	default:
		ws, resp, err = c.dialer.Dial(serverUrl, headers)
	}

	if err != nil {
		errMsg := fmt.Sprintf("Client: Failed to connect to %s: %v", c.connUrl, err)
		if resp != nil {
			errMsg = fmt.Sprintf("%s (Status: %s)", errMsg, resp.Status)
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if len(body) > 0 {
				errMsg = fmt.Sprintf("%s - Body: %s", errMsg, string(body))
			}
		}
		return fmt.Errorf("an error occurred %s", errMsg)
	}

	logf(LOG_INFO, "Client: Successfully connected to %s\n", c.connUrl)

	c.mu.Lock()
	c.conn = NewConnection(ws)
	c.isConnected = true
	c.mu.Unlock()

	go c.conn.writePump()
	go c.conn.readPump(c.handleIncomingMessage, c.handleDisconnect)

	return nil
}
