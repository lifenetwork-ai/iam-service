package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"github.com/genefriendway/human-network-iam/conf"
	middleware "github.com/genefriendway/human-network-iam/internal/delivery/http/middleware"
	routev1 "github.com/genefriendway/human-network-iam/internal/delivery/http/route"
	pkglogger "github.com/genefriendway/human-network-iam/packages/logger"
	pkglogtypes "github.com/genefriendway/human-network-iam/packages/logger/types"
	"github.com/genefriendway/human-network-iam/wire"
	"github.com/genefriendway/human-network-iam/wire/providers"
)

func RunApp(config *conf.Configuration) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger and environment settings
	initializeLoggerAndMode(config)

	// Initialize Gin router with middleware
	r := initializeRouter()

	// Initialize database connection
	db := providers.ProvideDBConnection()

	// Initialize the cache repository
	cacheRepository := providers.ProvideCacheRepository(ctx)

	// Initialize use cases
	ucases := wire.InitializeUseCases(db, cacheRepository)

	// Register routes
	routev1.RegisterRoutes(
		ctx,
		r,
		config,
		db,
		ucases.IdentityOrganizationUCase,
		ucases.IdentityUserUCase,
	)

	// Start server
	startServer(r, config)

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
	var logLevel pkglogtypes.Level
	switch strings.ToLower(config.LogLevel) {
	case "debug":
		logLevel = pkglogtypes.DebugLevel
		gin.SetMode(gin.DebugMode) // Development mode
	case "info":
		logLevel = pkglogtypes.InfoLevel
		gin.SetMode(gin.ReleaseMode) // Production mode
	default:
		// Default to info level if unspecified or invalid
		logLevel = pkglogtypes.InfoLevel
		gin.SetMode(gin.ReleaseMode)
	}

	// Set the log level in the logger package
	pkglogger.SetLogLevel(logLevel)

	// Retrieve the initialized logger
	appLogger := pkglogger.GetLogger()

	// Log application startup details
	appLogger.Infof("Application '%s' started with log level '%s' in '%s' mode", config.AppName, logLevel, config.Env)

	// Log additional details for debugging
	if logLevel == pkglogtypes.DebugLevel {
		appLogger.Debug("Debugging mode enabled. Verbose logging is active.")
	}
}

func initializeRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.DefaultPagination())
	r.Use(middleware.RequestLoggerMiddleware())
	r.Use(middleware.RequestDataGuardMiddleware())
	r.Use(middleware.XHeaderValidationMiddleware())
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
			pkglogger.GetLogger().Fatalf("Failed to run gin router: %v", err)
		}

		pkglogger.GetLogger().Infof("Server started on port %v", config.AppPort)
	}()
}

func waitForShutdownSignal(cancel context.CancelFunc) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	pkglogger.GetLogger().Debug("Shutting down gracefully...")
	cancel()
}
