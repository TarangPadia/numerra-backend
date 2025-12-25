package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OAuthState struct {
	ID              string    `gorm:"type:char(36);primaryKey"`
	OrganizationID  string    `gorm:"type:char(36);not null"`
	Provider        string    `gorm:"type:varchar(32);not null"`
	State           string    `gorm:"type:varchar(128);not null;index:idx_oauth_states_state_provider,priority:1"`
	CreatedByUserID string    `gorm:"type:char(36);not null"`
	ExpiresAt       time.Time `gorm:"not null"`
	CreatedAt       time.Time `gorm:"not null;autoCreateTime"`
}

func (m *OAuthState) TableName() string { return "oauth_states" }

func (o *OAuthState) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = uuid.New().String()
	if o.CreatedAt.IsZero() {
		o.CreatedAt = time.Now()
	}
	return nil
}
