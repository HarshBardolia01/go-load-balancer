package server

import (
	"go-load-balancer/internal/utils"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	rwMutex             *sync.RWMutex
	Schema              string `json:"schema" yaml:"schema"`
	ServerName          string `json:"serverName" yaml:"serverName"`
	IpAddress           string `json:"ipAddress" yaml:"ipAddress"`
	Port                int    `json:"port" yaml:"port"`
	ServerType          string `json:"serverType" yaml:"serverType"`
	Weight              int    `json:"weight" yaml:"weight"`
	IsAlive             bool   `json:"isAlive" yaml:"isAlive"`
	TotalRequestsServed int    `json:"totalRequestsServed" yaml:"totalRequestsServed"`
	CurrentConnections  int    `json:"currentConnections" yaml:"currentConnections"`
	MaxConnections      int    `json:"maxConnections" yaml:"maxConnections"`
	HeartBeatEndpoint   string `json:"heartBeatEndpoint" yaml:"heartBeatEndpoint"`
	HeartBeatUrl        string `json:"heartBeatUrl" yaml:"heartBeatUrl"`
	ServingEndpoint     string `json:"servingEndpoint" yaml:"servingEndpoint"`
	ServingUrl          string `json:"servingUrl" yaml:"servingUrl"`
	HttpServer          *http.Server
}

const (
	ServerTypeUnknown = "unknown"
	ServerTypeSlow    = "slow"
	ServerTypeMedium  = "medium"
	ServerTypeFast    = "fast"
)

func NewServer(
	schema string,
	serverName string,
	ipAddress string,
	port int,
	serverType string,
	weight int,
	heartBeatEndpoint string,
	servingEndpoint string,
	maxConnections int,
) *Server {

	heartBeatUrl := utils.GetURL(schema, ipAddress, port, heartBeatEndpoint)
	servingUrl := utils.GetURL(schema, ipAddress, port, servingEndpoint)

	return &Server{
		rwMutex:             &sync.RWMutex{},
		Schema:              schema,
		ServerName:          serverName,
		IpAddress:           ipAddress,
		Port:                port,
		ServerType:          serverType,
		Weight:              weight,
		IsAlive:             true,
		TotalRequestsServed: 0,
		CurrentConnections:  0,
		MaxConnections:      maxConnections,
		HeartBeatEndpoint:   heartBeatEndpoint,
		HeartBeatUrl:        heartBeatUrl,
		ServingEndpoint:     servingEndpoint,
		ServingUrl:          servingUrl,
	}
}

func (s *Server) IsMaxConnectionsReached() bool {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	return s.CurrentConnections >= s.MaxConnections
}

func (s *Server) IncrementCurrentConnections() {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	s.CurrentConnections++
}

func (s *Server) DecrementCurrentConnections() {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	s.CurrentConnections--
}

func (s *Server) IncrementTotalRequestsServed() {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	s.TotalRequestsServed++
}

func (s *Server) ResetTotalRequestsServed() {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	s.TotalRequestsServed = 0
}

func (s *Server) IsServerAlive() bool {
	s.rwMutex.RLock()
	defer s.rwMutex.RUnlock()
	return s.IsAlive
}

func (s *Server) SetServerAlive(isAlive bool) {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()
	s.IsAlive = isAlive
}

func (s *Server) Start() {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	mux := http.NewServeMux()
	addr := utils.GetURL(s.Schema, s.IpAddress, s.Port, "")
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.HttpServer = server

	mux.HandleFunc(s.HeartBeatEndpoint, s.HeartBeatHandler)
	mux.HandleFunc(s.ServingEndpoint, s.ServingHandler)

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func (s *Server) Stop() {
	s.rwMutex.Lock()
	defer s.rwMutex.Unlock()

	if err := s.HttpServer.Close(); err != nil {
		panic(err)
	}

	s.IsAlive = false
}

func (s *Server) HeartBeatHandler(w http.ResponseWriter, r *http.Request) {
	defer s.SetServerAlive(true)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) ServingHandler(w http.ResponseWriter, r *http.Request) {
	defer s.DecrementCurrentConnections()
	defer s.SetServerAlive(true)
	s.IncrementCurrentConnections()

	if s.IsMaxConnectionsReached() {
		w.WriteHeader(http.StatusTooManyRequests)
		return
	}

	var duration time.Duration

	switch s.ServerType {
	case ServerTypeSlow:
		duration = 5 * time.Second
	case ServerTypeMedium:
		duration = 2 * time.Second
	case ServerTypeFast:
		duration = 1 * time.Second
	default:
		duration = 1 * time.Second
	}

	time.Sleep(duration)
	s.IncrementTotalRequestsServed()
	w.WriteHeader(http.StatusOK)
}
