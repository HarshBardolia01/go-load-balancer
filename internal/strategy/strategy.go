package strategy

import "go-load-balancer/internal/server"

type StrategyHandler interface {
	GetNextServer() *server.Server
	UpdateServers(servers []*server.Server)
}
