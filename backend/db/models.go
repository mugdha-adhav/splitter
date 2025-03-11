package db

import (
	"time"
)

type User struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name"`
	Email        string    `json:"email" gorm:"unique"`
	PasswordHash string    `json:"-" gorm:"not null"` // Store hashed password, not exposed in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Expense struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title"`
	Amount      float64        `json:"amount"`
	PaidBy      uint           `json:"paid_by"`
	User        User           `json:"user" gorm:"foreignKey:PaidBy"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Splits      []ExpenseSplit `json:"splits,omitempty"`
}

type ExpenseSplit struct {
	ID         uint    `json:"id" gorm:"primaryKey"`
	ExpenseID  uint    `json:"expense_id"`
	Expense    Expense `json:"expense" gorm:"foreignKey:ExpenseID"`
	UserID     uint    `json:"user_id"`
	User       User    `json:"user" gorm:"foreignKey:UserID"`
	ShareRatio float64 `json:"share_ratio"`
	Amount     float64 `json:"amount"`
}

// AuthRequest represents the login/register request payload
type AuthRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name,omitempty"` // Only required for registration
}

// AuthResponse represents the login/register response
type AuthResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"` // JWT token for authentication
}
