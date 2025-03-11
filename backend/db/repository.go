package db

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Authentication-related functions
func (r *Repository) GetUserByEmail(email string) (*User, error) {
	var user User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *Repository) CreateUserWithAuth(user *User) error {
	return r.db.Create(user).Error
}

// User operations
func (r *Repository) GetUser(id uint) (*User, error) {
	var user User
	result := r.db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *Repository) ListUsers() ([]User, error) {
	var users []User
	result := r.db.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

func (r *Repository) UpdateUser(user *User) error {
	user.UpdatedAt = time.Now()
	result := r.db.Save(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *Repository) DeleteUser(id uint) error {
	// Check if user has any expenses or is part of any splits
	var count int64
	if err := r.db.Raw(`
		SELECT COUNT(*) FROM (
			SELECT paid_by FROM expenses WHERE paid_by = ?
			UNION
			SELECT user_id FROM expense_splits WHERE user_id = ?
		)`, id, id).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return errors.New("cannot delete user with associated expenses")
	}

	result := r.db.Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// Expense operations
func (r *Repository) ListExpenses() ([]Expense, error) {
	var expenses []Expense
	result := r.db.Preload("User").Preload("Splits").Find(&expenses)
	if result.Error != nil {
		return nil, result.Error
	}
	return expenses, nil
}

func (r *Repository) CreateExpense(expense *Expense) error {
	return r.db.Create(expense).Error
}

func (r *Repository) GetExpense(id uint) (*Expense, error) {
	var expense Expense
	result := r.db.Preload("User").Preload("Splits").First(&expense, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("expense not found")
		}
		return nil, result.Error
	}
	return &expense, nil
}

func (r *Repository) UpdateExpense(expense *Expense) error {
	expense.UpdatedAt = time.Now()

	// Start a transaction
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update the expense
		if err := tx.Save(expense).Error; err != nil {
			return err
		}

		// Delete existing splits
		if err := tx.Where("expense_id = ?", expense.ID).Delete(&ExpenseSplit{}).Error; err != nil {
			return err
		}

		// Create new splits
		for i := range expense.Splits {
			expense.Splits[i].ExpenseID = expense.ID
			if err := tx.Create(&expense.Splits[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Repository) DeleteExpense(id uint) error {
	result := r.db.Delete(&Expense{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("expense not found")
	}
	return nil
}

func (r *Repository) GetUserExpenses(userID uint) ([]Expense, error) {
	var expenses []Expense
	result := r.db.Preload("User").Preload("Splits").
		Where("paid_by = ?", userID).
		Or("id IN (?)", r.db.Table("expense_splits").Select("expense_id").Where("user_id = ?", userID)).
		Find(&expenses)
	if result.Error != nil {
		return nil, result.Error
	}
	return expenses, nil
}
