// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package sync

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/DoniLite/Mogoly/core"
	"github.com/google/uuid"
)

// Server

func (s *Server) handleMessage(msg *Message, client *connection) error {
	switch msg.Action.Type {
	case CREATE_SERVER:
		{
			var globalConf core.Config
			err := msg.DecodePayload(&globalConf)
			if err != nil {
				errMsg := NewErrorMessage("Invalid config struct provided", err.Error())
				if msg.RequestID != "" {
					errMsg.RequestID = msg.RequestID
				}
				client.sendMsg(errMsg)
				return nil
			}

			for _, singleServer := range globalConf.Servers {
				s.globalConf.Servers = append(s.globalConf.Servers, singleServer)
				s.HostConfig[singleServer.Name] = createSingleHttpServer(singleServer)
			}

			if globalConf.HealthCheckInterval != 0 {
				s.globalConf.HealthCheckInterval = globalConf.HealthCheckInterval
			}
			if globalConf.LogOutput != "" {
				s.globalConf.LogOutput = globalConf.LogOutput
			}
			s.globalConf.Middlewares = globalConf.Middlewares
			if msg.RequestID != "" {
				newMsg, err := NewMessage(CREATE_SERVER, struct {
					config  *core.Config
					success bool
				}{config: s.globalConf, success: true}, nil)

				if err != nil {
					client.send <- &Message{RequestID: msg.RequestID}
					return nil
				}

				newMsg.RequestID = msg.RequestID

				client.sendMsg(newMsg)
			}
			return nil
		}
	case ROLLBACK_SERVER:
		{
			var server core.Server
			err := msg.DecodePayload(&server)
			if err != nil {
				errMsg := NewErrorMessage("No server provided for roll backing", err.Error())
				if msg.RequestID != "" {
					errMsg.RequestID = msg.RequestID
				}
				client.sendMsg(errMsg)
				return nil
			}

			for idx, confServer := range s.globalConf.Servers {
				if confServer.Name == server.Name {
					if server.BalancingServers != nil {
						s.globalConf.Servers[idx].RollBack(server.BalancingServers)
					}
					s.globalConf.Servers[idx] = &server
					s.HostConfig[confServer.Name] = createSingleHttpServer(&server)
				}
			}

			if msg.RequestID != "" {
				newMsg, err := NewMessage(CREATE_SERVER, struct {
					config  *core.Config
					success bool
				}{config: s.globalConf, success: true}, nil)

				if err != nil {
					client.send <- &Message{RequestID: msg.RequestID}
					return nil
				}

				newMsg.RequestID = msg.RequestID

				client.sendMsg(newMsg)
			}

			return nil
		}
	}
	return nil
}

// Handling http request and trying to upgrade it to a websocket connection.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ServeHTTP: Failed to upgrade connection: %v\n", err)
		return
	}
	log.Printf("ServeHTTP: Client connected from %s\n", ws.RemoteAddr())

	conn := newConnection(ws)

	s.hub.register <- conn

	go conn.writePump()
	go conn.readPump(s.hub.handleIncomingMessage, s.hub.handleDisconnect)
}

// Launching the Hub in a goroutine.
func (s *Server) Run() {
	go s.hub.run()
}

// Client

// Connect to the given server url websocket with the provided headers.
func (c *Client) Connect(serverUrl string, headers http.Header) error {
	c.mu.Lock()
	if c.isConnected {
		c.mu.Unlock()
		return fmt.Errorf("client already connected")
	}
	c.connUrl = serverUrl
	c.headers = headers
	c.mu.Unlock()

	log.Printf("Client: Attempting to connect to %s...\n", serverUrl)
	ws, resp, err := c.dialer.Dial(c.connUrl, c.headers)
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
	log.Printf("Client: Successfully connected to %s\n", c.connUrl)

	c.mu.Lock()
	c.conn = newConnection(ws)
	c.isConnected = true
	c.mu.Unlock()

	go c.conn.writePump()
	go c.conn.readPump(c.handleIncomingMessage, c.handleDisconnect)

	return nil
}

