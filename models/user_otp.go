package models

import (
	"time"

	"gorm.io/gorm"
)

type UserOTP struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Email     string `gorm:"type:varchar(255);not null"`
	OTP       string `gorm:"type:varchar(10);not null"`
	ExpiresAt time.Time
}

func (m *UserOTP) TableName() string { return "user_otps" }

func (u *UserOTP) BeforeCreate(tx *gorm.DB) (err error) {
	return
}
