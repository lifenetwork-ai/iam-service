package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gorm.io/gorm/logger"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	"github.com/genefriendway/human-network-auth/conf"
	"github.com/genefriendway/human-network-auth/conf/database"
	"github.com/genefriendway/human-network-auth/constants"
	"github.com/genefriendway/human-network-auth/internal/domain"
	"github.com/genefriendway/human-network-auth/internal/dto"
	"github.com/genefriendway/human-network-auth/internal/interfaces"
	"github.com/genefriendway/human-network-auth/internal/middleware"
	routev1 "github.com/genefriendway/human-network-auth/internal/route"
	"github.com/genefriendway/human-network-auth/migrations"
	pkginterfaces "github.com/genefriendway/human-network-auth/pkg/interfaces"
	pkglogger "github.com/genefriendway/human-network-auth/pkg/logger"
	"github.com/genefriendway/human-network-auth/wire"
)

func RunApp(config *conf.Configuration) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger and environment settings
	initializeLoggerAndMode(config)

	// Initialize Gin router with middleware
	r := initializeRouter()

	// Initialize database connection
	db := database.DBConnWithLoglevel(logger.Info)
	if err := migrations.RunMigrations(db, config); err != nil {
		pkglogger.GetLogger().Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize use cases and queue
	authUCase := wire.GetAuthUCase(db, config)
	accountUCase := wire.GetAccountUCase(db, config)
	dataAccessUCase := wire.GetDataAccessUCase(db, config)
	iamUCase := wire.GetIAMUCase(db)

	// Initialize predefined policies
	initializePolicies(iamUCase)

	// Initialize predefined permissions
	initializePermissions(iamUCase)

	// Register routes
	routev1.RegisterRoutes(
		ctx,
		r,
		config,
		db,
		authUCase,
		accountUCase,
		dataAccessUCase,
		iamUCase,
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
	var logLevel pkginterfaces.Level
	switch strings.ToLower(config.LogLevel) {
	case "debug":
		logLevel = pkginterfaces.DebugLevel
		gin.SetMode(gin.DebugMode) // Development mode
	case "info":
		logLevel = pkginterfaces.InfoLevel
		gin.SetMode(gin.ReleaseMode) // Production mode
	default:
		// Default to info level if unspecified or invalid
		logLevel = pkginterfaces.InfoLevel
		gin.SetMode(gin.ReleaseMode)
	}

	// Set the log level in the logger package
	pkglogger.SetLogLevel(logLevel)

	// Retrieve the initialized logger
	appLogger := pkglogger.GetLogger()

	// Log application startup details
	appLogger.Infof("Application '%s' started with log level '%s' in '%s' mode", config.AppName, logLevel, config.Env)

	// Log additional details for debugging
	if logLevel == pkginterfaces.DebugLevel {
		appLogger.Debug("Debugging mode enabled. Verbose logging is active.")
	}
}

func initializePolicies(iamUCase interfaces.IAMUCase) {
	// Predefined policies
	policies := []dto.PolicyPayloadDTO{
		{
			Name:        constants.AdminPolicy.String(),
			Description: "Permissions for administrators",
		},
		{
			Name:        constants.UserPolicy.String(),
			Description: "Permissions for normal users",
		},
		{
			Name:        constants.ValidatorPolicy.String(),
			Description: "Permissions for validators",
		},
		{
			Name:        constants.DataOwnerPolicy.String(),
			Description: "Permissions for data owners",
		},
		{
			Name:        constants.DataUtilizerPolicy.String(),
			Description: "Permissions for data utilizers",
		},
	}

	// Check if policies already exist
	for _, policy := range policies {
		if _, err := iamUCase.CreatePolicy(policy); err != nil {
			if err.Error() == domain.ErrAlreadyExists.Error() {
				pkglogger.GetLogger().Infof("Policy '%s' already exists, skipping...\n", policy.Name)
			} else {
				pkglogger.GetLogger().Fatalf("Failed to initialize policy '%s': %v\n", policy.Name, err)
			}
		} else {
			pkglogger.GetLogger().Infof("Policy '%s' created successfully.\n", policy.Name)
		}
	}
}

// Initialize permissions for predefined policies
func initializePermissions(iamUCase interfaces.IAMUCase) {
	// Predefined permissions
	permissions := []dto.PermissionPayloadDTO{
		// AdminPolicy
		{
			PolicyName:  constants.AdminPolicy.String(),
			Resource:    constants.ResourceAccounts.String(),
			Action:      constants.ActionRead.String(),
			Description: "Allows reading accounts",
		},
		{
			PolicyName:  constants.AdminPolicy.String(),
			Resource:    constants.ResourceAccounts.String(),
			Action:      constants.ActionUpdate.String(),
			Description: "Allows updating accounts",
		},
		{
			PolicyName:  constants.AdminPolicy.String(),
			Resource:    constants.ResourceAccounts.String(),
			Action:      constants.ActionDelete.String(),
			Description: "Allows deleting accounts",
		},
		// ValidatorPolicy
		{
			PolicyName:  constants.ValidatorPolicy.String(),
			Resource:    constants.ResourceAccounts.String(),
			Action:      constants.ActionRead.String(),
			Description: "Allows reading accounts",
		},
		{
			PolicyName:  constants.ValidatorPolicy.String(),
			Resource:    constants.ResourceDataRequests.String(),
			Action:      constants.ActionWrite.String(),
			Description: "Allows creating data requests",
		},
		{
			PolicyName:  constants.ValidatorPolicy.String(),
			Resource:    constants.ResourceDataRequests.String(),
			Action:      constants.ActionRead.String(),
			Description: "Allows reading data requests",
		},
		// DataOwnerPolicy
		{
			PolicyName:  constants.DataOwnerPolicy.String(),
			Resource:    constants.ResourceAccounts.String(),
			Action:      constants.ActionRead.String(),
			Description: "Allows reading accounts",
		},
		{
			PolicyName:  constants.DataOwnerPolicy.String(),
			Resource:    constants.ResourceDataRequests.String(),
			Action:      constants.ActionApprove.String(),
			Description: "Allows approving data requests",
		},
		{
			PolicyName:  constants.DataOwnerPolicy.String(),
			Resource:    constants.ResourceValidators.String(),
			Action:      constants.ActionRead.String(),
			Description: "Allows reading validator details",
		},
		// DataUtilizerPolicy
		{
			PolicyName:  constants.DataUtilizerPolicy.String(),
			Resource:    constants.ResourceAccounts.String(),
			Action:      constants.ActionRead.String(),
			Description: "Allows reading accounts",
		},
		{
			PolicyName:  constants.DataUtilizerPolicy.String(),
			Resource:    constants.ResourceDataRequests.String(),
			Action:      constants.ActionRead.String(),
			Description: "Allows reading data requests",
		},
		// UserPolicy
		{
			PolicyName:  constants.UserPolicy.String(),
			Resource:    constants.ResourceAccounts.String(),
			Action:      constants.ActionRead.String(),
			Description: "Allows reading accounts",
		},
	}

	for _, permission := range permissions {
		if err := iamUCase.CreatePermission(permission); err != nil {
			if err.Error() == domain.ErrAlreadyExists.Error() {
				pkglogger.GetLogger().Infof("Permission '%s:%s' already exists for policy '%s', skipping...\n",
					permission.Resource, permission.Action, permission.PolicyName)
			} else {
				pkglogger.GetLogger().Fatalf("Failed to initialize permission '%s:%s' for policy '%s': %v\n",
					permission.Resource, permission.Action, permission.PolicyName, err)
			}
		} else {
			pkglogger.GetLogger().Infof("Permission '%s:%s' created successfully for policy '%s'.\n",
				permission.Resource, permission.Action, permission.PolicyName)
		}
	}
}

func initializeRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.DefaultPagination())
	r.Use(middleware.RequestLoggerMiddleware())
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
	}()
}

func waitForShutdownSignal(cancel context.CancelFunc) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	pkglogger.GetLogger().Debug("Shutting down gracefully...")
	cancel()
}
