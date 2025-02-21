package postgres

import (
	"fmt"
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

// func seedAdminAccount(db *gorm.DB, config *conf.Configuration) error {
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

// 	return nil
// }

func RunMigrations(db *gorm.DB, basePath string) error {
	log.Println("Running migrations...")

	// List of migration SQL files
	scriptFiles := []string{
		"01_identity_user.sql",
		"01_identity_organization.sql",
	}

	// Iterate over scripts and execute each
	for _, script := range scriptFiles {
		scriptPath := filepath.Join(basePath, script)

		log.Printf("Applying migration: %s\n", scriptPath)
		if err := applySQLScript(db, scriptPath); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", scriptPath, err)
		}
	}

	log.Println("Migrations completed successfully.")
	return nil
}
