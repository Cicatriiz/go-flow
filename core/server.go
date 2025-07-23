package core

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMetricsServer starts an HTTP server to expose the Prometheus metrics.
func StartMetricsServer(addr string) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		fmt.Printf("Metrics server listening on %s\n", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			fmt.Printf("Error starting metrics server: %v\n", err)
		}
	}()
}
