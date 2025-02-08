// internal/google_playstore/metrics/metrics.go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// **Metrics Registry**
var (
	// **Tracks number of events processed successfully**
	ProcessedEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rtdn_events_processed_total",
			Help: "Total number of RTDN events processed successfully",
		},
		[]string{"event_type"}, // ✅ Labels allow event categorization
	)

	// **Tracks number of events failed (moved to DLQ)**
	FailedEvents = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rtdn_events_failed_total",
			Help: "Total number of RTDN events that failed processing",
		},
		[]string{"event_type"},
	)

	// **Tracks size of DLQ queue**
	DLQSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "rtdn_dlq_size",
			Help: "Current number of messages in the DLQ",
		},
	)

	// **Tracks the processing duration of events**
	EventProcessingTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rtdn_event_processing_time_seconds",
			Help:    "Histogram of RTDN event processing time",
			Buckets: prometheus.DefBuckets, // ✅ Default Prometheus Buckets
		},
		[]string{"event_type"},
	)
)
