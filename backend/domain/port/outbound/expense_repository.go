package outbound

import "splitwise/domain/entity"

// ExpenseRepository defines the outbound port (driven port) for expense persistence
type ExpenseRepository interface {
	Create(expense *entity.Expense) error
	FindByID(id string) (*entity.Expense, error)
	FindByGroupID(groupID string) ([]*entity.Expense, error)
}


