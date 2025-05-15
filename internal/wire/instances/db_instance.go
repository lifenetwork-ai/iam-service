package instances

import (
	"sync"

	"gorm.io/gorm"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/infrastructures/database"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

var (
	dbOnce     sync.Once
	dbInstance *gorm.DB
)

// DBConnectionInstance provides a singleton instance of the PostgreSQL database connection.
func DBConnectionInstance() *gorm.DB {
	dbOnce.Do(func() {
		logger.GetLogger().Info("Initializing PostgreSQL database connection...")

		// Get the configuration
		config := conf.GetConfiguration()

		// Create a new PostgreSQL client
		pgsqlClient := database.NewPostgreSQLClient(&config.Database)

		// Connect and store the database instance
		dbInstance = pgsqlClient.Connect()

		logger.GetLogger().Info("PostgreSQL database connection established.")
	})
	return dbInstance
}
