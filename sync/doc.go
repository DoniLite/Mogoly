// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
# Mogoly Sync

This package provides real-time communication (RTC) capabilities for the Mogoly software,
enabling WebSocket-based communication between the CLI and GUI components.

# Features

  - WebSocket Server: Accept and manage multiple client connections
  - WebSocket Client: Connect to servers with automatic reconnection
  - Message Protocol: JSON-based request/response messaging system
  - Connection Hub: Centralized connection management and broadcasting
  - Request Correlation: Track requests and responses via unique IDs
  - Ping/Pong: Automatic connection keep-alive mechanism
  - Cross-Platform: Unix socket and Windows named pipe support

# Architecture

The sync package implements a hub-and-spoke WebSocket architecture:

	┌──────────┐
	│   GUI    │ (Client)
	└────┬─────┘
	     │ WebSocket
	     ▼
	┌─────────────────┐
	│   Hub Server    │
	└────┬────────────┘
	     │
	     ▼
	┌─────────────────┐
	│   CLI Client    │
	└─────────────────┘

The Hub manages:
  - Client registration/unregistration
  - Message routing between connections
  - Broadcasting to all clients
  - Connection lifecycle management

# Quick Start

## Server Setup

Create a WebSocket server to accept GUI connections:

	package main

	import (
	    "log"
	    "net/http"
	    "github.com/DoniLite/Mogoly/sync"
	)

	func main() {
	    // Define message handler
	    handler := func(msg *sync.Message, conn *sync.Connection) error {
	        log.Printf("Received: %+v", msg)

	        switch msg.Action.Type {
	        case sync.CREATE_SERVER:
	            // Handle server creation
	            var payload map[string]interface{}
	            if err := msg.DecodePayload(&payload); err != nil {
	                return err
	            }
	            log.Printf("Creating server: %+v", payload)

	            // Send response
	            response := &sync.Message{
	                RequestID: msg.RequestID,
	                Action: sync.Action{Type: sync.CREATE_SERVER},
	            }
	            response.Action.AddPayload(map[string]string{
	                "status": "created",
	            })
	            conn.SendMsg(response)

	        case sync.PING:
	            // Respond to ping
	            pong := &sync.Message{
	                RequestID: msg.RequestID,
	                Action: sync.Action{Type: sync.PING},
	            }
	            conn.SendMsg(pong)
	        }

	        return nil
	    }

	    // Create server
	    server := sync.NewServer(handler)
	    server.Run()

	    // Start HTTP server with WebSocket endpoint
	    http.Handle("/ws", server)
	    log.Fatal(http.ListenAndServe(":9000", nil))
	}

## Client Setup

Connect to a WebSocket server from the CLI:

	package main

	import (
	    "context"
	    "log"
	    "time"
	    "github.com/DoniLite/Mogoly/sync"
	)

	func main() {
	    // Create client
	    client, err := sync.NewClient("ws://localhost:9000/ws", nil)
	    if err != nil {
	        log.Fatal(err)
	    }
	    defer client.Close()

	    // Handle incoming messages
	    go func() {
	        for msg := range client.Incoming {
	            log.Printf("Received: %+v", msg)
	        }
	    }()

	    // Send request with response
	    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	    defer cancel()

	    payload := map[string]string{
	        "name": "backend-1",
	        "url":  "http://localhost:8080",
	    }

	    response, err := client.SendRequest(ctx, sync.CREATE_SERVER, payload, nil)
	    if err != nil {
	        log.Fatal(err)
	    }

	    log.Printf("Response: %+v", response)
	}

# Message Protocol

## Message Structure

All messages follow this JSON structure:

	{
	  "request_id": "unique-uuid",
	  "action": {
	    "type": 0,
	    "payload": {...}
	  },
	  "meta": "...",
	  "error": ""
	}

## Action Types

	const (
	    CREATE_SERVER   Action_Type = iota  // Create new server
	    ROLLBACK_SERVER                     // Rollback server configuration
	    ADD_SERVER                          // Add server to pool
	    KILL_SERVER                         // Stop server
	    REBOOT_SERVER                       // Restart server
	    ERROR                               // Error message
	    PING                                // Keep-alive ping
	)

## Creating Messages

	// Using helper function
	msg, err := sync.NewMessage(sync.CREATE_SERVER, payload, metadata)
	if err != nil {
	    log.Fatal(err)
	}

	// Manual creation
	msg := &sync.Message{
	    RequestID: uuid.NewString(),
	    Action: sync.Action{
	        Type: sync.CREATE_SERVER,
	    },
	}
	msg.Action.AddPayload(payload)

