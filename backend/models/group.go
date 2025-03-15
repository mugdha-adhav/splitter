package models

import "gorm.io/gorm"

type Group struct {
	gorm.Model

	Name       string    `gorm:"type:varchar(40);not null" json:"name" form:"name"`
	OwnerRefer uint      `gorm:"not null" json:"owner_id" form:"owner_id"`
	Owner      User      `gorm:"foreignKey:OwnerRefer;constraint:OnDelete:CASCADE;"`
	Members    []User    `gorm:"many2many:user_groups;"`
	Expenses   []Expense `gorm:"foreignKey:GroupID"`
}
