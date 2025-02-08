// internal/google_playstore/metrics/server.go
package metrics

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"subsnotifpro-go/internal/logger"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMetricsServer starts the Prometheus metrics server with health check
func StartMetricsServer(ctx context.Context) {
	port := os.Getenv("METRICS_PORT")
	if port == "" {
		port = "9090"
	}
	metricsAddr := fmt.Sprintf(":%s", port)

	srv := &http.Server{
		Addr:    metricsAddr,
		Handler: nil,
	}

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	logger.Log.Infof("📊 Prometheus metrics available at http://localhost:%s/metrics", port)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Log.Fatal("❌ Failed to start Prometheus metrics server:", err)
		}
	}()

	// ✅ Graceful shutdown for metrics server
	go func() {
		<-ctx.Done()
		logger.Log.Warn("🚦 Stopping Prometheus metrics server...")
		srv.Shutdown(context.Background())
	}()
}
