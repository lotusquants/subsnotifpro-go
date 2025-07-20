package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsRegistry holds all application metrics
type MetricsRegistry struct {
	// Event processing metrics
	ProcessedEvents     *prometheus.CounterVec
	FailedEvents        *prometheus.CounterVec
	EventProcessingTime *prometheus.HistogramVec

	// Queue metrics
	DLQSize       prometheus.Gauge
	QueueSize     *prometheus.GaugeVec
	QueueMessages *prometheus.CounterVec

	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Authentication metrics
	AuthenticationTotal   *prometheus.CounterVec
	AuthenticationLatency *prometheus.HistogramVec

	// Database metrics
	DatabaseConnections  *prometheus.GaugeVec
	DatabaseQueries      *prometheus.CounterVec
	DatabaseQueryLatency *prometheus.HistogramVec

	// Business metrics
	SubscriptionEvents  *prometheus.CounterVec
	Revenue             *prometheus.CounterVec
	ActiveSubscriptions *prometheus.GaugeVec

	// System metrics
	SystemInfo         *prometheus.GaugeVec
	ProcessCPUUsage    prometheus.Gauge
	ProcessMemoryUsage prometheus.Gauge

	// Rate limiting metrics
	RateLimitHits    *prometheus.CounterVec
	RateLimitAllowed *prometheus.CounterVec
}

// NewMetricsRegistry creates a new metrics registry (backward compatibility)
func NewMetricsRegistry() *MetricsRegistry {
	return GetOrCreateRegistry()
}

// Global registry instance
var Registry *MetricsRegistry
var registryInitialized bool

// GetOrCreateRegistry returns the global registry instance, creating it if it doesn't exist
func GetOrCreateRegistry() *MetricsRegistry {
	if Registry == nil {
		Registry = NewMetricsRegistryInstance()
	}
	return Registry
}

// NewMetricsRegistryInstance creates a new metrics registry with safe registration
func NewMetricsRegistryInstance() *MetricsRegistry {
	if registryInitialized {
		return Registry
	}

	registry := &MetricsRegistry{
		// Event processing metrics
		ProcessedEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subsnotifpro_events_processed_total",
				Help: "Total number of processed events",
			},
			[]string{"event_type", "platform", "status"},
		),

		FailedEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subsnotifpro_events_failed_total",
				Help: "Total number of failed events",
			},
			[]string{"event_type", "platform", "error_type"},
		),

		EventProcessingTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "subsnotifpro_event_processing_duration_seconds",
				Help:    "Time taken to process events",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"event_type", "platform"},
		),

		// Queue metrics
		DLQSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "subsnotifpro_dlq_size",
				Help: "Current size of dead letter queue",
			},
		),

		QueueSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "subsnotifpro_queue_size",
				Help: "Current size of message queues",
			},
			[]string{"queue_name"},
		),

		QueueMessages: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subsnotifpro_queue_messages_total",
				Help: "Total number of messages processed by queue",
			},
			[]string{"queue_name", "status"},
		),

		// HTTP metrics
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subsnotifpro_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status_code"},
		),

		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "subsnotifpro_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),

		HTTPResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "subsnotifpro_http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path"},
		),

		// Authentication metrics
		AuthenticationTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subsnotifpro_authentication_total",
				Help: "Total number of authentication attempts",
			},
			[]string{"method", "status"},
		),

		AuthenticationLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "subsnotifpro_authentication_duration_seconds",
				Help:    "Authentication duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),

		// Database metrics
		DatabaseConnections: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "subsnotifpro_database_connections",
				Help: "Current number of database connections",
			},
			[]string{"status"},
		),

		DatabaseQueries: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subsnotifpro_database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),

		DatabaseQueryLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "subsnotifpro_database_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),

		// Business metrics
		SubscriptionEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subsnotifpro_subscription_events_total",
				Help: "Total number of subscription events",
			},
			[]string{"event_type", "platform", "product_id"},
		),

		Revenue: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subsnotifpro_revenue_total",
				Help: "Total revenue generated",
			},
			[]string{"platform", "product_id", "currency"},
		),

		ActiveSubscriptions: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "subsnotifpro_active_subscriptions",
				Help: "Current number of active subscriptions",
			},
			[]string{"platform", "product_id"},
		),

		// System metrics
		SystemInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "subsnotifpro_system_info",
				Help: "System information",
			},
			[]string{"version", "go_version", "platform"},
		),

		ProcessCPUUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "subsnotifpro_process_cpu_usage_percent",
				Help: "Current CPU usage percentage",
			},
		),

		ProcessMemoryUsage: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "subsnotifpro_process_memory_usage_bytes",
				Help: "Current memory usage in bytes",
			},
		),

		// Rate limiting metrics
		RateLimitHits: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subsnotifpro_rate_limit_hits_total",
				Help: "Total number of rate limit hits",
			},
			[]string{"client_ip", "endpoint"},
		),

		RateLimitAllowed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "subsnotifpro_rate_limit_allowed_total",
				Help: "Total number of allowed requests",
			},
			[]string{"client_ip", "endpoint"},
		),
	}

	// Only register metrics if not already registered
	if !registryInitialized {
		safeRegister(registry)
		registryInitialized = true
	}

	return registry
}

