package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/health"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/routes"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// BuildInfo contains build-time information for observability
type BuildInfo struct {
	Version   string
	GitCommit string
	BuildTime string
	GoVersion string
}

// Application represents the cloud-native application framework
type Application struct {
	// Core configuration
	config *config.Config
	logger logger.Logger

	// Build information for observability
	buildInfo BuildInfo

	// HTTP servers
	httpServer    *http.Server
	metricsServer *http.Server

	// Application dependencies (to be injected)
	dependencies *routes.RouteDependencies

	// Lifecycle management
	shutdownFuncs []func() error
	healthChecks  []HealthCheck
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) error

// ApplicationOption represents a configuration option for the application
type ApplicationOption func(*Application) error

// NewApplication creates a new application instance with proper configuration
func NewApplication(ctx context.Context, buildInfo BuildInfo, opts ...ApplicationOption) (*Application, error) {
	// Load and validate configuration
	cfg := config.LoadConfig()
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Initialize enhanced logger
	contextLogger := logger.NewEnhancedLogger()

	app := &Application{
		config:        cfg,
		logger:        contextLogger,
		buildInfo:     buildInfo,
		shutdownFuncs: make([]func() error, 0),
		healthChecks:  make([]HealthCheck, 0),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(app); err != nil {
			return nil, fmt.Errorf("application option failed: %w", err)
		}
	}

	app.logger.Info("🚀 Application framework initialized",
		logger.F("version", buildInfo.Version),
		logger.F("commit", buildInfo.GitCommit),
		logger.F("build_time", buildInfo.BuildTime),
	)

	return app, nil
}

// WithDependencies sets the route dependencies for the application
func WithDependencies(deps *routes.RouteDependencies) ApplicationOption {
	return func(app *Application) error {
		app.dependencies = deps
		return nil
	}
}

// WithHealthCheck adds a health check to the application
func WithHealthCheck(check HealthCheck) ApplicationOption {
	return func(app *Application) error {
		app.healthChecks = append(app.healthChecks, check)
		return nil
	}
}

// WithShutdownFunc adds a shutdown function to the application
func WithShutdownFunc(fn func() error) ApplicationOption {
	return func(app *Application) error {
		app.shutdownFuncs = append(app.shutdownFuncs, fn)
		return nil
	}
}

// Initialize prepares the application for startup
func (app *Application) Initialize(ctx context.Context) error {
	app.logger.Info("🔧 Initializing application components...")

	// Validate that dependencies are set
	if app.dependencies == nil {
		return fmt.Errorf("application dependencies must be set")
	}

	// Initialize HTTP servers
	if err := app.initializeServers(ctx); err != nil {
		return fmt.Errorf("server initialization failed: %w", err)
	}

	app.logger.Info("✅ Application initialization completed")
	return nil
}

// Run starts the application and handles graceful shutdown
func (app *Application) Run(ctx context.Context) error {
	// Ensure application is initialized
	if app.httpServer == nil {
		if err := app.Initialize(ctx); err != nil {
			return err
		}
	}

	// Create error channel for server goroutines
	errChan := make(chan error, 2)

	// Start HTTP server
	go func() {
		app.logger.Info("🚀 Starting HTTP server", logger.F("port", app.config.ServerPort))
		if err := app.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Start metrics server
	go func() {
		app.logger.Info("📊 Starting metrics server", logger.F("port", "9090"))
		if err := app.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("metrics server error: %w", err)
		}
	}()

	// Log successful startup
	app.logStartupSuccess(ctx)

	// Wait for shutdown signal or error
	return app.waitForShutdown(ctx, errChan)
}

// Shutdown gracefully shuts down the application
func (app *Application) Shutdown(ctx context.Context) error {
	app.logger.Info("🚦 Starting graceful shutdown...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Shutdown HTTP servers
	if app.httpServer != nil {
		if err := app.httpServer.Shutdown(shutdownCtx); err != nil {
			app.logger.Error("❌ HTTP server shutdown error", logger.F("error", err.Error()))
		}
	}

	if app.metricsServer != nil {
		if err := app.metricsServer.Shutdown(shutdownCtx); err != nil {
			app.logger.Error("❌ Metrics server shutdown error", logger.F("error", err.Error()))
		}
	}

	// Run cleanup functions
	for i, cleanup := range app.shutdownFuncs {
		if err := cleanup(); err != nil {
			app.logger.Error("❌ Cleanup function failed",
				logger.F("index", i),
				logger.F("error", err.Error()))
		}
	}

	app.logger.Info("✅ Graceful shutdown completed")
	return nil
}

// GetConfig returns the application configuration
func (app *Application) GetConfig() *config.Config {
	return app.config
}

// GetLogger returns the application logger
func (app *Application) GetLogger() logger.Logger {
	return app.logger
}

// GetBuildInfo returns the application build information
func (app *Application) GetBuildInfo() BuildInfo {
	return app.buildInfo
}

// Private methods

func (app *Application) initializeServers(ctx context.Context) error {
	// Create main application router
	router := routes.SetupRouter(app.dependencies)

	// Add enhanced middleware
	router.Use(app.createCorrelationMiddleware())
	router.Use(app.createLoggingMiddleware())

	// Configure main HTTP server
	app.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%s", app.config.ServerPort),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create metrics server
	app.metricsServer = app.createMetricsServer()

	return nil
}

func (app *Application) createMetricsServer() *http.Server {
	metricsRouter := gin.New()
	metricsRouter.Use(gin.Recovery())

	// Prometheus metrics endpoint
	metricsRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoint
	metricsRouter.GET("/health", app.healthCheckHandler)

	// Readiness check endpoint
	metricsRouter.GET("/ready", app.readinessCheckHandler)

	// Liveness check endpoint
	metricsRouter.GET("/live", app.livenessCheckHandler)

	return &http.Server{
		Addr:         ":9090",
		Handler:      metricsRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
}

func (app *Application) createCorrelationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = fmt.Sprintf("req-%d", time.Now().UnixNano())
		}
		c.Header("X-Correlation-ID", correlationID)
		c.Set("correlation_id", correlationID)
		c.Next()
	}
}

