package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"splitwise/domain/entity"
	"splitwise/domain/port/outbound"
)

// PostgresExpenseSplitRepository implements ExpenseSplitRepository using PostgreSQL
type PostgresExpenseSplitRepository struct {
	db *sql.DB
}

// NewPostgresExpenseSplitRepository creates a new PostgreSQL expense split repository
func NewPostgresExpenseSplitRepository(db *sql.DB) outbound.ExpenseSplitRepository {
	return &PostgresExpenseSplitRepository{db: db}
}

// Create inserts a new expense split into the database
func (r *PostgresExpenseSplitRepository) Create(split *entity.ExpenseSplit) error {
	query := `
		INSERT INTO expense_splits (id, expense_id, user_id, amount, share_type, share_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	_, err := r.db.Exec(query,
		split.ID,
		split.ExpenseID,
		split.UserID,
		split.Amount,
		string(split.ShareType),
		split.ShareValue,
		split.CreatedAt,
		split.UpdatedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23503" { // Foreign key violation
				return fmt.Errorf("invalid expense or user reference: %w", err)
			}
		}
		return fmt.Errorf("failed to create expense split: %w", err)
	}
	return nil
}

// FindByExpenseID retrieves all splits for an expense
func (r *PostgresExpenseSplitRepository) FindByExpenseID(expenseID string) ([]*entity.ExpenseSplit, error) {
	expenseUUID, err := uuid.Parse(expenseID)
	if err != nil {
		return nil, fmt.Errorf("invalid expense ID: %w", err)
	}

	query := `
		SELECT id, expense_id, user_id, amount, share_type, share_value, created_at, updated_at
		FROM expense_splits
		WHERE expense_id = $1
		ORDER BY created_at ASC
	`
	
	rows, err := r.db.Query(query, expenseUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query expense splits: %w", err)
	}
	defer rows.Close()

	splits := make([]*entity.ExpenseSplit, 0)
	for rows.Next() {
		var split entity.ExpenseSplit
		var shareTypeStr string
		
		err := rows.Scan(
			&split.ID,
			&split.ExpenseID,
			&split.UserID,
			&split.Amount,
			&shareTypeStr,
			&split.ShareValue,
			&split.CreatedAt,
			&split.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense split: %w", err)
		}
		
		split.ShareType = entity.ShareType(shareTypeStr)
		splits = append(splits, &split)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expense splits: %w", err)
	}

	return splits, nil
}

// FindByUserID retrieves all splits for a user
func (r *PostgresExpenseSplitRepository) FindByUserID(userID string) ([]*entity.ExpenseSplit, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	query := `
		SELECT id, expense_id, user_id, amount, share_type, share_value, created_at, updated_at
		FROM expense_splits
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to query expense splits: %w", err)
	}
	defer rows.Close()

	splits := make([]*entity.ExpenseSplit, 0)
	for rows.Next() {
		var split entity.ExpenseSplit
		var shareTypeStr string
		
		err := rows.Scan(
			&split.ID,
			&split.ExpenseID,
			&split.UserID,
			&split.Amount,
			&shareTypeStr,
			&split.ShareValue,
			&split.CreatedAt,
			&split.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense split: %w", err)
		}
		
		split.ShareType = entity.ShareType(shareTypeStr)
		splits = append(splits, &split)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expense splits: %w", err)
	}

	return splits, nil
}

// FindByExpenseAndUser finds a specific split for an expense and user
func (r *PostgresExpenseSplitRepository) FindByExpenseAndUser(expenseID, userID string) (*entity.ExpenseSplit, error) {
	expenseUUID, err := uuid.Parse(expenseID)
	if err != nil {
		return nil, fmt.Errorf("invalid expense ID: %w", err)
	}
	
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	query := `
		SELECT id, expense_id, user_id, amount, share_type, share_value, created_at, updated_at
		FROM expense_splits
		WHERE expense_id = $1 AND user_id = $2
		LIMIT 1
	`
	
	var split entity.ExpenseSplit
	var shareTypeStr string
	
	err = r.db.QueryRow(query, expenseUUID, userUUID).Scan(
		&split.ID,
		&split.ExpenseID,
		&split.UserID,
		&split.Amount,
		&shareTypeStr,
		&split.ShareValue,
		&split.CreatedAt,
		&split.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("expense split not found")
		}
		return nil, fmt.Errorf("failed to find expense split: %w", err)
	}
	
	split.ShareType = entity.ShareType(shareTypeStr)
	return &split, nil
}

// DeleteByExpenseID deletes all splits for an expense
func (r *PostgresExpenseSplitRepository) DeleteByExpenseID(expenseID string) error {
	expenseUUID, err := uuid.Parse(expenseID)
	if err != nil {
		return fmt.Errorf("invalid expense ID: %w", err)
	}

	query := `DELETE FROM expense_splits WHERE expense_id = $1`
	_, err = r.db.Exec(query, expenseUUID)
	if err != nil {
		return fmt.Errorf("failed to delete expense splits: %w", err)
	}
	return nil
}

// Update updates an existing expense split
func (r *PostgresExpenseSplitRepository) Update(split *entity.ExpenseSplit) error {
	query := `
		UPDATE expense_splits
		SET amount = $1, share_type = $2, share_value = $3, updated_at = $4
		WHERE id = $5
	`
	
	result, err := r.db.Exec(query,
		split.Amount,
		string(split.ShareType),
		split.ShareValue,
		split.UpdatedAt,
		split.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update expense split: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("expense split not found")
	}

	return nil
}

