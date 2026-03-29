package handler

import "github.com/DoniLite/Mogoly/cli/actions"

func init() {
	actions.RegisterHandler(actions.ActionLBCreate, BootStrapServer)
	actions.RegisterHandler(actions.ActionLBList, ListServers)
	actions.RegisterHandler(actions.ActionLBAddBackend, AddBackend)
	actions.RegisterHandler(actions.ActionLBRemoveBackend, RemoveBackend)
	actions.RegisterHandler(actions.ActionLBHealth, CheckServerHealth)
}