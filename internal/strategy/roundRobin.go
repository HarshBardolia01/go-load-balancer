package strategy

import (
	"go-load-balancer/internal/server"
	"sync"
)

type RoundRobinStrategyHandler struct {
	Lock               *sync.Mutex
	CurrentServerIndex int
	Servers            []*server.Server
}

func NewRoundRobinStrategyHandler() StrategyHandler {
	return &RoundRobinStrategyHandler{
		Lock:               &sync.Mutex{},
		CurrentServerIndex: 0,
		Servers:            []*server.Server{},
	}
}

func (rrsh *RoundRobinStrategyHandler) GetNextServer() *server.Server {
	rrsh.CurrentServerIndex = (rrsh.CurrentServerIndex + 1) % len(rrsh.Servers)
	return rrsh.Servers[rrsh.CurrentServerIndex]
}

func (rrsh *RoundRobinStrategyHandler) UpdateServers(servers []*server.Server) {
	rrsh.Servers = servers
}
