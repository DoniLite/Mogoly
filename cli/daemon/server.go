package daemon

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/DoniLite/Mogoly/cli/actions"
	mogoly_sync "github.com/DoniLite/Mogoly/sync"
)

// Server represents the daemon server
type Server struct {
	socketPath   string
	listener     net.Listener
	mu           sync.RWMutex
	running      bool
	logWriter    io.Writer
	shutdownChan chan struct{}
	syncServer   *mogoly_sync.Server
	httpServer   *http.Server
}

// NewServer creates a new daemon server
func NewServer(socketPath string) (*Server, error) {
	if socketPath == "" {
		socketPath = GetSocketPath()
	}

	logPath, err := GetLogFilePath()
	if err != nil {
		return nil, err
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	s := &Server{
		socketPath:   socketPath,
		logWriter:    io.MultiWriter(os.Stdout, logFile),
		shutdownChan: make(chan struct{}),
	}

	// Initialize sync server
	s.syncServer = mogoly_sync.NewServer(s.handleMessage, nil)

	return s, nil
}

// Start starts the daemon server
func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server is already running")
	}
	s.running = true
	s.mu.Unlock()

	// Remove existing socket file
	if err := os.RemoveAll(s.socketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket: %v", err)
	}

	// Create listener
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to create listener: %v", err)
	}
	s.listener = listener

	// Set socket permissions
	if err := os.Chmod(s.socketPath, 0666); err != nil {
		return fmt.Errorf("failed to set socket permissions: %v", err)
	}

	s.log("Daemon started on socket: %s", s.socketPath)

	// Write PID file
	if err := s.writePIDFile(); err != nil {
		s.log("Warning: failed to write PID file: %v", err)
	}

	// Start sync server hub
	s.syncServer.Run()

	// Start HTTP server
	s.httpServer = &http.Server{
		Handler: s.syncServer,
	}

	go func() {
		if err := s.httpServer.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.log("HTTP server error: %v", err)
		}
	}()

	// Handle shutdown signals
	go s.handleSignals()

	return nil
}

// handleMessage processes a message from the client
func (s *Server) handleMessage(msg *mogoly_sync.Message, conn *mogoly_sync.Connection) error {
	s.log("Received request: %d (ReqID: %s)", msg.Action.Type, msg.RequestID)

	ctx := context.Background()
	var response *mogoly_sync.Message

	reqID := msg.RequestID

	if handler, ok := actions.GetHandler(msg.Action.Type); ok {
		response = handler(ctx, reqID, msg.Action.Payload)
	} else {
		response = NewErrorMessage(reqID, msg.Action.Type, fmt.Sprintf("unknown action: %d", msg.Action.Type))
	}

	if response != nil {
		conn.SendMsg(response)
	}

	return nil
}

// Stop stops the daemon server
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("server is not running")
	}

	s.log("Stopping daemon...")

	s.running = false
	close(s.shutdownChan)

	// Stop HTTP server
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}

	// Close listener if not already closed by Shutdown
	if s.listener != nil {
		s.listener.Close()
	}

	// Remove socket file
	os.RemoveAll(s.socketPath)

	// Remove PID file
	pidPath, _ := GetPIDFilePath()
	os.RemoveAll(pidPath)

	s.log("Daemon stopped")
	return nil
}

// handleSignals handles OS signals for graceful shutdown
func (s *Server) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		s.log("Received signal: %v", sig)
		s.Stop()
	case <-s.shutdownChan:
		return
	}
}

// writePIDFile writes the current process ID to a file
func (s *Server) writePIDFile() error {
	pidPath, err := GetPIDFilePath()
	if err != nil {
		return err
	}

	pid := os.Getpid()
	return os.WriteFile(pidPath, fmt.Appendf(nil, "%d", pid), 0644)
}

// log writes a log message
func (s *Server) log(format string, args ...any) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("[%s] %s\n", timestamp, fmt.Sprintf(format, args...))
	s.logWriter.Write([]byte(msg))
}

// Wait waits for the server to shutdown
func (s *Server) Wait() {
	<-s.shutdownChan
}
