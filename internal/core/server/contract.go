package server

import (
	"net/url"
	"sync"
)

type LoadBalancer struct {
	Current int
	Mutex   sync.Mutex
}

type Server struct {
	URL       *url.URL
	IsHealthy bool
	Mutex     sync.Mutex
}

type Config struct {
	Port                string   `json:"port"`
	HealthCheckInterval string   `json:"healthCheckInterval"`
	Servers             []string `json:"servers"`
}
