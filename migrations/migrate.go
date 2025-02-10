package migrations

import (
	"log"
	"os"
	"path/filepath"

	"gorm.io/gorm"

	"github.com/genefriendway/human-network-iam/conf"
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
	// var count int64
	// db.Model(&domain.Account{}).Where("role = ?", constants.Admin.String()).Count(&count)

	// if count == 0 {
	// 	adminEmail := config.AdminAccount.AdminEmail
	// 	adminPassword := config.AdminAccount.AdminPassword

	// 	if adminEmail == "" || adminPassword == "" {
	// 		return errors.New("missing ADMIN credentials in environment variables")
	// 	}

	// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to hash password: %w", err)
	// 	}
	// 	hashedPasswordString := string(hashedPassword)

	// 	admin := &domain.Account{
	// 		Email:        adminEmail,
	// 		Username:     "admin",
	// 		PasswordHash: &hashedPasswordString,
	// 		Role:         constants.Admin.String(),
	// 	}

	// 	if err := db.Create(admin).Error; err != nil {
	// 		return fmt.Errorf("failed to create ADMIN: %w", err)
	// 	}
	// 	fmt.Println("ADMIN account created successfully.")
	// }

	return nil
}

func RunMigrations(db *gorm.DB, config *conf.Configuration) error {
	log.Println("Running migrations...")

	sqlScripts := []string{
		filepath.Join(".", "migrations", "sql", "01_organization.sql"),
	}

	for _, script := range sqlScripts {
		if err := applySQLScript(db, script); err != nil {
			return err
		}
	}

	// Seed admin account
	if err := seedAdminAccount(db, config); err != nil {
		return err
	}

	log.Println("Migrations completed.")
	return nil
}
