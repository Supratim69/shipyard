package routes

import (
	"net/http"
	"os"
	"path/filepath"

	"oneclickdevenv/backend/db"
	"oneclickdevenv/backend/models"
	"oneclickdevenv/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DestroyRequest struct {
	InstanceID string `json:"instance_id"`
}

func DestroyVM(c *gin.Context) {
	var req DestroyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	tokenUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	instanceUUID, err := uuid.Parse(req.InstanceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid instance_id"})
		return
	}

	var instance models.Instance
	if err := db.DB.First(&instance, "id = ?", instanceUUID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Instance not found"})
		return
	}

	if instance.UserID.String() != tokenUserID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
		return
	}

	tfDir := filepath.Join("..", "infra", "terraform", instance.UserID.String()+"_"+instance.ID.String())

	if err := services.DestroyInfrastructure(tfDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terraform destroy failed"})
		return
	}

	if err := db.DB.Delete(&instance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete instance from DB"})
		return
	}

	if err := os.RemoveAll(tfDir); err != nil {
		// Log error, but don't fail the API if destroy succeeded
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Destruction successful",
		"instance_id": instance.ID.String(),
	})
}
