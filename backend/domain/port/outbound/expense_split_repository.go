package outbound

import "splitwise/domain/entity"

// ExpenseSplitRepository defines the interface for expense split persistence
type ExpenseSplitRepository interface {
	// Create creates a new expense split
	Create(split *entity.ExpenseSplit) error
	
	// FindByExpenseID finds all splits for an expense
	FindByExpenseID(expenseID string) ([]*entity.ExpenseSplit, error)
	
	// FindByUserID finds all splits for a user
	FindByUserID(userID string) ([]*entity.ExpenseSplit, error)
	
	// FindByExpenseAndUser finds a specific split for an expense and user
	FindByExpenseAndUser(expenseID, userID string) (*entity.ExpenseSplit, error)
	
	// DeleteByExpenseID deletes all splits for an expense
	DeleteByExpenseID(expenseID string) error
	
	// Update updates an existing expense split
	Update(split *entity.ExpenseSplit) error
}

