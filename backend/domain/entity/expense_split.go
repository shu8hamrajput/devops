package entity

import (
	"time"

	"github.com/google/uuid"
)

// ShareType represents the type of expense split
type ShareType string

const (
	ShareTypeEqual        ShareType = "EQUAL"         // Split equally among all users
	ShareTypePercentage   ShareType = "PERCENTAGE"    // Split by percentage
	ShareTypeExactAmount  ShareType = "EXACT_AMOUNT"  // Split by exact amounts
	ShareTypeShares       ShareType = "SHARES"        // Split by shares (e.g., 2:1:1)
)

// ExpenseSplit represents how an expense is split among users
type ExpenseSplit struct {
	ID          uuid.UUID
	ExpenseID   uuid.UUID
	UserID      uuid.UUID
	Amount      float64  // The amount this user owes
	ShareType   ShareType
	ShareValue  float64  // Percentage, exact amount, or share count depending on ShareType
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewExpenseSplit creates a new expense split
func NewExpenseSplit(expenseID, userID uuid.UUID, amount float64, shareType ShareType, shareValue float64) *ExpenseSplit {
	now := time.Now()
	return &ExpenseSplit{
		ID:         uuid.New(),
		ExpenseID:  expenseID,
		UserID:     userID,
		Amount:     amount,
		ShareType:  shareType,
		ShareValue: shareValue,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// UpdateTimestamp updates the UpdatedAt timestamp
func (es *ExpenseSplit) UpdateTimestamp() {
	es.UpdatedAt = time.Now()
}
