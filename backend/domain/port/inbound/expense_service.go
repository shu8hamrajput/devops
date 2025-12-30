package inbound

import "splitwise/domain/entity"

// ExpenseService defines the inbound port (driving port) for expense operations
type ExpenseService interface {
	CreateExpense(description string, amount float64, paidBy, groupID string) (*entity.Expense, error)
	CreateExpenseWithSplit(description string, amount float64, paidBy, groupID string, shareType entity.ShareType, userShares map[string]float64) (*entity.Expense, error)
	CreateUserToUserExpense(description string, amount float64, paidBy, owedBy string) (*entity.Expense, error)
	GetExpenseByID(id string) (*entity.Expense, error)
	GetExpensesByGroup(groupID string) ([]*entity.Expense, error)
	GetExpensesByUser(userID string) ([]*entity.Expense, error)
	UpdateExpense(id, description string, amount float64, paidBy, groupID, owedBy string) (*entity.Expense, error)
	DeleteExpense(id string) error
}


