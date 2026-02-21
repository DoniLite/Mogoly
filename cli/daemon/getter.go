package daemon

import "github.com/DoniLite/Mogoly/core/router"

func GetServerRouter() *router.RouterState {
	return server.mogolyRouter
}