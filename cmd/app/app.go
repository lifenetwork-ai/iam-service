package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	_ "github.com/lifenetwork-ai/iam-service/docs" // Import generated docs
	middleware "github.com/lifenetwork-ai/iam-service/internal/delivery/http/middleware"
	routev1 "github.com/lifenetwork-ai/iam-service/internal/delivery/http/route"
	"github.com/lifenetwork-ai/iam-service/internal/wire"
	"github.com/lifenetwork-ai/iam-service/internal/wire/instances"
	"github.com/lifenetwork-ai/iam-service/internal/workers"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"
)

func RunApp(config *conf.Configuration) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger and environment settings
	initializeLoggerAndMode(config)

	// Initialize Gin router with middleware
	r := initializeRouter()

	// Initialize database connection
	db := instances.DBConnectionInstance()

	// Initialize the cache repository
	cacheRepository := instances.CacheRepositoryInstance(ctx)

	// Init repositories
	repos := wire.InitializeRepos(db, cacheRepository)

	// Initialize use cases
	ucases := wire.InitializeUseCases(db, repos)

	// Register routes
	routev1.RegisterRoutes(
		ctx,
		r,
		config,
		ucases,
		repos,
	)

	// Start server
	startServer(r, config)

	// Start workers
	go workers.NewOTPDeliveryWorker(
		ucases.CourierUCase,
		ucases.TenantUCase,
		instances.OTPQueueRepositoryInstance(ctx),
	).Start(ctx, constants.OTPDeliveryWorkerInterval)

	go workers.NewOTPRetryWorker(
		ucases.CourierUCase,
		instances.OTPQueueRepositoryInstance(ctx),
	).Start(ctx, constants.OTPRetryWorkerInterval)

	go workers.NewZaloRefreshTokenWorker(
		instances.SMSServiceInstance(repos.ZaloTokenRepo),
	).Start(ctx, 12*time.Hour)

	// Handle shutdown signals
	waitForShutdownSignal(cancel)
}

// Helper Functions
func initializeLoggerAndMode(config *conf.Configuration) {
	// Validate configuration
	if config == nil {
		panic("configuration cannot be nil")
	}

	// Determine the log level from the configuration
	var logLevel logger.Level
	switch strings.ToLower(config.LogLevel) {
	case "debug":
		logLevel = logger.DebugLevel
		gin.SetMode(gin.DebugMode) // Development mode
	case "info":
		logLevel = logger.InfoLevel
		gin.SetMode(gin.ReleaseMode) // Production mode
	default:
		// Default to info level if unspecified or invalid
		logLevel = logger.InfoLevel
		gin.SetMode(gin.ReleaseMode)
	}

	// Set the log level in the logger package
	logger.SetLogLevel(logLevel)

	// Retrieve the initialized logger
	appLogger := logger.GetLogger()

	// Log application startup details
	appLogger.Infof("Application '%s' started with log level '%s' in '%s' mode", config.AppName, logLevel, config.Env)

	// Log additional details for debugging
	if logLevel == logger.DebugLevel {
		appLogger.Debug("Debugging mode enabled. Verbose logging is active.")
	}
}

func initializeRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestTracingMiddleware())
	r.Use(middleware.DefaultPagination())
	r.Use(middleware.RequestLoggerMiddleware())
	r.Use(middleware.RequestDataGuardMiddleware())
	r.Use(gin.Recovery())
	return r
}

func startServer(
	r *gin.Engine,
	config *conf.Configuration,
) {
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("%s is still alive", config.AppName),
		})
	})

	if config.Env != "PROD" {
		r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
	}

	go func() {
		if err := r.Run(fmt.Sprintf("0.0.0.0:%v", config.AppPort)); err != nil {
			logger.GetLogger().Fatalf("Failed to run gin router: %v", err)
		}

		logger.GetLogger().Infof("Server started on port %v", config.AppPort)
	}()
}

func waitForShutdownSignal(cancel context.CancelFunc) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	logger.GetLogger().Debug("Shutting down gracefully...")
	cancel()
}
