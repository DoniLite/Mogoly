package actions

import (
	"context"

	"github.com/DoniLite/Mogoly/cloud"
	"github.com/DoniLite/Mogoly/core/server"
	"github.com/DoniLite/Mogoly/sync"
)

type ActionListFunc func(ctx context.Context, reqID string, payload any) *sync.Message

type ActionHandlerSet map[sync.Action_Type]ActionListFunc

// DaemonLogsPayload for daemon.logs action
type DaemonLogsPayload struct {
	TailLines int  `json:"tail_lines"`
	Follow    bool `json:"follow"`
}

type ServiceConfigPayload = cloud.ServiceConfig
type ServerCreatePayload = server.Server
type ServerAddBackendPayload struct {
	Name   string         `json:"name"`
	Server *server.Server `json:"server"`
}
type ServerRemoveBackendPayload struct {
	BaseServerName string `json:"base_server_name"`
	BackendName    string `json:"backend_name"`
}

type CheckServerHealthPayload struct {
	Name       string `json:"name"`
	SelfOnly   bool   `json:"self_only"`
}

// Cloud payload types
type CloudCreatePayload struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Version      string `json:"version"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	DatabaseName string `json:"database_name"`
}

type CloudLogsPayload struct {
	ID        string `json:"id"`
	TailLines int    `json:"tail_lines"`
	Follow    bool   `json:"follow"`
}

// Load balancer payload types
type LBCreatePayload struct {
	Name       string `json:"name"`
	ConfigPath string `json:"config_path"`
}

type LBAddBackendPayload struct {
	LBName      string `json:"lb_name"`
	BackendName string `json:"backend_name"`
	BackendURL  string `json:"backend_url"`
}