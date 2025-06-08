package db

import (
	"fmt"
	"oneclickdevenv/backend/models"
)

func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	err := DB.AutoMigrate(
		&models.User{},
		&models.Instance{},
	)
	if err != nil {
		return fmt.Errorf("auto migration failed: %v", err)
	}

	fmt.Println("✅ Database migrated successfully")
	return nil
}