func (c *Client) handleIncomingMessage(msg *Message, conn *connection) error {
	log.Printf("Client: Received message type %d (ReqID: %s)\n", msg.Action.Type, msg.RequestID) // Debug

	// Check if it's a pending request
	c.pendingMu.Lock()
	if msg.RequestID != "" {
		if respChan, ok := c.pendingRequests[msg.RequestID]; ok {
			log.Printf("Client: Correlated response for RequestID %s\n", msg.RequestID)
			select {
			case respChan <- msg:
			default:
				log.Printf("Warning: No listener for response channel of RequestID %s\n", msg.RequestID)
			}
			delete(c.pendingRequests, msg.RequestID)
			c.pendingMu.Unlock()
			return nil
		}
	}
	c.pendingMu.Unlock()

	select {
	case c.Incoming <- msg:
	default:
		log.Printf("Warning: Client Incoming channel full. Message type %d dropped.\n", msg.Action.Type)
	}
	return nil
}

func (c *Client) handleDisconnect(conn *connection) {
	c.mu.Lock()
	if c.conn != conn {
		c.mu.Unlock()
		log.Printf("Client: Received disconnect signal for an old/stale connection (%p)\n", conn.ws)
		return
	}
	c.isConnected = false
	c.conn = nil
	log.Println("Client: Connection lost.")
	c.mu.Unlock()

	// Clean the pending request for this connection
	c.pendingMu.Lock()
	if len(c.pendingRequests) > 0 {
		log.Printf("Client: Cleaning up %d pending requests due to disconnect.\n", len(c.pendingRequests))
		for reqID, respChan := range c.pendingRequests {
			close(respChan)
			delete(c.pendingRequests, reqID)
		}
	}
	c.pendingMu.Unlock()
}

// sending message to the server asynchronously.
func (c *Client) Send(msg *Message) error {
	c.mu.Lock()
	conn := c.conn
	isConnected := c.isConnected
	c.mu.Unlock()

	if !isConnected || conn == nil {
		return fmt.Errorf("client not connected")
	}
	log.Printf("Client: Sending message type %d async\n", msg.Action.Type) // Debug
	conn.sendMsg(msg)
	return nil
}

// sending a request and waiting for the response based on the RequestID.
func (c *Client) SendRequest(ctx context.Context, msgType Action_Type, payload any, meta any) (*Message, error) {
	c.mu.Lock()
	conn := c.conn
	isConnected := c.isConnected
	c.mu.Unlock()

	if !isConnected || conn == nil {
		return nil, fmt.Errorf("client not connected")
	}

	requestID := uuid.NewString()
	msg, err := NewMessage(msgType, payload, meta)

	if err != nil {
		return nil, err
	}

	msg.RequestID = requestID

	respChan := make(chan *Message, 1)

	c.pendingMu.Lock()
	c.pendingRequests[requestID] = respChan
	c.pendingMu.Unlock()

	// Cleaning the request before the response (success, error, timeout)
	defer func() {
		c.pendingMu.Lock()
		delete(c.pendingRequests, requestID)
		c.pendingMu.Unlock()
	}()

	// Send the request
	log.Printf("Client: Sending request %s (Type: %d)\n", requestID, msg.Action.Type)
	conn.sendMsg(msg)

	// Waiting for the response
	select {
	case resp := <-respChan:
		log.Printf("Client: Received response for request %s (Type: %d, Error: '%s')\n", requestID, resp.Action.Type, resp.Error)
		if resp.Error != "" || resp.Action.Type == ERROR {
			errMsg := resp.Error
			if errMsg == "" {
				errMsg = "received error event"
			}
			var errPayload ErrorPayload
			if resp.DecodePayload(&errPayload) == nil && errPayload.Details != "" {
				errMsg = fmt.Sprintf("%s: %s", errMsg, errPayload.Details)
			}
			return nil, fmt.Errorf("server error response for request %s: %s", requestID, errMsg)
		}
		return resp, nil

	case <-ctx.Done():
		log.Printf("Client: Context done while waiting for response to request %s: %v\n", requestID, ctx.Err())
		return nil, fmt.Errorf("request %s timed out or was canceled: %w", requestID, ctx.Err())
	}
}

// Close the websocket connection and stopping the client.
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Println("Client: Close called.")

	if c.conn != nil && c.isConnected {
		c.conn.closeSend()
	}
	c.isConnected = false
}
