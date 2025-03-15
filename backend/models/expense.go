package models

import "gorm.io/gorm"

type Expense struct {
	gorm.Model

	Amount      float64 `gorm:"not null" json:"amount" form:"amount"`
	GroupID     uint    `gorm:"not null" json:"group_id" form:"group_id"`
	Group       Group   `gorm:"constraint:OnDelete:CASCADE;"`
	CreatedByID uint    `gorm:"not null" json:"created_by_id" form:"created_by_id"`
	CreatedBy   User    `gorm:"constraint:OnDelete:CASCADE;"`
	Users       []User  `gorm:"many2many:user_expenses;"`
}