func (app *Application) createLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()
		correlationID, _ := c.Get("correlation_id")

		correlationIDStr := ""
		if correlationID != nil {
			correlationIDStr = fmt.Sprintf("%v", correlationID)
		}

		app.logger.Info("HTTP request completed",
			logger.F("method", method),
			logger.F("path", path),
			logger.F("status", status),
			logger.F("duration_ms", duration.Milliseconds()),
			logger.F("correlation_id", correlationIDStr),
		)
	}
}

func (app *Application) healthCheckHandler(c *gin.Context) {
	ctx := c.Request.Context()
	status := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   app.buildInfo.Version,
		"service":   "subsnotifpro-api",
	}

	// Run health checks
	healthy := true
	for i, check := range app.healthChecks {
		if err := check(ctx); err != nil {
			healthy = false
			status[fmt.Sprintf("check_%d", i)] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		}
	}

	// Check basic health from dependencies
	if app.dependencies != nil && app.dependencies.HealthChecker != nil {
		dbHealth := app.dependencies.HealthChecker.CheckDatabase(ctx)
		if dbHealth.Status != health.StatusHealthy {
			healthy = false
			status["database"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  dbHealth.Message,
			}
		} else {
			status["database"] = map[string]interface{}{
				"status": "healthy",
			}
		}

		if app.config.MessagingType == config.MessagingTypeRabbitMQ {
			rabbitmqHealth := app.dependencies.HealthChecker.CheckRabbitMQ(ctx)
			if rabbitmqHealth.Status != health.StatusHealthy {
				healthy = false
				status["messaging"] = map[string]interface{}{
					"status": "unhealthy",
					"error":  rabbitmqHealth.Message,
				}
			} else {
				status["messaging"] = map[string]interface{}{
					"status": "healthy",
				}
			}
		}
	}

	status["status"] = "healthy"
	if !healthy {
		status["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, status)
		return
	}

	c.JSON(http.StatusOK, status)
}

func (app *Application) readinessCheckHandler(c *gin.Context) {
	ready := true
	status := gin.H{
		"ready":     ready,
		"timestamp": time.Now().UTC(),
		"version":   app.buildInfo.Version,
	}

	// Check if all critical components are ready
	if app.dependencies == nil {
		ready = false
		status["dependencies"] = "not_ready"
	}

	status["ready"] = ready

	if ready {
		c.JSON(http.StatusOK, status)
	} else {
		c.JSON(http.StatusServiceUnavailable, status)
	}
}

func (app *Application) livenessCheckHandler(c *gin.Context) {
	// Simple liveness check - if we can respond, we're alive
	c.JSON(http.StatusOK, gin.H{
		"alive":     true,
		"timestamp": time.Now().UTC(),
		"version":   app.buildInfo.Version,
	})
}

func (app *Application) logStartupSuccess(ctx context.Context) {
	app.logger.Info("✅ SubsNotifPro API startup complete - All systems operational")
	app.logger.Info("====================================================")
	app.logger.Info("🔗 HTTP server listening", logger.F("port", app.config.ServerPort))
	app.logger.Info("📊 Metrics server listening", logger.F("port", "9090"))
	app.logger.Info("📦 Messaging backend", logger.F("type", string(app.config.MessagingType)))
	app.logger.Info("====================================================")
}

func (app *Application) waitForShutdown(ctx context.Context, errChan chan error) error {
	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-errChan:
		app.logger.Error("❌ Server error occurred", logger.F("error", err))
		return err
	case sig := <-sigChan:
		app.logger.Info("🚦 Shutdown signal received", logger.F("signal", sig.String()))
		return app.Shutdown(ctx)
	}
}

// Utility functions

func validateConfig(cfg *config.Config) error {
	if cfg.ServerPort == "" {
		return fmt.Errorf("server port is required")
	}
	if cfg.Database.Type == "" {
		return fmt.Errorf("database type is required")
	}
	if cfg.MessagingType == "" {
		return fmt.Errorf("messaging type is required")
	}
	return nil
}
