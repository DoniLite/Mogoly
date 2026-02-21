package handler

import "github.com/DoniLite/Mogoly/cli/actions"

func init() {
	actions.RegisterHandler(actions.ActionServerCreate, BootStrapServer)
	actions.RegisterHandler(actions.ActionServerList, ListServers)
	actions.RegisterHandler(actions.ActionServerAddBackend, AddBackend)
	actions.RegisterHandler(actions.ActionServerRemoveBackend, RemoveBackend)
	actions.RegisterHandler(actions.ActionServerHealth, CheckServerHealth)
}