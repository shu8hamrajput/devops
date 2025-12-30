package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"splitwise/domain/entity"
	"splitwise/domain/port/outbound"
)

// PostgresExpenseRepository implements ExpenseRepository using PostgreSQL
type PostgresExpenseRepository struct {
	db *sql.DB
}

// NewPostgresExpenseRepository creates a new PostgreSQL expense repository
func NewPostgresExpenseRepository(db *sql.DB) outbound.ExpenseRepository {
	return &PostgresExpenseRepository{db: db}
}

// Create inserts a new expense into the database
func (r *PostgresExpenseRepository) Create(expense *entity.Expense) error {
	query := `
		INSERT INTO expenses (id, description, amount, paid_by, group_id, owed_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	
	var groupID interface{}
	if expense.GroupID != nil {
		groupID = expense.GroupID
	} else {
		groupID = nil
	}

	var owedBy interface{}
	if expense.OwedBy != nil {
		owedBy = expense.OwedBy
	} else {
		owedBy = nil
	}

	_, err := r.db.Exec(query,
		expense.ID,
		expense.Description,
		expense.Amount,
		expense.PaidBy,
		groupID,
		owedBy,
		expense.CreatedAt,
		expense.UpdatedAt,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23503" { // Foreign key violation
				return fmt.Errorf("invalid user or group reference: %w", err)
			}
		}
		return fmt.Errorf("failed to create expense: %w", err)
	}
	return nil
}

// FindByID retrieves an expense by ID
func (r *PostgresExpenseRepository) FindByID(id string) (*entity.Expense, error) {
	expenseID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid expense ID: %w", err)
	}

	expense := &entity.Expense{}
	var groupID sql.NullString
	var owedBy sql.NullString

	query := `
		SELECT id, description, amount, paid_by, group_id, owed_by, created_at, updated_at
		FROM expenses
		WHERE id = $1 AND deleted_at IS NULL
	`
	err = r.db.QueryRow(query, expenseID).Scan(
		&expense.ID,
		&expense.Description,
		&expense.Amount,
		&expense.PaidBy,
		&groupID,
		&owedBy,
		&expense.CreatedAt,
		&expense.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("expense not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	// Convert nullable fields
	if groupID.Valid {
		groupUUID, err := uuid.Parse(groupID.String)
		if err != nil {
			return nil, fmt.Errorf("invalid group ID in database: %w", err)
		}
		expense.GroupID = &groupUUID
	} else {
		expense.GroupID = nil
	}

	if owedBy.Valid {
		owedByUUID, err := uuid.Parse(owedBy.String)
		if err != nil {
			return nil, fmt.Errorf("invalid owed_by ID in database: %w", err)
		}
		expense.OwedBy = &owedByUUID
	} else {
		expense.OwedBy = nil
	}

	return expense, nil
}

// FindByGroupID retrieves all expenses for a group
func (r *PostgresExpenseRepository) FindByGroupID(groupID string) ([]*entity.Expense, error) {
	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %w", err)
	}

	query := `
		SELECT id, description, amount, paid_by, group_id, owed_by, created_at, updated_at
		FROM expenses
		WHERE group_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(query, groupUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}
	defer rows.Close()

	expenses := []*entity.Expense{}
	for rows.Next() {
		expense := &entity.Expense{}
		var groupID sql.NullString
		var owedBy sql.NullString

		err := rows.Scan(
			&expense.ID,
			&expense.Description,
			&expense.Amount,
			&expense.PaidBy,
			&groupID,
			&owedBy,
			&expense.CreatedAt,
			&expense.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}

		// Convert nullable fields
		if groupID.Valid {
			groupUUID, err := uuid.Parse(groupID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid group ID in database: %w", err)
			}
			expense.GroupID = &groupUUID
		} else {
			expense.GroupID = nil
		}

		if owedBy.Valid {
			owedByUUID, err := uuid.Parse(owedBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid owed_by ID in database: %w", err)
			}
			expense.OwedBy = &owedByUUID
		} else {
			expense.OwedBy = nil
		}

		expenses = append(expenses, expense)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expenses: %w", err)
	}

	return expenses, nil
}

// FindByUserID retrieves all expenses for a user (both paid and owed), sorted by UpdatedAt descending
func (r *PostgresExpenseRepository) FindByUserID(userID string) ([]*entity.Expense, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	query := `
		SELECT id, description, amount, paid_by, group_id, owed_by, created_at, updated_at
		FROM expenses
		WHERE (paid_by = $1 OR owed_by = $1) AND deleted_at IS NULL
		ORDER BY updated_at DESC
	`
	rows, err := r.db.Query(query, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}
	defer rows.Close()

	expenses := []*entity.Expense{}
	for rows.Next() {
		expense := &entity.Expense{}
		var groupID sql.NullString
		var owedBy sql.NullString

		err := rows.Scan(
			&expense.ID,
			&expense.Description,
			&expense.Amount,
			&expense.PaidBy,
			&groupID,
			&owedBy,
			&expense.CreatedAt,
			&expense.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}

		// Convert nullable fields
		if groupID.Valid {
			groupUUID, err := uuid.Parse(groupID.String)
			if err != nil {
				return nil, fmt.Errorf("invalid group ID in database: %w", err)
			}
			expense.GroupID = &groupUUID
		} else {
			expense.GroupID = nil
		}

		if owedBy.Valid {
			owedByUUID, err := uuid.Parse(owedBy.String)
			if err != nil {
				return nil, fmt.Errorf("invalid owed_by ID in database: %w", err)
			}
			expense.OwedBy = &owedByUUID
		} else {
			expense.OwedBy = nil
		}

		expenses = append(expenses, expense)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expenses: %w", err)
	}

	return expenses, nil
}

// Update updates an existing expense
func (r *PostgresExpenseRepository) Update(expense *entity.Expense) error {
	var groupID interface{}
	if expense.GroupID != nil {
		groupID = expense.GroupID
	} else {
		groupID = nil
	}

	var owedBy interface{}
	if expense.OwedBy != nil {
		owedBy = expense.OwedBy
	} else {
		owedBy = nil
	}

	query := `
		UPDATE expenses
		SET description = $1, amount = $2, paid_by = $3, group_id = $4, owed_by = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(query,
		expense.Description,
		expense.Amount,
		expense.PaidBy,
		groupID,
		owedBy,
		expense.UpdatedAt,
		expense.ID,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23503" { // Foreign key violation
				return fmt.Errorf("invalid user or group reference: %w", err)
			}
		}
		return fmt.Errorf("failed to update expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("expense not found")
	}

	return nil
}

// Delete performs a soft delete on an expense by setting deleted_at timestamp
func (r *PostgresExpenseRepository) Delete(id string) error {
	expenseID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid expense ID: %w", err)
	}

	query := `
		UPDATE expenses
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(query, expenseID)
	if err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("expense not found")
	}

	return nil
}
