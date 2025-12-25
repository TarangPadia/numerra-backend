package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Organization struct {
	OrganizationID     string  `gorm:"type:char(36);primaryKey"`
	OrganizationName   string  `gorm:"type:varchar(255)"`
	IncorporationState string  `gorm:"type:varchar(255)"`
	IncorporationYear  int     `gorm:"type:int"`
	Industry           string  `gorm:"type:varchar(255)"`
	Revenue            float64 `gorm:"type:decimal(15,2)"`
}

func (m *Organization) TableName() string { return "organizations" }

func (o *Organization) BeforeCreate(tx *gorm.DB) (err error) {
	o.OrganizationID = uuid.New().String()
	return
}
