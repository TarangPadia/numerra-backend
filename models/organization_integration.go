package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IntegrationStatus string

const (
	IntegrationConnected    IntegrationStatus = "CONNECTED"
	IntegrationDisconnected IntegrationStatus = "DISCONNECTED"
	IntegrationNeedsReauth  IntegrationStatus = "NEEDS_REAUTH"
)

type OrganizationIntegration struct {
	ID                string            `gorm:"type:char(36);primaryKey"`
	OrganizationID    string            `gorm:"type:char(36);not null;index;uniqueIndex:uq_org_provider,priority:1"`
	Provider          string            `gorm:"type:varchar(32);not null;uniqueIndex:uq_org_provider,priority:2"`
	Status            IntegrationStatus `gorm:"type:enum('CONNECTED','DISCONNECTED','NEEDS_REAUTH');not null;default:'DISCONNECTED'"`
	ExternalAccountID *string           `gorm:"type:varchar(255);null"`
	Scopes            *string           `gorm:"type:text;null"`
	AccessTokenEnc    *string           `gorm:"type:text;null"`
	RefreshTokenEnc   *string           `gorm:"type:text;null"`
	AccessExpiresAt   *time.Time        `gorm:"null"`
	RefreshExpiresAt  *time.Time        `gorm:"null"`
	ConnectedByUserID *string           `gorm:"type:char(36);null"`
	ConnectedAt       *time.Time        `gorm:"null"`
	LastRefreshedAt   *time.Time        `gorm:"null"`
	UpdatedAt         time.Time         `gorm:"autoUpdateTime"`
}

func (m *OrganizationIntegration) TableName() string { return "organization_integrations" }

func (o *OrganizationIntegration) BeforeCreate(tx *gorm.DB) (err error) {
	o.ID = uuid.New().String()
	return nil
}
