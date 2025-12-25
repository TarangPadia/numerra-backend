package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	ROLE_OWNER     Role = "ROLE_OWNER"
	ROLE_ADMIN     Role = "ROLE_ADMIN"
	ROLE_EDITOR    Role = "ROLE_EDITOR"
	ROLE_SPECTATOR Role = "ROLE_SPECTATOR"
)

type OrganizationMember struct {
	MemberID       string    `gorm:"type:char(36);primaryKey"`
	UserID         string    `gorm:"type:char(36);not null"`
	OrganizationID string    `gorm:"type:char(36);not null"`
	Role           Role      `gorm:"type:enum('ROLE_OWNER','ROLE_ADMIN','ROLE_EDITOR','ROLE_SPECTATOR');not null;default:'ROLE_SPECTATOR'"`
	JoinedAt       time.Time `gorm:"not null"`

	User         User         `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Organization Organization `gorm:"foreignKey:OrganizationID;references:OrganizationID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

func (m *OrganizationMember) TableName() string { return "organization_members" }

func (m *OrganizationMember) BeforeCreate(tx *gorm.DB) (err error) {
	m.MemberID = uuid.New().String()
	if m.Role == "" {
		m.Role = ROLE_SPECTATOR
	}
	if m.JoinedAt.IsZero() {
		m.JoinedAt = time.Now()
	}
	return
}
