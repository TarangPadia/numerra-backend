package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordResetCode struct {
	ID        string    `gorm:"type:char(36);primaryKey"`
	Email     string    `gorm:"type:varchar(255);not null"`
	Code      string    `gorm:"type:text;not null"`
	ExpiresAt time.Time `gorm:"not null"`
}

func (m *PasswordResetCode) TableName() string { return "password_reset_codes" }

func (p *PasswordResetCode) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
