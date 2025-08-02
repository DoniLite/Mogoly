package core

import "sync"

func NewServerPool() *ServerPool {
	return &ServerPool{
		Servers: make(map[string]*Server, 0),
		Mu: sync.Mutex{},
	}
}
