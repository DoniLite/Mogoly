package daemon

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/DoniLite/Mogoly/cli/actions"
	mogoly_sync "github.com/DoniLite/Mogoly/sync"
)

// Client represents a daemon client that communicates via socket
type Client struct {
	socketPath string
	timeout    time.Duration
	syncClient *mogoly_sync.Client
}

// NewClient creates a new daemon client
func NewClient(socketPath string) *Client {
	if socketPath == "" {
		socketPath = GetSocketPath()
	}

	return &Client{
		socketPath: socketPath,
		timeout:    10 * time.Second,
		syncClient: mogoly_sync.NewClient(),
	}
}

// Send a request to the daemon
func (c *Client) SendAction(ctx context.Context, actionType mogoly_sync.Action_Type, payload any) (*mogoly_sync.Message, error) {
	url := fmt.Sprintf("ws+unix://%s:/", c.socketPath)
	err := c.syncClient.Connect(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %v", err)
	}

	return c.syncClient.SendRequest(ctx, actionType, payload, nil)
}

// Ping the server to check if it is alive
func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.SendAction(ctx, actions.ActionDaemonPing, nil)
	if err != nil {
		return err
	}

	if resp.Error != "" {
		return fmt.Errorf("ping failed: %s", resp.Error)
	}

	return nil
}

// Sends a request with retry logic
func (c *Client) SendRequestWithRetry(ctx context.Context, actionType mogoly_sync.Action_Type, payload any, maxRetries int) (*mogoly_sync.Message, error) {
	var lastErr error

	for i := range maxRetries {
		resp, err := c.SendAction(ctx, actionType, payload)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		// Wait before retry (exponential backoff)
		if i < maxRetries-1 {
			waitTime := time.Duration(i+1) * 100 * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(waitTime):
				continue
			}
		}
	}

	return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}

// Get logs from the daemon
func (c *Client) GetLogs(ctx context.Context, payload *actions.DaemonLogsPayload) (string, error) {
	resp, err := c.SendAction(ctx, actions.ActionDaemonLogs, payload)
	if err != nil {
		return "", err
	}

	if resp.Error != "" {
		return "", fmt.Errorf("error: %s", resp.Error)
	}

	// Decode logs
	var result map[string]string
	if err := DecodePayload(resp, &result); err != nil {
		return "", fmt.Errorf("failed to decode logs response: %v", err)
	}

	return result["logs"], nil
}

func (c *Client) StreamLogs(ctx context.Context, payload *actions.DaemonLogsPayload, output io.Writer) error {
	logs, err := c.GetLogs(ctx, payload)
	if err != nil {
		return err
	}
	output.Write([]byte(logs))
	return nil
}

// SetTimeout sets the client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// Close closes the client connection
func (c *Client) Close() {
	c.syncClient.Close()
}
