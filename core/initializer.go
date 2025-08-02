package core

import (
	"net/http/httputil"
	"net/url"
	"sync"
)

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

func NewLoadBalancer(strategy BalancerStrategy, proxy *httputil.ReverseProxy) *LoadBalancer {
	return &LoadBalancer{
		strategy: strategy,
		proxy:    proxy,
	}
}

func NewProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	return proxy
}
