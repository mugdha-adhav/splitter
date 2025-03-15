package models

import "gorm.io/gorm"

type User struct {
	gorm.Model

	Name     string `gorm:"type:varchar(40);unique;not null" json:"name,omitempty" form:"name,omitempty"`
	Password string `gorm:"size:255;not null" json:"password,omitempty" form:"password,omitempty"`
	Email    string `gorm:"type:varchar(40);unique;not null" json:"email" form:"email,omitempty"`
	// Add relationships
	OwnedGroups     []Group   `gorm:"foreignKey:OwnerRefer"`
	Groups          []Group   `gorm:"many2many:user_groups;"`
	CreatedExpenses []Expense `gorm:"foreignKey:CreatedByID"`
	SharedExpenses  []Expense `gorm:"many2many:user_expenses;"`
}
