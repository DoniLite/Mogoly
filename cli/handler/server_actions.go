package handler

import (
	"context"
	"fmt"

	"github.com/DoniLite/Mogoly/cli/actions"
	"github.com/DoniLite/Mogoly/cli/daemon"
	"github.com/DoniLite/Mogoly/core/server"
	"github.com/DoniLite/Mogoly/sync"
)

func BootStrapServer(ctx context.Context, reqID string, payload any) *sync.Message {
	serverConfig := payload.(*actions.ServerCreatePayload)

	router := daemon.GetServerRouter()

	router.AddServer(serverConfig)
	svr, err := router.GetServer(serverConfig.Name)
	if err != nil {
		return sync.NewErrorMessage(err.Error(), fmt.Sprintf("request id: %s", reqID))
	}

	msg, err := sync.NewMessage(actions.ActionLBCreate, svr, map[string]any{
		"URL": svr.URL,
	})
	if err != nil {
		return sync.NewErrorMessage(err.Error(), fmt.Sprintf("request id: %s", reqID))
	}

	return msg
}

func ListServers(ctx context.Context, reqID string, payload any) *sync.Message {
	router := daemon.GetServerRouter()

	servers := router.ListServers()

	msg, err := sync.NewMessage(actions.ActionLBList, servers, map[string]any{
		"count": len(servers),
	})
	if err != nil {
		return sync.NewErrorMessage(err.Error(), fmt.Sprintf("request id: %s", reqID))
	}

	return msg
}

func AddBackend(ctx context.Context, reqID string, payload any) *sync.Message {
	parsedPayload := payload.(*actions.ServerAddBackendPayload)

	router := daemon.GetServerRouter()

	svr, err := router.GetServer(parsedPayload.Name)
	if err != nil {
		return sync.NewErrorMessage(err.Error(), fmt.Sprintf("request id: %s", reqID))
	}

	svr.AddNewBalancingServer(parsedPayload.Server)

	msg, err := sync.NewMessage(actions.ActionLBAddBackend, svr, map[string]any{
		"URL": svr.URL,
	})
	if err != nil {
		return sync.NewErrorMessage(err.Error(), fmt.Sprintf("request id: %s", reqID))
	}

	return msg
}

func RemoveBackend(ctx context.Context, reqID string, payload any) *sync.Message {
	parsedPayload := payload.(*actions.ServerRemoveBackendPayload)

	router := daemon.GetServerRouter()

	svr, err := router.GetServer(parsedPayload.BaseServerName)
	if err != nil {
		return sync.NewErrorMessage(err.Error(), fmt.Sprintf("request id: %s", reqID))
	}

	svr.DelBalancingServer(parsedPayload.BackendName)

	msg, err := sync.NewMessage(actions.ActionLBRemoveBackend, svr, map[string]any{
		"URL": svr.URL,
	})
	if err != nil {
		return sync.NewErrorMessage(err.Error(), fmt.Sprintf("request id: %s", reqID))
	}

	return msg
}

func CheckServerHealth(ctx context.Context, reqID string, payload any) *sync.Message {

	var selfStatus *server.ServerStatus
	var allServersStatus *server.HealthCheckStatus

	parsedPayload := payload.(*actions.CheckServerHealthPayload)

	router := daemon.GetServerRouter()

	svr, err := router.GetServer(parsedPayload.Name)
	if err != nil {
		return sync.NewErrorMessage(err.Error(), fmt.Sprintf("request id: %s", reqID))
	}

	selfStatus, err = svr.CheckHealthSelf()
	if err != nil {
		return sync.NewErrorMessage(err.Error(), fmt.Sprintf("request id: %s", reqID))
	}

	if !parsedPayload.SelfOnly {
		allServersStatus, err = svr.CheckHealthAll()
		if err != nil {
			return sync.NewErrorMessage(err.Error(), fmt.Sprintf("request id: %s", reqID))
		}
	}

	msg, err := sync.NewMessage(actions.ActionLBHealth, struct {
		SelfStatus       *server.ServerStatus      `json:"self_status"`
		AllServersStatus *server.HealthCheckStatus `json:"all_servers_status"`
	}{
		SelfStatus:       selfStatus,
		AllServersStatus: allServersStatus,
	}, map[string]any{
		"URL": svr.URL,
	})
	if err != nil {
		return sync.NewErrorMessage(err.Error(), fmt.Sprintf("request id: %s", reqID))
	}

	return msg
}
