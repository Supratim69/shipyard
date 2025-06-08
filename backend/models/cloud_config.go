package models

import (
	"time"

	"github.com/google/uuid"
)

type CloudConfig struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	InstanceID uuid.UUID `gorm:"type:uuid;not null"`
	Template   string    `gorm:"type:text;not null"`
	InjectedAt time.Time

	Instance Instance `gorm:"foreignKey:InstanceID"`
}
