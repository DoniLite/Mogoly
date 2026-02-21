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