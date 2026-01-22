package server

import (
	"fmt"
	"time"
)

const (
	serverTypeFast   = "fast"
	serverTypeSlow   = "slow"
	serverTypeMedium = "medium"
)

func (s *Server) IncActiveConnections() {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()

	s.activeConnections++
}

func (s *Server) DecActiveConnections() {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()

	s.activeConnections--
}

func (s *Server) IncTotalRequestsServed() {
	s.statsMutex.Lock()
	defer s.statsMutex.Unlock()

	s.totalRequestsServed++
}

func (s *Server) GetActiveConnections() int64 {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	return s.activeConnections
}

func (s *Server) GetTotalRequestsServed() int64 {
	s.statsMutex.RLock()
	defer s.statsMutex.RUnlock()

	return s.totalRequestsServed
}

func (s *Server) GetWeight() int {
	s.infoMutex.RLock()
	defer s.infoMutex.RUnlock()

	return s.Weight
}

func (s *Server) SetWeight(weight int) {
	s.infoMutex.Lock()
	defer s.infoMutex.Unlock()

	s.Weight = weight
}

func (s *Server) GetServerType() string {
	s.infoMutex.RLock()
	defer s.infoMutex.RUnlock()

	return s.ServerType
}

func (s *Server) SetServerType(serverType string) {
	s.infoMutex.Lock()
	defer s.infoMutex.Unlock()

	s.ServerType = serverType
}

func (s *Server) SetIsHeartbeatRunning(isHeartbeatRunning bool) {
	s.infoMutex.Lock()
	defer s.infoMutex.Unlock()

	s.isHeartbeatRunning = isHeartbeatRunning
}

func (s *Server) GetIsHeartbeatRunning() bool {
	s.infoMutex.RLock()
	defer s.infoMutex.RUnlock()

	return s.isHeartbeatRunning
}

func (s *Server) GetWaitTime() time.Duration {
	serverType := s.GetServerType()

	switch serverType {
	case serverTypeFast:
		return time.Second * 2
	case serverTypeSlow:
		return time.Second * 6
	case serverTypeMedium:
		return time.Second * 4
	default:
		return time.Second * 1
	}
}

func GetURL(schema string, host string, port int, endpoint string) string {
	return fmt.Sprintf("%s://%s:%d%s", schema, host, port, endpoint)
}

