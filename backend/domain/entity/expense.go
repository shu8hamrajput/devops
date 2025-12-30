package entity

import (
	"time"

	"github.com/google/uuid"
)

// Expense represents an expense in the system
type Expense struct {
	ID          uuid.UUID
	Description string
	Amount      float64
	PaidBy      uuid.UUID  // User ID who paid
	GroupID     *uuid.UUID // Optional: nil for user-to-user expenses
	OwedBy      *uuid.UUID // Optional: User ID who owes (for user-to-user expenses)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewExpense creates a new expense entity for a group
func NewExpense(description string, amount float64, paidBy, groupID uuid.UUID) *Expense {
	now := time.Now()
	return &Expense{
		ID:          uuid.New(),
		Description: description,
		Amount:      amount,
		PaidBy:      paidBy,
		GroupID:     &groupID,
		OwedBy:      nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewUserToUserExpense creates a new expense between two users
func NewUserToUserExpense(description string, amount float64, paidBy, owedBy uuid.UUID) *Expense {
	now := time.Now()
	return &Expense{
		ID:          uuid.New(),
		Description: description,
		Amount:      amount,
		PaidBy:      paidBy,
		GroupID:     nil,
		OwedBy:      &owedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}


