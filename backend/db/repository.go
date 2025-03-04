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
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO users (id, name, email) VALUES (?, ?, ?)`
	_, err = tx.Exec(query, user.ID, user.Name, user.Email)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) GetUser(id string) (*User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var user User
	query := `SELECT id, name, email, created_at FROM users WHERE id = ?`
	err = tx.QueryRow(query, id).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	return &user, tx.Commit()
}

func (r *Repository) ListUsers() ([]User, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `SELECT id, name, email, created_at FROM users`
	rows, err := tx.Query(query)
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

	return users, tx.Commit()
}

func (r *Repository) UpdateUser(user *User) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `UPDATE users SET name = ?, email = ? WHERE id = ?`
	result, err := tx.Exec(query, user.Name, user.Email, user.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return tx.Commit()
}

func (r *Repository) DeleteUser(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First check if user has any expenses or is part of any expense splits
	var count int
	query := `
		SELECT COUNT(*) FROM (
			SELECT paid_by FROM expenses WHERE paid_by = ?
			UNION
			SELECT user_id FROM expense_splits WHERE user_id = ?
		) AS user_expenses`
	err = tx.QueryRow(query, id, id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot delete user with associated expenses")
	}

	// Delete the user if no expenses are found
	query = `DELETE FROM users WHERE id = ?`
	result, err := tx.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return tx.Commit()
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
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `SELECT id, description, amount, paid_by, created_at FROM expenses`
	rows, err := tx.Query(query)
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
	return expenses, tx.Commit()
}

func (r *Repository) GetUserExpenses(userID string) ([]Expense, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `
		SELECT DISTINCT e.id, e.description, e.amount, e.paid_by, e.created_at 
		FROM expenses e
		LEFT JOIN expense_splits es ON e.id = es.expense_id
		WHERE e.paid_by = ? OR es.user_id = ?`

	rows, err := tx.Query(query, userID, userID)
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
	return expenses, tx.Commit()
}

func (r *Repository) getExpenseSplits(expenseID string) ([]ExpenseSplit, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	query := `SELECT user_id, share FROM expense_splits WHERE expense_id = ?`
	rows, err := tx.Query(query, expenseID)
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
	return splits, tx.Commit()
}

func (r *Repository) UpdateExpense(expense *Expense) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update expense details
	query := `UPDATE expenses SET description = ?, amount = ?, paid_by = ? WHERE id = ?`
	result, err := tx.Exec(query, expense.Description, expense.Amount, expense.PaidBy, expense.ID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("expense not found")
	}

	// Delete existing splits
	query = `DELETE FROM expense_splits WHERE expense_id = ?`
	_, err = tx.Exec(query, expense.ID)
	if err != nil {
		return err
	}

	// Insert new splits
	query = `INSERT INTO expense_splits (expense_id, user_id, share) VALUES (?, ?, ?)`
	for _, split := range expense.SplitAmong {
		_, err = tx.Exec(query, expense.ID, split.UserID, split.Share)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) DeleteExpense(id string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete expense splits first (due to foreign key constraint)
	query := `DELETE FROM expense_splits WHERE expense_id = ?`
	_, err = tx.Exec(query, id)
	if err != nil {
		return err
	}

	// Delete the expense
	query = `DELETE FROM expenses WHERE id = ?`
	result, err := tx.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("expense not found")
	}

	return tx.Commit()
}
