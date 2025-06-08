package models

import (
	"time"

	"github.com/google/uuid"
)

type Instance struct {
	ID            uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID        uuid.UUID `gorm:"type:uuid;not null"`
	Name          string    `gorm:"not null"`
	Stack         string
	Zone          string
	MachineType   string
	Image         string
	DiskSize      int
	Status        string `gorm:"default:'provisioning'"`
	IP            string
	ExpiresAt     *time.Time
	RepoName      *string
	RepoURL       *string
	CloudInitUsed *bool
	LogsURL       *string
	CreatedAt     time.Time
	UpdatedAt     time.Time

	User User `gorm:"foreignKey:UserID"`
}
