package models

import (
	"gorm.io/gorm"
)

type SessionMetadata struct {
	SessionID     uint    `gorm:"primaryKey;autoIncrement"`
	UserEmail     string  `gorm:"type:varchar(255);not null;uniqueIndex"`
	SelectedOrgID *string `gorm:"type:char(36);null"`
}

func (m *SessionMetadata) TableName() string { return "session_metadata" }

func (s *SessionMetadata) BeforeCreate(tx *gorm.DB) (err error) {
	return nil
}
