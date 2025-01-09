package migrations

import (
	"log"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

func applySQLScript(db *gorm.DB, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := db.Exec(string(content)).Error; err != nil {
		return err
	}

	log.Printf("Successfully applied migration: %s", filepath.Base(filePath))
	return nil
}

func RunMigrations(db *gorm.DB) error {
	log.Println("Running migrations...")

	sqlScripts := []string{
		"./migrations/sql/01_accounts.sql",
		"./migrations/sql/02_data_owners.sql",
		"./migrations/sql/03_data_utilizers.sql",
		"./migrations/sql/04_validators.sql",
		"./migrations/sql/05_refresh_tokens.sql",
		"./migrations/sql/06_data_access_requests.sql",
	}

	for _, script := range sqlScripts {
		if err := applySQLScript(db, script); err != nil {
			return err
		}
	}

	log.Println("Migrations completed.")
	return nil
}
