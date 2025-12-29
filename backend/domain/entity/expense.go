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
	PaidBy      uuid.UUID // User ID who paid
	GroupID     uuid.UUID
	CreatedAt   time.Time
}

// NewExpense creates a new expense entity
func NewExpense(description string, amount float64, paidBy, groupID uuid.UUID) *Expense {
	return &Expense{
		ID:          uuid.New(),
		Description: description,
		Amount:      amount,
		PaidBy:      paidBy,
		GroupID:     groupID,
		CreatedAt:   time.Now(),
	}
}


