package database

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/conf"
	"github.com/genefriendway/human-network-iam/infrastructures/interfaces"
	"github.com/genefriendway/human-network-iam/packages/logger"
)

// postgreSQL implements SQLDBConnection interface
type postgreSQL struct {
	config *conf.DatabaseConfiguration
}

// NewPostgreSQLClient creates a new sql client instance
func NewPostgreSQLClient(config *conf.DatabaseConfiguration) interfaces.SQLClient {
	return &postgreSQL{
		config: config,
	}
}

func (pgsql *postgreSQL) getDBConnectionURL() string {
	config := pgsql.config

	// Determine SSL mode based on the configuration
	sslMode := "disable"
	/*
		if config.SSLMode {
			sslMode = "enable"
		}
	*/

	// Format for PostgreSQL connection URL
	return fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		config.DbHost, config.DbPort,
		config.DbUser, config.DbName, config.DbPassword, sslMode)
}

func (pgsql *postgreSQL) Connect() *gorm.DB {
	// Get the PostgreSQL connection URL
	dbUrl := pgsql.getDBConnectionURL()
	var db *gorm.DB
	var err error

	// Open the database connection with a custom log level
	db, err = gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Get the SQL database object
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}

	// Configure the connection pool
	config := pgsql.config
	dbMaxOpenConnsInt, err := strconv.Atoi(config.DbMaxOpenConns)
	if err != nil {
		logger.GetLogger().Fatalf("Failed to convert db max open connections to int: %v", err)
	}

	dbMaxIdleConnsInt, err := strconv.Atoi(config.DbMaxIdleConns)
	if err != nil {
		logger.GetLogger().Fatalf("failed to convert db max idle connection time to int: %v", err)
	}

	dbConnMaxLifetimeInMinuteInt, err := strconv.Atoi(config.DbConnMaxLifetimeInMinute)
	if err != nil {
		logger.GetLogger().Fatalf("failed to convert db max connection lifetime to int: %v", err)
	}

	sqlDB.SetMaxIdleConns(dbMaxIdleConnsInt)                                            // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(dbMaxOpenConnsInt)                                            // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Duration(dbConnMaxLifetimeInMinuteInt) * time.Minute) // Maximum amount of time a connection may be reused

	return db
}
