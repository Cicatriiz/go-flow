package core

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ComponentLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "goflow_component_latency_seconds",
			Help: "Latency of component execution.",
		},
		[]string{"component"},
	)
	ComponentErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "goflow_component_errors_total",
			Help: "Total number of component errors.",
		},
		[]string{"component"},
	)
)
