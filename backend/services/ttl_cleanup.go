package services

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"oneclickdevenv/backend/db"
	"oneclickdevenv/backend/models"
)

func StartTTLReaper() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			<-ticker.C
			now := time.Now()
			var expired []models.Instance
			if err := db.DB.Where("status = ? AND expires_at IS NOT NULL AND expires_at <= ?", "active", now).Find(&expired).Error; err != nil {
				log.Println("TTL reaper DB error:", err)
				continue
			}
			for _, inst := range expired {
				log.Printf("Auto-destroying expired instance %s (user %s)\n", inst.ID, inst.UserID)
				tfDir := filepath.Join("..", "infra", "terraform", inst.UserID.String()+"_"+inst.ID.String())
				if err := DestroyInfrastructure(tfDir); err != nil {
					log.Printf("Failed to destroy instance %s: %v\n", inst.ID, err)
					continue
				}
				db.DB.Model(&inst).Update("status", "destroyed")
				os.RemoveAll(tfDir)
			}
		}
	}()
}