## Decoding Payloads

	var payload map[string]interface{}
	if err := msg.DecodePayload(&payload); err != nil {
	    log.Fatal(err)
	}

	// Or decode to struct
	type ServerPayload struct {
	    Name string `json:"name"`
	    URL  string `json:"url"`
	}

	var serverPayload ServerPayload
	if err := msg.DecodePayload(&serverPayload); err != nil {
	    log.Fatal(err)
	}

# Server API

## Creating a Server

	handler := func(msg *sync.Message, conn *sync.Connection) error {
	    // Handle message
	    return nil
	}

	server := sync.NewServer(handler)

## Running the Server

	// Start the hub (required before accepting connections)
	server.Run()

	// Use as HTTP handler
	http.Handle("/ws", server)

## Message Handler

The message handler receives all incoming messages:

	func handleMessage(msg *sync.Message, conn *sync.Connection) error {
	    switch msg.Action.Type {
	    case sync.CREATE_SERVER:
	        // Decode payload
	        var payload CreateServerPayload
	        if err := msg.DecodePayload(&payload); err != nil {
	            return err
	        }

	        // Process request
	        result := createServer(payload)

	        // Send response
	        response := &sync.Message{
	            RequestID: msg.RequestID,
	            Action: sync.Action{Type: sync.CREATE_SERVER},
	        }
	        response.Action.AddPayload(result)
	        conn.SendMsg(response)

	    case sync.ERROR:
	        log.Printf("Error from client: %s", msg.Error)

	    default:
	        log.Printf("Unknown action type: %d", msg.Action.Type)
	    }

	    return nil
	}

# Client API

## Creating a Client

	// Basic connection
	client, err := sync.NewClient("ws://localhost:9000/ws", nil)

	// With custom headers
	headers := http.Header{}
	headers.Set("Authorization", "Bearer token123")
	client, err := sync.NewClient("ws://localhost:9000/ws", headers)

## Sending Messages

### Fire-and-Forget

	msg := &sync.Message{
	    Action: sync.Action{Type: sync.PING},
	}

	err := client.Send(msg)

### Request-Response Pattern

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	payload := map[string]string{"name": "server-1"}
	response, err := client.SendRequest(ctx, sync.CREATE_SERVER, payload, nil)
	if err != nil {
	    log.Fatal(err)
	}

	var result map[string]interface{}
	response.DecodePayload(&result)

## Receiving Messages

	go func() {
	    for msg := range client.Incoming {
	        switch msg.Action.Type {
	        case sync.PING:
	            log.Println("Received ping")

	        default:
	            log.Printf("Received: %+v", msg)
	        }
	    }
	}()

## Closing Connection

	defer client.Close()

# Connection Management

## Connection Parameters

	const (
	    writeWait      = 10 * time.Second  // Write deadline
	    pongWait       = 60 * time.Second  // Read deadline
	    pingPeriod     = 54 * time.Second  // Ping interval (< pongWait)
	    maxMessageSize = 8192              // Max message size (8KB)
	)

## Connection Lifecycle

	┌──────────────┐
	│  Connect     │
	└──────┬───────┘
	       │
	       ▼
	┌──────────────┐
	│  Register    │ (Hub adds to clients map)
	└──────┬───────┘
	       │
	       ▼
	┌──────────────┐
	│  Active      │ (Send/Receive messages)
	│  + Ping/Pong │
	└──────┬───────┘
	       │
	       ▼
	┌──────────────┐
	│  Disconnect  │ (Error, timeout, or close)
	└──────┬───────┘
	       │
	       ▼
	┌──────────────┐
	│  Unregister  │ (Hub removes from clients)
	└──────────────┘

## Hub Operations

The Hub manages all active connections:

	type Hub struct {
	    clients        map[*Connection]bool      // Active connections
	    register       chan *Connection          // Registration requests
	    unregister     chan *Connection          // Unregistration requests
	    broadcast      chan *Message             // Broadcast messages
	    messageHandler func(*Message, *Connection) error
	}

### Broadcasting

	// Hub automatically broadcasts to all clients
	hub.broadcast <- message

### Connection Count

	// Access via RWMutex
	hub.mu.RLock()
	count := len(hub.clients)
	hub.mu.RUnlock()

# Error Handling

## Error Messages

	// Create error message
	errMsg := sync.NewErrorMessage("Operation failed", "Details here")

	// Send error to client
	conn.SendMsg(errMsg)

## Error Payload

	type ErrorPayload struct {
	    Code    int    `json:"code,omitempty"`
	    Details string `json:"details"`
	}

## Handling Errors in Client

	response, err := client.SendRequest(ctx, sync.CREATE_SERVER, payload, nil)
	if err != nil {
	    log.Printf("Request failed: %v", err)
	    return
	}

	if response.Error != "" {
	    var errPayload sync.ErrorPayload
	    response.DecodePayload(&errPayload)
	    log.Printf("Server error: %s (code: %d)",
	        errPayload.Details, errPayload.Code)
	}

