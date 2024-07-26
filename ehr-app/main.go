package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define metrics
var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "my_app_requests_total",
			Help: "Total number of requests received",
		},
		[]string{"method", "status_code"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "my_app_request_duration_seconds",
			Help:    "Histogram of request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	errorCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "my_app_errors_total",
			Help: "Total number of errors",
		},
	)
)

func init() {
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(errorCounter)
}

// home handler
func home(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// Handle the request
	if _, err := fmt.Fprintf(w, "EHR System"); err != nil {
		errorCounter.Inc()
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	duration := time.Since(start).Seconds()
	requestDuration.WithLabelValues(r.Method).Observe(duration)
	requestCounter.WithLabelValues(r.Method, http.StatusText(http.StatusOK)).Inc()
}

// main function
func main() {
	http.HandleFunc("/", home)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8081", nil)
}
