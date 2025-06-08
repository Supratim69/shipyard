package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	GitHubID  string    `gorm:"uniqueIndex;not null"`
	Name      string
	Email     string
	AvatarURL string
	CreatedAt time.Time
	UpdatedAt time.Time
	Instances []Instance `gorm:"foreignKey:UserID"`
}
