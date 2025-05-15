package postgres

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	// Read directory contents
	files, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	// Collect and sort .sql files
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}
	sort.Strings(sqlFiles) // ensure consistent order

	// Apply each SQL script
	for _, file := range sqlFiles {
		scriptPath := filepath.Join(basePath, file)
		log.Printf("Applying migration: %s\n", scriptPath)
		if err := applySQLScript(db, scriptPath); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", scriptPath, err)
		}
	}

	log.Println("Migrations completed successfully.")
	return nil
}
