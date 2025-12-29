package service

import (
	"fmt"

	"github.com/google/uuid"
	"splitwise/domain/entity"
	"splitwise/domain/port/inbound"
	"splitwise/domain/port/outbound"
)

// expenseService implements the ExpenseService port
type expenseService struct {
	expenseRepo outbound.ExpenseRepository
	userRepo    outbound.UserRepository
	groupRepo   outbound.GroupRepository
}

// NewExpenseService creates a new expense service
func NewExpenseService(
	expenseRepo outbound.ExpenseRepository,
	userRepo outbound.UserRepository,
	groupRepo outbound.GroupRepository,
) inbound.ExpenseService {
	return &expenseService{
		expenseRepo: expenseRepo,
		userRepo:    userRepo,
		groupRepo:   groupRepo,
	}
}

// CreateExpense creates a new expense
func (s *expenseService) CreateExpense(description string, amount float64, paidBy, groupID string) (*entity.Expense, error) {
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}

	paidByUUID, err := uuid.Parse(paidBy)
	if err != nil {
		return nil, fmt.Errorf("invalid paidBy user ID: %w", err)
	}

	groupUUID, err := uuid.Parse(groupID)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %w", err)
	}

	// Validate that user exists
	_, err = s.userRepo.FindByID(paidBy)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Validate that group exists
	_, err = s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	expense := entity.NewExpense(description, amount, paidByUUID, groupUUID)
	if err := s.expenseRepo.Create(expense); err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	return expense, nil
}

// GetExpenseByID retrieves an expense by ID
func (s *expenseService) GetExpenseByID(id string) (*entity.Expense, error) {
	expense, err := s.expenseRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}
	return expense, nil
}

// GetExpensesByGroup retrieves all expenses for a group
func (s *expenseService) GetExpensesByGroup(groupID string) ([]*entity.Expense, error) {
	expenses, err := s.expenseRepo.FindByGroupID(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}
	return expenses, nil
}


