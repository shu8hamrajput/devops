package repository

import (
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
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
		if expense.GroupID != nil && expense.GroupID.String() == groupID {
			expenses = append(expenses, expense)
		}
	}

	return expenses, nil
}

// FindByUserID retrieves all expenses for a user (both paid and owed), sorted by UpdatedAt descending
func (r *memoryExpenseRepository) FindByUserID(userID string) ([]*entity.Expense, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	expenses := make([]*entity.Expense, 0)
	for _, expense := range r.expenses {
		// Include expenses where user paid or owes
		if expense.PaidBy == userUUID || (expense.OwedBy != nil && *expense.OwedBy == userUUID) {
			expenses = append(expenses, expense)
		}
	}

	// Sort by UpdatedAt descending (most recently modified first)
	sort.Slice(expenses, func(i, j int) bool {
		return expenses[i].UpdatedAt.After(expenses[j].UpdatedAt)
	})

	return expenses, nil
}

// Update updates an existing expense
func (r *memoryExpenseRepository) Update(expense *entity.Expense) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.expenses[expense.ID.String()]; !exists {
		return fmt.Errorf("expense not found: %s", expense.ID.String())
	}

	r.expenses[expense.ID.String()] = expense
	return nil
}

// Delete performs a soft delete on an expense (removes from map)
func (r *memoryExpenseRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.expenses[id]; !exists {
		return fmt.Errorf("expense not found: %s", id)
	}

	delete(r.expenses, id)
	return nil
}

