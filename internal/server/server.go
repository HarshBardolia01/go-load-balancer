package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Schema     string
	Host       string
	Port       int
	InstanceId string

	Weight     int
	ServerType string

	ServingEp string

	LbSchema string
	LbHost   string
	LbPort   int

	LbRegisterEp   string
	LbDeregisterEp string
	LbHeartbeatEp  string

	lbRegisterURL   string
	lbDeregisterURL string
	lbHeartbeatURL  string

	activeConnections   int64
	totalRequestsServed int64

	heartbeatInterval  time.Duration
	stopHeartbeatCh    chan struct{}
	startHeartbeatCh   chan struct{}
	isHeartbeatRunning bool

	server *http.Server
	client *http.Client

	statsMutex *sync.RWMutex
	infoMutex  *sync.RWMutex
}

func NewServer(
	schema string,
	host string,
	port int,
	instanceId string,
	weight int,
	serverType string,
	servingEp string,
	lbSchema string,
	lbHost string,
	lbPort int,
	lbRegisterEp string,
	lbDeregisterEp string,
	lbHeartbeatEp string,
) *Server {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	return &Server{
		Schema:              schema,
		Host:                host,
		Port:                port,
		InstanceId:          instanceId,
		infoMutex:           &sync.RWMutex{},
		Weight:              weight,
		ServerType:          serverType,
		ServingEp:           servingEp,
		LbSchema:            lbSchema,
		LbHost:              lbHost,
		LbPort:              lbPort,
		LbRegisterEp:        lbRegisterEp,
		LbDeregisterEp:      lbDeregisterEp,
		LbHeartbeatEp:       lbHeartbeatEp,
		lbRegisterURL:       GetURL(lbSchema, lbHost, lbPort, lbRegisterEp),
		lbDeregisterURL:     GetURL(lbSchema, lbHost, lbPort, lbDeregisterEp),
		lbHeartbeatURL:      GetURL(lbSchema, lbHost, lbPort, lbHeartbeatEp),
		activeConnections:   0,
		totalRequestsServed: 0,
		statsMutex:          &sync.RWMutex{},
		heartbeatInterval:   0,
		stopHeartbeatCh:     make(chan struct{}),
		startHeartbeatCh:    make(chan struct{}),
		isHeartbeatRunning:  false,
		server:              nil,
		client:              client,
	}
}

func (s *Server) GetHandler() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get(s.ServingEp, s.HandleRequest)

	return r
}

func (s *Server) StartHttpServer() error {
	handler := s.GetHandler()

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.Host, s.Port),
		Handler:      handler,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Second * 60,
	}

	fmt.Printf("Server %s has started on %s:%d\n", s.InstanceId, s.Host, s.Port)
	return s.server.ListenAndServe()
}

func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
	defer s.DecActiveConnections()

	s.IncActiveConnections()
	s.IncTotalRequestsServed()

	time.Sleep(s.GetWaitTime())

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Handled Request from server %s\n", s.InstanceId)
}

func (s *Server) SendRequest(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return s.client.Do(req)
}

func (s *Server) RegisterWithLoadBalancer() error {
	registrationInfo := map[string]interface{}{
		"instanceId":      s.InstanceId,
		"host":            s.Host,
		"port":            s.Port,
		"weight":          s.Weight,
		"servingEndpoint": s.ServingEp,
	}

	body, err := json.Marshal(registrationInfo)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := s.SendRequest(ctx, http.MethodPost, s.lbRegisterURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register with load balancer. Status: %s", resp.Status)
	}

	respBytes, _ := io.ReadAll(resp.Body)

	fmt.Printf("Server %s registered with load balancer\n", s.InstanceId)
	fmt.Printf("Response body: %s\n", string(respBytes))

	type RegistrationResponse struct {
		HeartbeatInterval int `json:"heartbeatInterval"`
	}

	var rr RegistrationResponse
	err = json.Unmarshal(respBytes, &rr)

	if err != nil {
		return err
	}

	s.heartbeatInterval = time.Duration(rr.HeartbeatInterval) * time.Second
	return nil
}

func (s *Server) DeregisterFromLoadBalancer() error {
	deregistrationInfo := map[string]interface{}{
		"instanceId": s.InstanceId,
		"host":       s.Host,
		"port":       s.Port,
	}

	body, err := json.Marshal(deregistrationInfo)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := s.SendRequest(ctx, http.MethodDelete, s.lbDeregisterURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to deregister from load balancer. Status: %s", resp.Status)
	}

	fmt.Printf("Server %s deregistered from load balancer\n", s.InstanceId)
	return nil
}

func (s *Server) SendHeartbeatToLoadBalancer() error {
	heartbeatInfo := map[string]interface{}{
		"instanceId": s.InstanceId,
		"host":       s.Host,
		"port":       s.Port,
	}

	body, err := json.Marshal(heartbeatInfo)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := s.SendRequest(ctx, http.MethodPatch, s.lbHeartbeatURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send heartbeat to load balancer. Status: %s", resp.Status)
	}

	return nil
}

func (s *Server) RunHeartbeat(context context.Context) {
	if s.heartbeatInterval <= 0 {
		fmt.Println("Invalid heartbeat interval. Using default value of 30 seconds")
		s.heartbeatInterval = 30 * time.Second
	}

	ticker := time.NewTicker(s.heartbeatInterval)
	running := true

	for {
		select {
		case <-context.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			if running {
				s.SendHeartbeatToLoadBalancer()
			}
		case <-s.stopHeartbeatCh:
			running = false
		case <-s.startHeartbeatCh:
			running = true
		}
	}
}

func (s *Server) StopHeartbeat() {
	if !s.GetIsHeartbeatRunning() {
		fmt.Println("Heartbeat already stopped")
		return
	}

	s.stopHeartbeatCh <- struct{}{}
	s.SetIsHeartbeatRunning(false)
}

func (s *Server) StartHeartbeat() {
	if s.GetIsHeartbeatRunning() {
		fmt.Println("Heartbeat already running")
		return
	}

	s.startHeartbeatCh <- struct{}{}
	s.SetIsHeartbeatRunning(true)
}

func GetServer(config *Config) *Server {
	return NewServer(
		config.Schema,
		config.Host,
		config.Port,
		config.InstanceId,
		config.Weight,
		config.ServerType,
		config.ServingEp,
		config.LbSchema,
		config.LbHost,
		config.LbPort,
		config.LbRegisterEp,
		config.LbDeregisterEp,
		config.LbHeartbeatEp,
	)
}

func (s *Server) GracefulShutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.StopHeartbeat()

	if err := s.DeregisterFromLoadBalancer(); err != nil {
		fmt.Printf("Failed to deregister from load balancer: %v\n", err)
	}

	if s.server != nil {
		return s.server.Shutdown(ctx)
	}

	return nil
}

func (s *Server) Run(ctx context.Context) error {
	// 1. Start HTTP server
	errCh := make(chan error, 1)

	go func() {
		if err := s.StartHttpServer(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// 2. Wait for server to start
	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(500 * time.Millisecond):
	}

	// 3. Register with load balancer
	if err := s.RegisterWithLoadBalancer(); err != nil {
		return err
	}

	// 4. Start heartbeat
	go s.RunHeartbeat(ctx)
	s.SetIsHeartbeatRunning(true)

	// 5. Run CLI
	s.RunCLI()

	// 6. Shutdown HTTP server
	return s.GracefulShutdown()
}
