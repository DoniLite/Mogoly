package core

import (
	"fmt"

	"github.com/google/uuid"
)

func (sp *ServerPool) AddNewServer(server *Server) string {
	serverUUID := uuid.NewString()
	sp.Servers[serverUUID] = server
	return serverUUID
}

func (sp *ServerPool) GetServer(uuid string) (*Server, error) {
	if server, ok := sp.Servers[uuid]; ok {
		return server, nil
	}
	return nil, fmt.Errorf("This server %s not exist", uuid)
}

func (sp *ServerPool) GetAllServer() []*Server {
	var servers []*Server
	for _ , s := range sp.Servers {
		servers = append(servers, s)
	}
	return  servers
}