package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// Define metrics
// Gauge Options - represent a single numerical value that can arbitrarily go up and down.
// Summary Options - represent values which are used to track the distribution of a set of observed values
// Counter Options -  represent cumulative metrics that represent a single monotonically increasing counter whose value can only increase or be reset to zero on restart. They are typically used to represent the number of requests served, tasks completed, or errors.

var (
	cpuUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cpu_usage",
		Help: "CPU usage of the server in percentage",
	}, []string{"instance"})

	memoryUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "memory_usage",
		Help: "Memory usage of the server in percentage",
	}, []string{"instance"})

	networkLatency = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "network_latency_seconds",
		Help: "Network latency in seconds",
	}, []string{"instance"})

	httpStatusCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_status_count",
		Help: "Count of HTTP status codes returned by the server",
	}, []string{"status"})
)

func init() {
	// Register metrics
	prometheus.MustRegister(cpuUsage)
	prometheus.MustRegister(memoryUsage)
	prometheus.MustRegister(networkLatency)
	prometheus.MustRegister(httpStatusCount)
}

func main() {

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/", handler)

	//Ensure that the HTTP server can continue to serve requests while the CPU usage is being updated.
	go func() {
		for {
			// Get real CPU usage
			percent, err := cpu.Percent(0, false)
			if err == nil && len(percent) > 0 {
				cpuUsage.WithLabelValues("localhost").Set(percent[0]) // Set the first CPU core's usage
			}
			// Get real memory usage
			memStats, err := mem.VirtualMemory()
			if err == nil {
				memoryUsage.WithLabelValues("localhost").Set(memStats.UsedPercent)
			}

			// Get network latency measurement
			latency := measureNetworkLatency("http://localhost:8080/metrics")
			networkLatency.WithLabelValues("localhost").Observe(latency)

			time.Sleep(5 * time.Second)
		}
	}()

	log.Fatal(http.ListenAndServe(":8081", nil))
}

// Simulate network latency measurement
func simulateNetworkLatency() float64 {
	// For demo, we return a random value
	return float64(100+rand.Intn(200)) / 1000
}

// Measure network latency by sending an HTTP request
func measureNetworkLatency(url string) float64 {
	startTime := time.Now()
	resp, err := http.Get(url) // Send a GET request
	if err != nil {
		log.Printf("Error measuring latency: %v", err)
		return 0 // Return 0 if there is an error
	}
	defer resp.Body.Close()

	return time.Since(startTime).Seconds()
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Simulate network latency
	latency := 200 * time.Millisecond
	time.Sleep(latency) // Introduce the delay

	// Simulate different responses based on the request path
	switch r.URL.Path {
	case "/success":
		httpStatusCount.WithLabelValues("200").Inc()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success!"))
	case "/error":
		httpStatusCount.WithLabelValues("500").Inc()
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error!"))
	default:
		httpStatusCount.WithLabelValues("404").Inc()
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found!"))
	}
}
