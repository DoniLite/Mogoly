package core

import "sync"

func NewServerPool() *ServerPool {
	return &ServerPool{
		Servers: make(map[string]*Server, 0),
		Mu: sync.Mutex{},
	}
}

func NewRoundRobinBalancer(pool *ServerPool) *RoundRobinBalancer {
	return  &RoundRobinBalancer{
		Pool: pool,
		Idx: -1,
		Mu: sync.Mutex{},
	}
}
