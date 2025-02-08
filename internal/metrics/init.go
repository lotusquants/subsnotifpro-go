package metrics

import "github.com/prometheus/client_golang/prometheus"

// **Initialize Prometheus Metrics**
func Init() {
	prometheus.MustRegister(ProcessedEvents)
	prometheus.MustRegister(FailedEvents)
	prometheus.MustRegister(DLQSize)
	prometheus.MustRegister(EventProcessingTime)
}
