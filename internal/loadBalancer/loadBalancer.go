package loadbalancer

import (
	"go-load-balancer/internal/server"
	"go-load-balancer/internal/strategy"
	"go-load-balancer/internal/utils"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type LoadBalancer struct {
	rwMutex             *sync.RWMutex
	wg                  *sync.WaitGroup
	Schema              string           `json:"schema" yaml:"schema"`
	IpAddress           string           `json:"ipAddress" yaml:"ipAddress"`
	Port                int              `json:"port" yaml:"port"`
	HeartBeatTime       time.Duration    `json:"heartBeatTime" yaml:"heartBeatTime"`
	Strategy            string           `json:"strategy" yaml:"strategy"`
	ServerCount         int              `json:"serverCount" yaml:"serverCount"`
	TotalServers        []*server.Server `json:"totalServers" yaml:"totalServers"`
	UnhealthyServers    []*server.Server `json:"unhealthyServers" yaml:"unhealthyServers"`
	HealthyServers      []*server.Server `json:"healthyServers" yaml:"healthyServers"`
	TotalRequestsServed int              `json:"totalRequestsServed" yaml:"totalRequestsServed"`
	TotalRequestsFailed int              `json:"totalRequestsFailed" yaml:"totalRequestsFailed"`
	StrategyHandler     strategy.StrategyHandler
	LbServer            *http.Server
}

const (
	StrategyRoundRobin               = "round-robin"
	StrategyWeightedRoundRobin       = "weighted-round-robin"
	StrategyRandom                   = "random"
	StrategyLeastConnections         = "least-connections"
	StrategyWeightedLeastConnections = "weighted-least-connections"
)

func NewLoadBalancer(
	schema string,
	ipAddress string,
	port int,
	strategyName string,
	heartBeatTime time.Duration,
	servers []*server.Server,
) *LoadBalancer {

	var strategyHandler strategy.StrategyHandler

	switch strategyName {
	case StrategyRoundRobin:
		strategyHandler = strategy.NewRoundRobinStrategyHandler()
	case StrategyWeightedRoundRobin:
		strategyHandler = strategy.NewRoundRobinStrategyHandler()
	case StrategyRandom:
		strategyHandler = strategy.NewRoundRobinStrategyHandler()
	case StrategyLeastConnections:
		strategyHandler = strategy.NewRoundRobinStrategyHandler()
	case StrategyWeightedLeastConnections:
		strategyHandler = strategy.NewRoundRobinStrategyHandler()
	default:
		strategyHandler = strategy.NewRoundRobinStrategyHandler()
	}

	return &LoadBalancer{
		rwMutex:             &sync.RWMutex{},
		wg:                  &sync.WaitGroup{},
		Schema:              schema,
		HeartBeatTime:       heartBeatTime,
		IpAddress:           ipAddress,
		Port:                port,
		Strategy:            strategyName,
		ServerCount:         len(servers),
		TotalRequestsServed: 0,
		TotalRequestsFailed: 0,
		TotalServers:        servers,
		UnhealthyServers:    []*server.Server{},
		HealthyServers:      []*server.Server{},
		StrategyHandler:     strategyHandler,
	}
}

func (lb *LoadBalancer) Start() {
	lb.rwMutex.Lock()
	defer lb.rwMutex.Unlock()

	mux := http.NewServeMux()
	addr := utils.GetURL(lb.Schema, lb.IpAddress, lb.Port, "")
	lb.LbServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	mux.HandleFunc("/", lb.HandleRequest)

	if err := lb.LbServer.ListenAndServe(); err != nil {
		panic(err)
	}
}

func (lb *LoadBalancer) Stop() {
	lb.rwMutex.Lock()
	defer lb.rwMutex.Unlock()

	for _, svr := range lb.TotalServers {
		lb.wg.Add(1)
		go func(s *server.Server) {
			s.Stop()
			lb.wg.Done()
		}(svr)
	}

	lb.wg.Wait()

	if err := lb.LbServer.Close(); err != nil {
		panic(err)
	}
}

func (lb *LoadBalancer) HandleRequest(w http.ResponseWriter, r *http.Request) {
	server := lb.StrategyHandler.GetNextServer()
	if server == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	url, err := url.Parse(server.ServingUrl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = server.Schema
		req.URL.Host = server.IpAddress
		req.Header = r.Header
	}

	proxy.ServeHTTP(w, r)
}
