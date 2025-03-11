package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate all schemas
	if err := db.AutoMigrate(&User{}, &Expense{}, &ExpenseSplit{}); err != nil {
		return nil, err
	}

	return db, nil
}
