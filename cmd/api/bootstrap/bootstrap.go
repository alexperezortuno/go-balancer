package bootstrap

import (
	"github.com/alexperezortuno/go-balancer/internal/core/server"
	"log"
	"net/http"
	"net/url"
	"time"
)

func Run() {
	config, err := server.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading configuration: %s", err.Error())
	}

	healthCheckInterval, err := time.ParseDuration(config.HealthCheckInterval)
	if err != nil {
		log.Fatalf("Invalid health check interval: %s", err.Error())
	}

	var srvs []*server.Server
	for _, serverUrl := range config.Servers {
		u, _ := url.Parse(serverUrl)
		srv := &server.Server{URL: u, IsHealthy: true}
		srvs = append(srvs, srv)
		go server.HealthCheck(srv, healthCheckInterval)
	}

	lb := server.LoadBalancer{Current: 0}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		srv := lb.GetNextServer(srvs)
		if srv == nil {
			http.Error(w, "No healthy server available", http.StatusServiceUnavailable)
			return
		}

		// adding this header just for checking from which server the request is being handled.
		// this is not recommended from security perspective as we don't want to let the client know which server is handling the request.
		w.Header().Add("X-Forwarded-Server", srv.URL.String())
		srv.ReverseProxy().ServeHTTP(w, r)
	})

	go func() {
		log.Println("Starting load balancer on port", config.Port)
		err := http.ListenAndServe(config.Port, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		if err != nil {
			log.Fatalf("Error starting load balancer: %s\n", err.Error())
		}
	}()
	defer Run()
}
