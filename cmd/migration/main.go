package main

import (
	"log"
	"os"

	"github.com/lifenetwork-ai/iam-service/internal/adapters/postgres"
	"github.com/lifenetwork-ai/iam-service/internal/wire/instances"
)

func main() {
	// Initialize database connection
	db := instances.DBConnectionInstance()

	// Get migration scripts path (default or CLI argument)
	basePath := "internal/adapters/postgres/scripts"
	if len(os.Args) > 1 {
		basePath = os.Args[1] // Allow passing a migration path
	}

	log.Printf("Running database migrations from: %s", basePath)

	// Run migrations
	if err := postgres.RunMigrations(db, basePath); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
		os.Exit(1)
	}

	log.Println("Database migration completed successfully.")
}
