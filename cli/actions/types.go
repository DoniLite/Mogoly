package actions

import (
	"context"

	"github.com/DoniLite/Mogoly/sync"
)

type ActionListFunc func(ctx context.Context, reqID string, payload any) *sync.Message

type ActionHandlerSet map[sync.Action_Type]ActionListFunc

// DaemonLogsPayload for daemon.logs action
type DaemonLogsPayload struct {
	TailLines int  `json:"tail_lines"`
	Follow    bool `json:"follow"`
}