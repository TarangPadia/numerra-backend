package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID                string `gorm:"type:char(36);primaryKey"`
	Email             string `gorm:"type:varchar(32);uniqueIndex"`
	FirstName         string `gorm:"type:varchar(32)"`
	LastName          string `gorm:"type:varchar(32)"`
	ShowWelcomePrompt bool   `gorm:"default:true"`
	IsEmailVerified   bool   `gorm:"default:false"`
}

func (m *User) TableName() string { return "users" }

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}
