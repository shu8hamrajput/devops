package repository

import (
	"fmt"
	"sync"

	"splitwise/domain/entity"
	"splitwise/domain/port/outbound"
)

// memoryExpenseRepository is an in-memory implementation of ExpenseRepository
type memoryExpenseRepository struct {
	expenses map[string]*entity.Expense
	mu       sync.RWMutex
}

// NewMemoryExpenseRepository creates a new in-memory expense repository
func NewMemoryExpenseRepository() outbound.ExpenseRepository {
	return &memoryExpenseRepository{
		expenses: make(map[string]*entity.Expense),
	}
}

// Create stores a new expense
func (r *memoryExpenseRepository) Create(expense *entity.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.expenses[expense.ID.String()] = expense
	return nil
}

// FindByID retrieves an expense by ID
func (r *memoryExpenseRepository) FindByID(id string) (*entity.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	expense, exists := r.expenses[id]
	if !exists {
		return nil, fmt.Errorf("expense not found: %s", id)
	}

	return expense, nil
}

// FindByGroupID retrieves all expenses for a group
func (r *memoryExpenseRepository) FindByGroupID(groupID string) ([]*entity.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	expenses := make([]*entity.Expense, 0)
	for _, expense := range r.expenses {
		if expense.GroupID.String() == groupID {
			expenses = append(expenses, expense)
		}
	}

	return expenses, nil
}

