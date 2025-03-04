package db

import (
	"database/sql"
	"errors"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// User operations
func (r *Repository) CreateUser(user *User) error {
	query := `INSERT INTO users (id, name, email) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, user.ID, user.Name, user.Email)
	return err
}

func (r *Repository) GetUser(id string) (*User, error) {
	var user User
	query := `SELECT id, name, email, created_at FROM users WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return &user, err
}

func (r *Repository) ListUsers() ([]User, error) {
	query := `SELECT id, name, email, created_at FROM users`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// Expense operations
func (r *Repository) CreateExpense(expense *Expense) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert expense
	query := `INSERT INTO expenses (id, description, amount, paid_by) VALUES (?, ?, ?, ?)`
	_, err = tx.Exec(query, expense.ID, expense.Description, expense.Amount, expense.PaidBy)
	if err != nil {
		return err
	}

	// Insert splits
	query = `INSERT INTO expense_splits (expense_id, user_id, share) VALUES (?, ?, ?)`
	for _, split := range expense.SplitAmong {
		_, err = tx.Exec(query, expense.ID, split.UserID, split.Share)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) GetExpense(id string) (*Expense, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var expense Expense
	query := `SELECT id, description, amount, paid_by, created_at FROM expenses WHERE id = ?`
	err = tx.QueryRow(query, id).Scan(&expense.ID, &expense.Description, &expense.Amount, &expense.PaidBy, &expense.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("expense not found")
	}
	if err != nil {
		return nil, err
	}

	// Get splits
	query = `SELECT user_id, share FROM expense_splits WHERE expense_id = ?`
	rows, err := tx.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var split ExpenseSplit
		if err := rows.Scan(&split.UserID, &split.Share); err != nil {
			return nil, err
		}
		expense.SplitAmong = append(expense.SplitAmong, split)
	}

	return &expense, tx.Commit()
}

func (r *Repository) ListExpenses() ([]Expense, error) {
	query := `SELECT id, description, amount, paid_by, created_at FROM expenses`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var expense Expense
		if err := rows.Scan(&expense.ID, &expense.Description, &expense.Amount, &expense.PaidBy, &expense.CreatedAt); err != nil {
			return nil, err
		}

		// Get splits for each expense
		splits, err := r.getExpenseSplits(expense.ID)
		if err != nil {
			return nil, err
		}
		expense.SplitAmong = splits
		expenses = append(expenses, expense)
	}
	return expenses, nil
}

func (r *Repository) GetUserExpenses(userID string) ([]Expense, error) {
	query := `
		SELECT DISTINCT e.id, e.description, e.amount, e.paid_by, e.created_at 
		FROM expenses e
		LEFT JOIN expense_splits es ON e.id = es.expense_id
		WHERE e.paid_by = ? OR es.user_id = ?`

	rows, err := r.db.Query(query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []Expense
	for rows.Next() {
		var expense Expense
		if err := rows.Scan(&expense.ID, &expense.Description, &expense.Amount, &expense.PaidBy, &expense.CreatedAt); err != nil {
			return nil, err
		}

		splits, err := r.getExpenseSplits(expense.ID)
		if err != nil {
			return nil, err
		}
		expense.SplitAmong = splits
		expenses = append(expenses, expense)
	}
	return expenses, nil
}

func (r *Repository) getExpenseSplits(expenseID string) ([]ExpenseSplit, error) {
	query := `SELECT user_id, share FROM expense_splits WHERE expense_id = ?`
	rows, err := r.db.Query(query, expenseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var splits []ExpenseSplit
	for rows.Next() {
		var split ExpenseSplit
		if err := rows.Scan(&split.UserID, &split.Share); err != nil {
			return nil, err
		}
		splits = append(splits, split)
	}
	return splits, nil
}
