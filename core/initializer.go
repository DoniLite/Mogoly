package core

import "sync"

func NewServerPool() *ServerPool {
	return &ServerPool{
		servers: make(map[string]*Server, 0),
		mu:      sync.Mutex{},
	}
}

func NewRoundRobinBalancer(pool *ServerPool) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		pool: pool,
		idx:  -1,
		mu:   sync.Mutex{},
	}
}

func NewLoadBalancer(strategy BalancerStrategy) *LoadBalancer {
	return &LoadBalancer{
		strategy: strategy,
	}
}
