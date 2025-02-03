package migrations

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/conf"
	"github.com/genefriendway/human-network-iam/constants"
	"github.com/genefriendway/human-network-iam/internal/domain"
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

func seedAdminAccount(db *gorm.DB, config *conf.Configuration) error {
	var count int64
	db.Model(&domain.Account{}).Where("role = ?", constants.Admin.String()).Count(&count)

	if count == 0 {
		adminEmail := config.AdminAccount.AdminEmail
		adminPassword := config.AdminAccount.AdminPassword

		if adminEmail == "" || adminPassword == "" {
			return errors.New("missing ADMIN credentials in environment variables")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		hashedPasswordString := string(hashedPassword)

		admin := &domain.Account{
			Email:        adminEmail,
			Username:     "admin",
			PasswordHash: &hashedPasswordString,
			Role:         constants.Admin.String(),
		}

		if err := db.Create(admin).Error; err != nil {
			return fmt.Errorf("failed to create ADMIN: %w", err)
		}
		fmt.Println("ADMIN account created successfully.")
	}

	return nil
}

func RunMigrations(db *gorm.DB, config *conf.Configuration) error {
	log.Println("Running migrations...")

	sqlScripts := []string{
		// "./migrations/sql/01_accounts.sql",
		// "./migrations/sql/02_data_owners.sql",
		// "./migrations/sql/03_data_utilizers.sql",
		// "./migrations/sql/04_validators.sql",
		// "./migrations/sql/05_refresh_tokens.sql",
		// "./migrations/sql/06_data_access_requests.sql",
		// "./migrations/sql/07_data_access_request_requesters.sql",
		// "./migrations/sql/08_iam_permissions.sql",
		// "./migrations/sql/09_account_policies.sql",
		// "./migrations/sql/10_iam_policies.sql",
		// "./migrations/sql/11_file_infos.sql",
		// "./migrations/sql/12_add_requester_request_detail.sql",
	}

	for _, script := range sqlScripts {
		if err := applySQLScript(db, script); err != nil {
			return err
		}
	}

	// // Seed admin account
	// if err := seedAdminAccount(db, config); err != nil {
	// 	return err
	// }

	log.Println("Migrations completed.")
	return nil
}