# Cross-Platform Support

## Unix Sockets (Linux/macOS)

	// See connect_unix.go
	client, err := sync.NewClientUnix("/tmp/mogoly.sock")

## Named Pipes (Windows)

	// See connect_windows.go
	client, err := sync.NewClientPipe(`\\.\pipe\mogoly`)

# Best Practices

1. **Context Timeouts**: Always use context with timeout for SendRequest()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

2. **Handle Disconnections**: Implement reconnection logic in production clients

3. **Message Size**: Keep messages under 8KB or adjust maxMessageSize constant

4. **Concurrent Safety**: The hub and connection types are thread-safe

5. **Clean Shutdown**: Always defer client.Close() to clean up resources

6. **Error Handling**: Check errors from SendRequest and Send methods

7. **Channel Buffering**: The Incoming channel is buffered, but process messages promptly

8. **Ping/Pong**: The package automatically handles ping/pong for keep-alive

9. **Request IDs**: Use unique request IDs for proper correlation

10. **Handler Performance**: Keep message handlers fast to avoid blocking

# Advanced Usage

## Custom Message Validation

	func validateMessage(msg *sync.Message) error {
	    if msg.RequestID == "" {
	        return fmt.Errorf("request ID required")
	    }

	    if msg.Action.Type < sync.CREATE_SERVER || msg.Action.Type > sync.PING {
	        return fmt.Errorf("invalid action type: %d", msg.Action.Type)
	    }

	    return nil
	}

## Connection Authentication

	func handleMessage(msg *sync.Message, conn *sync.Connection) error {
	    // Verify authentication on first message
	    if !conn.IsAuthenticated() {
	        var authPayload AuthPayload
	        if err := msg.DecodePayload(&authPayload); err != nil {
	            conn.SendMsg(sync.NewErrorMessage("Auth failed", "Invalid payload"))
	            return err
	        }

	        if !validateToken(authPayload.Token) {
	            conn.SendMsg(sync.NewErrorMessage("Auth failed", "Invalid token"))
	            return fmt.Errorf("authentication failed")
	        }

	        conn.MarkAuthenticated()
	    }

	    // Handle authenticated messages
	    // ...
	}

## Broadcasting Server Status

	// Broadcast to all connected clients
	status := ServerStatus{
	    Name:    "backend-1",
	    Healthy: true,
	}

	statusMsg := &sync.Message{
	    Action: sync.Action{Type: sync.STATUS_UPDATE},
	}
	statusMsg.Action.AddPayload(status)

	// Send via hub's broadcast channel
	hub.broadcast <- statusMsg

# Types Reference

## Message

	type Message struct {
	    RequestID string  // Unique request identifier (UUID)
	    Action    Action  // Action type and payload
	    Meta      []byte  // Additional metadata
	    Error     string  // Error message (if any)
	}

## Action

	type Action struct {
	    Type    Action_Type     // Action type enum
	    Payload json.RawMessage // JSON-encoded payload
	}

## Connection

	type Connection struct {
	    ws   *websocket.Conn  // Underlying WebSocket connection
	    send chan *Message    // Outgoing message channel (buffered: 256)
	}

## Client

	type Client struct {
	    conn            *Connection         // Connection wrapper
	    Incoming        chan *Message       // Incoming messages channel
	    isConnected     bool                // Connection state
	    pendingRequests map[string]chan *Message // Request correlation
	}

## Server

	type Server struct {
	    upgrader   websocket.Upgrader  // WebSocket upgrader
	    hub        *Hub                // Connection hub
	    msgHandler HandlerFunc         // Message handler function
	}

# Performance Considerations

- **Message Buffering**: Send channels are buffered (256 messages)
- **Concurrent Handlers**: Message handlers run in goroutines
- **Memory Usage**: Each connection allocates ~2KB for buffers
- **Ping Overhead**: Minimal (ping every 54 seconds)
- **JSON Marshaling**: Consider protobuf for high-throughput scenarios

# Troubleshooting

## Connection Refused

- Verify WebSocket server is running
- Check the URL and port
- Ensure firewall allows connections

## Messages Not Received

- Check the Incoming channel is being read
- Verify message handler doesn't block
- Check for network connectivity issues

## High Memory Usage

- Process Incoming messages promptly
- Monitor number of pending requests
- Check for goroutine leaks in handlers

## Connection Drops

- Review ping/pong timing
- Check network stability
- Implement reconnection logic
- Monitor server load

# See Also

  - Package core: Load balancing engine that uses sync for communication
  - Package cloud: Docker service management
*/
package sync
