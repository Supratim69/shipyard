package models

import (
	"time"

	"github.com/google/uuid"
)

type InstanceEvent struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	InstanceID uuid.UUID `gorm:"type:uuid;not null"`
	Type       string    `gorm:"not null"`
	Message    string
	Timestamp  time.Time

	Instance Instance `gorm:"foreignKey:InstanceID"`
}
