package inbound

import "splitwise/domain/entity"

// ExpenseService defines the inbound port (driving port) for expense operations
type ExpenseService interface {
	CreateExpense(description string, amount float64, paidBy, groupID string) (*entity.Expense, error)
	GetExpenseByID(id string) (*entity.Expense, error)
	GetExpensesByGroup(groupID string) ([]*entity.Expense, error)
}


