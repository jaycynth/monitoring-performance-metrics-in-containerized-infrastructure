package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// Define metrics
var (
	cpuUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cpu_usage",
		Help: "CPU usage of the server in percentage",
	}, []string{"instance"})

	memoryUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "memory_usage",
		Help: "Memory usage of the server in percentage",
	}, []string{"instance", "type"})

	networkLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "network_latency_seconds",
			Help:    "Network latency in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"host", "method", "endpoint"},
	)

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
	http.HandleFunc("/", uiHandler)
	http.HandleFunc("/simulate", simulateHandler)

	go func() {
		for {
			// Get real CPU usage
			percent, err := cpu.Percent(0, false)
			if err == nil && len(percent) > 0 {
				cpuUsage.WithLabelValues("localhost").Set(percent[0])
			}

			// Get real memory usage
			memStats, err := mem.VirtualMemory()
			if err == nil {
				memoryUsage.WithLabelValues("localhost", "used").Set(memStats.UsedPercent)
				memoryUsage.WithLabelValues("localhost", "free").Set((float64(memStats.Free) / float64(memStats.Total)) * 100)
				memoryUsage.WithLabelValues("localhost", "total").Set(float64(memStats.Total))
				memoryUsage.WithLabelValues("localhost", "cached").Set((float64(memStats.Cached) / float64(memStats.Total)) * 100)
				memoryUsage.WithLabelValues("localhost", "swap_used").Set((float64(memStats.SwapTotal-memStats.SwapFree) / float64(memStats.SwapTotal)) * 100)
			}

			// Get network latency measurement
			latency := measureNetworkLatency("http://localhost:8081/success")
			networkLatency.WithLabelValues("localhost", "GET", "success").Observe(latency)

			time.Sleep(5 * time.Second)
		}
	}()

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func measureNetworkLatency(url string) float64 {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error measuring network latency: %v", err)
		return 0.0
	}
	defer resp.Body.Close()

	latency := time.Since(start).Seconds() * 1000 // Convert to milliseconds
	return latency
}

func uiHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("ui").Parse(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>Performance Metrics Simulator</title>
			<style>
				body { font-family: Arial, sans-serif; padding: 20px; line-height: 1.6; }
				h1 { margin-bottom: 20px; }
				.container { max-width: 600px; margin: 0 auto; }
				label { display: block; margin-bottom: 10px; }
				input[type="text"], input[type="submit"] { padding: 10px; width: 100%; margin-bottom: 20px; }
				.button-container { margin-bottom: 20px; }
				input[type="submit"] { background-color: #4CAF50; color: white; border: none; cursor: pointer; }
				input[type="submit"]:hover { background-color: #45a049; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>Performance Metrics Simulator</h1>
				<form action="/simulate" method="post">
					<div class="button-container">
						<label for="successCalls">Number of calls to success endpoint:</label>
						<input type="text" id="successCalls" name="successCalls" placeholder="Enter number of success calls">
					</div>
					<div class="button-container">
						<label for="errorCalls">Number of calls to error endpoint:</label>
						<input type="text" id="errorCalls" name="errorCalls" placeholder="Enter number of error calls">
					</div>
					<div class="button-container">
						<input type="submit" name="action" value="Simulate Requests">
					</div>
				</form>
				<form action="/simulate" method="post">
					<div class="button-container">
						<input type="submit" name="spikeCPU" value="Spike CPU Usage">
					</div>
				</form>
				<form action="/simulate" method="post">
					<div class="button-container">
						<input type="submit" name="spikeMemory" value="Spike Memory Usage">
					</div>
				</form>
				<form action="/simulate" method="post">
					<div class="button-container">
						<input type="submit" name="spikeLatency" value="Spike Network Latency">
					</div>
				</form>
			</div>
		</body>
		</html>
	`))
	tmpl.Execute(w, nil)
}

func simulateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()

		if r.FormValue("action") == "Simulate Requests" {
			successCalls, _ := strconv.Atoi(r.FormValue("successCalls"))
			errorCalls, _ := strconv.Atoi(r.FormValue("errorCalls"))

			for i := 0; i < successCalls; i++ {
				go trackHTTPStatus("http://localhost:8081/success", "200")
			}
			for i := 0; i < errorCalls; i++ {
				go trackHTTPStatus("http://localhost:8081/error", "500")
			}

			fmt.Fprintln(w, "Requests simulation triggered successfully!")
		}

		if r.FormValue("spikeCPU") != "" {
			go spikeCPU()
			fmt.Fprintln(w, "CPU spike triggered successfully!")
		}

		if r.FormValue("spikeMemory") != "" {
			go spikeMemory()
			fmt.Fprintln(w, "Memory spike triggered successfully!")
		}

		if r.FormValue("spikeLatency") != "" {
			go simulateNetworkLatency()
			fmt.Fprintln(w, "Network latency simulation triggered successfully!")
		}
	}
}

func trackHTTPStatus(url string, code string) {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error making request to URL %s: %v", url, err)
		return
	}
	defer resp.Body.Close()

	httpStatusCount.WithLabelValues(code).Inc()
}

func spikeCPU() {
	log.Println("Spiking CPU usage...")
	for i := 0; i < 100_000_000; i++ { // Increase the loop count
		_ = rand.Intn(1000) * rand.Intn(1000)
		if i%1_000_000 == 0 { // Reduce the frequency of metric updates
			percent, err := cpu.Percent(0, false)
			if err == nil && len(percent) > 0 {
				cpuUsage.WithLabelValues("localhost").Set(percent[0])
				log.Printf("CPU usage: %.2f%%", percent[0])
			}
		}
	}
}

func spikeMemory() {
	log.Println("Spiking memory usage...")
	for j := 0; j < 5; j++ { //If the spike is too short, Prometheus might miss it so increase the duration of the spike or perform multiple spikes in a loop.
		memBlock := make([]byte, 500_000_000) // Allocate 500MB
		for i := range memBlock {
			memBlock[i] = byte(i % 256) // Write to the allocated memory to ensure it's used. The memory might be optimized away by the Go runtime if not used.
		}

		memStats, err := mem.VirtualMemory()
		if err == nil {
			memoryUsage.WithLabelValues("localhost", "used").Set(memStats.UsedPercent)
			memoryUsage.WithLabelValues("localhost", "free").Set((float64(memStats.Free) / float64(memStats.Total)) * 100)
			memoryUsage.WithLabelValues("localhost", "total").Set(float64(memStats.Total))
			memoryUsage.WithLabelValues("localhost", "cached").Set((float64(memStats.Cached) / float64(memStats.Total)) * 100)
			memoryUsage.WithLabelValues("localhost", "swap_used").Set((float64(memStats.SwapTotal-memStats.SwapFree) / float64(memStats.SwapTotal)) * 100)
		}
		time.Sleep(5 * time.Second)
	}
}

func simulateNetworkLatency() {
	log.Println("Simulating network latency...")
	// Update network latency metric while simulating latency
	for i := 0; i < 10; i++ { // Increased to 10 iterations
		latency := measureNetworkLatency("http://localhost:8081/success")
		networkLatency.WithLabelValues("localhost", "GET", "success").Observe(latency)
		time.Sleep(1 * time.Second) // Reduced sleep to 1 second to observe more frequent data points
	}
}
