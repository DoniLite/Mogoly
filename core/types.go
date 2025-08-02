package core

import "sync"

type Server struct {
}

type ServerPool struct {
	Servers map[string]*Server
	Mu      sync.Mutex
}