// safeRegister registers metrics safely, avoiding duplicate registration
func safeRegister(registry *MetricsRegistry) {
	metrics := []prometheus.Collector{
		registry.ProcessedEvents,
		registry.FailedEvents,
		registry.EventProcessingTime,
		registry.DLQSize,
		registry.QueueSize,
		registry.QueueMessages,
		registry.HTTPRequestsTotal,
		registry.HTTPRequestDuration,
		registry.HTTPResponseSize,
		registry.AuthenticationTotal,
		registry.AuthenticationLatency,
		registry.DatabaseConnections,
		registry.DatabaseQueries,
		registry.DatabaseQueryLatency,
		registry.SubscriptionEvents,
		registry.Revenue,
		registry.ActiveSubscriptions,
		registry.SystemInfo,
		registry.ProcessCPUUsage,
		registry.ProcessMemoryUsage,
		registry.RateLimitHits,
		registry.RateLimitAllowed,
	}

	for _, metric := range metrics {
		if err := prometheus.Register(metric); err != nil {
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				// Log error without logger dependency
				fmt.Printf("Failed to register metric: %v\n", err)
			}
		}
	}
}

// init initializes the global metrics registry safely
func init() {
	Registry = GetOrCreateRegistry()
}

// RecordEventProcessed records a successfully processed event
func RecordEventProcessed(eventType, platform, status string) {
	Registry.ProcessedEvents.WithLabelValues(eventType, platform, status).Inc()
}

// RecordEventFailed records a failed event
func RecordEventFailed(eventType, platform, errorType string) {
	Registry.FailedEvents.WithLabelValues(eventType, platform, errorType).Inc()
}

// RecordEventProcessingTime records the time taken to process an event
func RecordEventProcessingTime(eventType, platform string, duration time.Duration) {
	Registry.EventProcessingTime.WithLabelValues(eventType, platform).Observe(duration.Seconds())
}

// RecordHTTPRequest records an HTTP request
func RecordHTTPRequest(method, path, statusCode string, duration time.Duration, responseSize int64) {
	Registry.HTTPRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
	Registry.HTTPRequestDuration.WithLabelValues(method, path, statusCode).Observe(duration.Seconds())
	Registry.HTTPResponseSize.WithLabelValues(method, path, statusCode).Observe(float64(responseSize))
}

// RecordAuthentication records an authentication attempt
func RecordAuthentication(method, status string, duration time.Duration) {
	Registry.AuthenticationTotal.WithLabelValues(method, status).Inc()
	Registry.AuthenticationLatency.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordDatabaseQuery records a database query
func RecordDatabaseQuery(database, operation, status string, duration time.Duration) {
	Registry.DatabaseQueries.WithLabelValues(database, operation, status).Inc()
	Registry.DatabaseQueryLatency.WithLabelValues(database, operation).Observe(duration.Seconds())
}

// RecordSubscriptionEvent records a subscription event
func RecordSubscriptionEvent(platform, eventType, productID string) {
	Registry.SubscriptionEvents.WithLabelValues(platform, eventType, productID).Inc()
}

// RecordRevenue records revenue
func RecordRevenue(platform, currency, productID string, amount float64) {
	Registry.Revenue.WithLabelValues(platform, currency, productID).Add(amount)
}

// SetActiveSubscriptions sets the number of active subscriptions
func SetActiveSubscriptions(platform, productID string, count float64) {
	Registry.ActiveSubscriptions.WithLabelValues(platform, productID).Set(count)
}

// SetDLQSize sets the current DLQ size
func SetDLQSize(size float64) {
	Registry.DLQSize.Set(size)
}

// SetQueueSize sets the current queue size
func SetQueueSize(queueName, queueType string, size float64) {
	Registry.QueueSize.WithLabelValues(queueName, queueType).Set(size)
}

// RecordQueueMessage records a message processed from a queue
func RecordQueueMessage(queueName, status string) {
	Registry.QueueMessages.WithLabelValues(queueName, status).Inc()
}

// RecordRateLimit records rate limiting metrics
func RecordRateLimit(clientIP, endpoint string, allowed bool) {
	if allowed {
		Registry.RateLimitAllowed.WithLabelValues(clientIP, endpoint).Inc()
	} else {
		Registry.RateLimitHits.WithLabelValues(clientIP, endpoint).Inc()
	}
}

// SetSystemInfo sets system information
func SetSystemInfo(version, goVersion, platform string) {
	Registry.SystemInfo.WithLabelValues(version, goVersion, platform).Set(1)
}

// MetricsServer provides HTTP server for metrics endpoint
type MetricsServer struct {
	port string
}

// NewMetricsServer creates a new metrics server
func NewMetricsServer(port string) *MetricsServer {
	return &MetricsServer{
		port: port,
	}
}

// Start starts the metrics server
func (s *MetricsServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoint for the metrics server
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
	}

	fmt.Printf("Starting metrics server on port %s\n", s.port)

	go func() {
		<-ctx.Done()
		fmt.Println("Shutting down metrics server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	return server.ListenAndServe()
}
