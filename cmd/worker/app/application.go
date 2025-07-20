package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/pkg/messaging"
	"subsnotifpro-go/queue"

	"github.com/streadway/amqp"
)

// BuildInfo contains build-time information for observability
type BuildInfo struct {
	Version   string
	GitCommit string
	BuildTime string
	GoVersion string
}

// WorkerDependencies contains all the dependencies needed by the worker
type WorkerDependencies struct {
	// Messaging
	MessagePublisher messaging.MessagePublisher
	RabbitMQManager  *queue.RabbitMQManager
	Channel          *amqp.Channel

	// Consumers (interfaces to be defined)
	Consumers []WorkerConsumer
}

// WorkerConsumer represents a consumer that can be started and stopped
type WorkerConsumer interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Name() string
}

// Application represents the cloud-native worker application framework
type Application struct {
	// Core configuration
	config *config.Config
	logger logger.Logger

	// Build information for observability
	buildInfo BuildInfo

	// Worker dependencies
	dependencies *WorkerDependencies

	// Lifecycle management
	shutdownFuncs []func() error
	healthChecks  []HealthCheck

	// Worker management
	workerWaitGroup sync.WaitGroup
	shutdownSignal  chan struct{}
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) error

// ApplicationOption represents a configuration option for the application
type ApplicationOption func(*Application) error

// NewApplication creates a new worker application instance with proper configuration
func NewApplication(ctx context.Context, buildInfo BuildInfo, opts ...ApplicationOption) (*Application, error) {
	// Load and validate configuration
	cfg := config.LoadConfig()
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Initialize enhanced logger
	contextLogger := logger.NewEnhancedLogger()

	app := &Application{
		config:         cfg,
		logger:         contextLogger,
		buildInfo:      buildInfo,
		shutdownFuncs:  make([]func() error, 0),
		healthChecks:   make([]HealthCheck, 0),
		shutdownSignal: make(chan struct{}),
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(app); err != nil {
			return nil, fmt.Errorf("application option failed: %w", err)
		}
	}

	app.logger.Info("🚀 Worker application framework initialized",
		logger.F("version", buildInfo.Version),
		logger.F("commit", buildInfo.GitCommit),
		logger.F("build_time", buildInfo.BuildTime),
		logger.F("messaging_type", string(cfg.MessagingType)),
	)

	return app, nil
}

