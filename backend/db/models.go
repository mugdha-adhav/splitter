package db

import "time"

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type Expense struct {
	ID          string         `json:"id"`
	Description string         `json:"description"`
	Amount      float64        `json:"amount"`
	PaidBy      string         `json:"paid_by"`
	SplitAmong  []ExpenseSplit `json:"split_among"`
	CreatedAt   time.Time      `json:"created_at"`
}

type ExpenseSplit struct {
	UserID string  `json:"user_id"`
	Share  float64 `json:"share"`
}
