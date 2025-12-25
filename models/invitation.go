package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	StatusPending    = "PENDING"
	StatusIncomplete = "INCOMPLETE"
	StatusConfirmed  = "CONFIRMED"
)

type Invitation struct {
	ID             string `gorm:"type:char(36);primaryKey"`
	UserEmail      string `gorm:"type:varchar(255);not null"`
	OrganizationID string `gorm:"type:char(36);not null"`
	InviterUserID  string `gorm:"type:char(36);not null"`
	ExpiresAt      time.Time
	Role           Role   `gorm:"type:enum('ROLE_OWNER','ROLE_ADMIN','ROLE_EDITOR','ROLE_SPECTATOR')"`
	Status         string `gorm:"type:varchar(32);default:'PENDING'"`
}

func (m *Invitation) TableName() string { return "invitations" }

func (i *Invitation) BeforeCreate(tx *gorm.DB) (err error) {
	i.ID = uuid.New().String()
	return
}
