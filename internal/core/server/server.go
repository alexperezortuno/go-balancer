package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

func (lb *LoadBalancer) GetNextServer(servers []*Server) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	for i := 0; i < len(servers); i++ {
		idx := lb.Current % len(servers)
		nextServer := servers[idx]
		lb.Current++

		nextServer.Mutex.Lock()
		isHealthy := nextServer.IsHealthy
		nextServer.Mutex.Unlock()

		if isHealthy {
			return nextServer
		}
	}

	return nil
}

func HealthCheck(s *Server, healthCheckInterval time.Duration) {
	for range time.Tick(healthCheckInterval) {
		res, err := http.Head(s.URL.String())
		s.Mutex.Lock()
		if err != nil || res.StatusCode != http.StatusOK {
			fmt.Printf("%s is down\n", s.URL)
			s.IsHealthy = false
		} else {
			s.IsHealthy = true
		}
		s.Mutex.Unlock()
	}
}

func (s *Server) ReverseProxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}

func LoadConfig(file string) (Config, error) {
	var config Config

	data, err := os.ReadFile(file)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
