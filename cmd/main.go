package main

import (
	"log"
	"os"
	"strings"

	app "github.com/lifenetwork-ai/iam-service/cmd/app"
	"github.com/lifenetwork-ai/iam-service/conf"
	_ "github.com/lifenetwork-ai/iam-service/docs"
	"github.com/lifenetwork-ai/iam-service/internal/adapters/postgres"
	"github.com/lifenetwork-ai/iam-service/internal/wire/instances"
)

// @title IAM Service API
// @version 1.0
// @description Identity and Access Management Service
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@lifenetwork.ai
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @BasePath /
// @securityDefinitions.basic BasicAuth
func main() {
	config := conf.GetConfiguration()

	// Conditionally run DB migrations when AUTO_MIGRATE=true
	if strings.EqualFold(os.Getenv("AUTO_MIGRATE"), "true") {
		db := instances.DBConnectionInstance()
		basePath := os.Getenv("MIGRATIONS_PATH")
		if basePath == "" {
			basePath = "internal/adapters/postgres/scripts"
		}
		log.Printf("AUTO_MIGRATE=true: running database migrations from: %s", basePath)
		if err := postgres.RunMigrations(db, basePath); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
		log.Println("Database migration completed successfully.")
	}

	app.RunApp(config)
}
