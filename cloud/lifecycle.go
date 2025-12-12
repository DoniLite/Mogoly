package cloud

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/api/types/container"
)

// StopInstance stops a running instance without deleting it
func (m *CloudManager) StopInstance(id string) error {
	m.mu.RLock()
	instance, exists := m.instances[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("instance not found: %s", id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := m.dockerClient.ContainerStop(ctx, instance.ContainerID, container.StopOptions{}); err != nil {
		return fmt.Errorf("error stopping container: %v", err)
	}

	m.mu.Lock()
	instance.Status = "stopped"
	m.mu.Unlock()

	return nil
}

// StartInstance starts a stopped instance
func (m *CloudManager) StartInstance(id string) error {
	m.mu.RLock()
	instance, exists := m.instances[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("instance not found: %s", id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := m.dockerClient.ContainerStart(ctx, instance.ContainerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("error starting container: %v", err)
	}

	m.mu.Lock()
	instance.Status = "running"
	m.mu.Unlock()

	return nil
}

// RestartInstance restarts a running or stopped instance
func (m *CloudManager) RestartInstance(id string) error {
	m.mu.RLock()
	instance, exists := m.instances[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("instance not found: %s", id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	timeout := 10
	if err := m.dockerClient.ContainerRestart(ctx, instance.ContainerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("error restarting container: %v", err)
	}

	m.mu.Lock()
	instance.Status = "running"
	m.mu.Unlock()

	return nil
}

// RefreshInstanceStatus syncs instance status from Docker
func (m *CloudManager) RefreshInstanceStatus(id string) error {
	m.mu.RLock()
	instance, exists := m.instances[id]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("instance not found: %s", id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inspection, err := m.dockerClient.ContainerInspect(ctx, instance.ContainerID)
	if err != nil {
		return fmt.Errorf("error inspecting container: %v", err)
	}

	m.mu.Lock()
	instance.Status = inspection.State.Status
	m.mu.Unlock()

	return nil
}

// GetInstanceLogs fetches the last N lines of logs from a container
func (m *CloudManager) GetInstanceLogs(id string, tailLines int) (string, error) {
	m.mu.RLock()
	instance, exists := m.instances[id]
	m.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("instance not found: %s", id)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tail := fmt.Sprintf("%d", tailLines)
	reader, err := m.dockerClient.ContainerLogs(ctx, instance.ContainerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	})
	if err != nil {
		return "", fmt.Errorf("error fetching logs: %v", err)
	}
	defer reader.Close()

	// Read all logs
	logs, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("error reading logs: %v", err)
	}

	return string(logs), nil
}
