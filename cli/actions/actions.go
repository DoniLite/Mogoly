package actions

import "github.com/DoniLite/Mogoly/sync"

var handlers ActionHandlerSet

func init() {
	handlers = ActionHandlerSet{}
}

func RegisterHandler(action sync.Action_Type, handler ActionListFunc) {
	handlers[action] = handler
}

func GetHandler(action sync.Action_Type) (ActionListFunc, bool) {
	if handler, ok := handlers[action]; ok {
		return handler, true
	}
	return nil, false
}