// WithDependencies sets the worker dependencies for the application
func WithDependencies(deps *WorkerDependencies) ApplicationOption {
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

// Initialize prepares the worker application for startup
func (app *Application) Initialize(ctx context.Context) error {
	app.logger.Info("🔧 Initializing worker application components...")

	// Validate that dependencies are set
	if app.dependencies == nil {
		return fmt.Errorf("worker dependencies must be set")
	}

	// Initialize messaging components
	if err := app.initializeMessaging(ctx); err != nil {
		return fmt.Errorf("messaging initialization failed: %w", err)
	}

	app.logger.Info("✅ Worker application initialization completed")
	return nil
}

// Run starts the worker application and handles graceful shutdown
func (app *Application) Run(ctx context.Context) error {
	// Ensure application is initialized
	if app.dependencies == nil {
		if err := app.Initialize(ctx); err != nil {
			return err
		}
	}

	// Create context that can be cancelled for graceful shutdown
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start all consumers
	if err := app.startConsumers(workerCtx); err != nil {
		return fmt.Errorf("failed to start consumers: %w", err)
	}

	// Log successful startup
	app.logStartupSuccess(ctx)

	// Wait for shutdown signal
	return app.waitForShutdown(workerCtx, cancel)
}

// Shutdown gracefully shuts down the worker application
func (app *Application) Shutdown(ctx context.Context) error {
	app.logger.Info("🚦 Starting graceful worker shutdown...")

	// Signal shutdown to all components
	close(app.shutdownSignal)

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Stop all consumers
	if err := app.stopConsumers(shutdownCtx); err != nil {
		app.logger.Error("❌ Error stopping consumers", logger.F("error", err.Error()))
	}

	// Wait for all workers to finish with timeout
	workersDone := make(chan struct{})
	go func() {
		app.workerWaitGroup.Wait()
		close(workersDone)
	}()

	select {
	case <-workersDone:
		app.logger.Info("✅ All workers stopped gracefully")
	case <-shutdownCtx.Done():
		app.logger.Warn("⚠️ Worker shutdown timeout reached")
	}

	// Run cleanup functions
	for i, cleanup := range app.shutdownFuncs {
		if err := cleanup(); err != nil {
			app.logger.Error("❌ Cleanup function failed",
				logger.F("index", i),
				logger.F("error", err.Error()))
		}
	}

	app.logger.Info("✅ Graceful worker shutdown completed")
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

// GetDependencies returns the worker dependencies
func (app *Application) GetDependencies() *WorkerDependencies {
	return app.dependencies
}

// Private methods

func (app *Application) initializeMessaging(ctx context.Context) error {
	if app.dependencies.MessagePublisher == nil {
		return fmt.Errorf("message publisher is required")
	}

	// Initialize RabbitMQ specific components if needed
	if app.config.MessagingType == config.MessagingTypeRabbitMQ {
		if app.dependencies.RabbitMQManager == nil {
			return fmt.Errorf("RabbitMQ manager is required for RabbitMQ messaging")
		}
		if app.dependencies.Channel == nil {
			return fmt.Errorf("RabbitMQ channel is required for RabbitMQ messaging")
		}
	}

	app.logger.Info("📨 Messaging components initialized",
		logger.F("type", string(app.config.MessagingType)),
		logger.F("consumers_count", len(app.dependencies.Consumers)),
	)

	return nil
}

func (app *Application) startConsumers(ctx context.Context) error {
	if len(app.dependencies.Consumers) == 0 {
		app.logger.Warn("⚠️ No consumers configured")
		return nil
	}

	app.logger.Info("🚀 Starting worker consumers...",
		logger.F("count", len(app.dependencies.Consumers)))

	for _, consumer := range app.dependencies.Consumers {
		consumer := consumer // Capture loop variable
		app.workerWaitGroup.Add(1)

		go func() {
			defer app.workerWaitGroup.Done()

			app.logger.Info("🔄 Starting consumer", logger.F("name", consumer.Name()))

			if err := consumer.Start(ctx); err != nil {
				app.logger.Error("❌ Consumer error",
					logger.F("name", consumer.Name()),
					logger.F("error", err.Error()))
			}

			app.logger.Info("🛑 Consumer stopped", logger.F("name", consumer.Name()))
		}()
	}

	return nil
}

func (app *Application) stopConsumers(ctx context.Context) error {
	app.logger.Info("🛑 Stopping worker consumers...",
		logger.F("count", len(app.dependencies.Consumers)))

	for _, consumer := range app.dependencies.Consumers {
		if err := consumer.Stop(ctx); err != nil {
			app.logger.Error("❌ Error stopping consumer",
				logger.F("name", consumer.Name()),
				logger.F("error", err.Error()))
		}
	}

	return nil
}

func (app *Application) logStartupSuccess(ctx context.Context) {
	app.logger.Info("✅ SubsNotifPro Worker startup complete - All systems operational")
	app.logger.Info("====================================================")
	app.logger.Info("⚙️ Worker application running", logger.F("version", app.buildInfo.Version))
	app.logger.Info("📨 Messaging backend", logger.F("type", string(app.config.MessagingType)))
	app.logger.Info("🔄 Active consumers", logger.F("count", len(app.dependencies.Consumers)))
	for _, consumer := range app.dependencies.Consumers {
		app.logger.Info("  📦 Consumer active", logger.F("name", consumer.Name()))
	}
	app.logger.Info("====================================================")
}

func (app *Application) waitForShutdown(ctx context.Context, cancel context.CancelFunc) error {
	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	select {
	case sig := <-sigChan:
		app.logger.Info("🚦 Shutdown signal received", logger.F("signal", sig.String()))
		cancel() // Cancel the worker context
		return app.Shutdown(ctx)
	case <-ctx.Done():
		app.logger.Info("🚦 Context cancelled, shutting down")
		return app.Shutdown(ctx)
	}
}

// RunHealthChecks runs all configured health checks
func (app *Application) RunHealthChecks(ctx context.Context) error {
	for i, check := range app.healthChecks {
		if err := check(ctx); err != nil {
			return fmt.Errorf("health check %d failed: %w", i, err)
		}
	}
	return nil
}

// Utility functions

func validateConfig(cfg *config.Config) error {
	if cfg.MessagingType == "" {
		return fmt.Errorf("messaging type is required")
	}
	if cfg.Database.Type == "" {
		return fmt.Errorf("database type is required")
	}
	return nil
}
