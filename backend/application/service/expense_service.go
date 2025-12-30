package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"splitwise/domain/entity"
	"splitwise/domain/port/inbound"
	"splitwise/domain/port/outbound"
)

// expenseService implements the ExpenseService port
type expenseService struct {
	expenseRepo  outbound.ExpenseRepository
	userRepo     outbound.UserRepository
	groupRepo    outbound.GroupRepository
	splitService *expenseSplitService
}

// NewExpenseService creates a new expense service
func NewExpenseService(
	expenseRepo outbound.ExpenseRepository,
	userRepo outbound.UserRepository,
	groupRepo outbound.GroupRepository,
	splitRepo outbound.ExpenseSplitRepository,
) inbound.ExpenseService {
	return &expenseService{
		expenseRepo:  expenseRepo,
		userRepo:     userRepo,
		groupRepo:    groupRepo,
		splitService: NewExpenseSplitService(splitRepo, userRepo, groupRepo),
	}
}

// CreateExpense creates a new expense in a group
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
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	// Update group timestamp
	group.UpdateTimestamp()
	s.groupRepo.Update(group)

	expense := entity.NewExpense(description, amount, paidByUUID, groupUUID)
	if err := s.expenseRepo.Create(expense); err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	// Auto-split expense equally among group members by default
	_, err = s.splitService.SplitExpense(expense, entity.ShareTypeEqual, nil)
	if err != nil {
		// If splitting fails, delete the expense to maintain consistency
		s.expenseRepo.Delete(expense.ID.String())
		return nil, fmt.Errorf("failed to split expense: %w", err)
	}

	return expense, nil
}

// CreateExpenseWithSplit creates a new expense with custom split configuration
func (s *expenseService) CreateExpenseWithSplit(description string, amount float64, paidBy, groupID string, shareType entity.ShareType, userShares map[string]float64) (*entity.Expense, error) {
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
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	// Update group timestamp
	group.UpdateTimestamp()
	s.groupRepo.Update(group)

	expense := entity.NewExpense(description, amount, paidByUUID, groupUUID)
	if err := s.expenseRepo.Create(expense); err != nil {
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	// Split expense according to the provided configuration
	_, err = s.splitService.SplitExpense(expense, shareType, userShares)
	if err != nil {
		// If splitting fails, delete the expense to maintain consistency
		s.expenseRepo.Delete(expense.ID.String())
		return nil, fmt.Errorf("failed to split expense: %w", err)
	}

	return expense, nil
}

// CreateUserToUserExpense creates a new expense between two users
func (s *expenseService) CreateUserToUserExpense(description string, amount float64, paidBy, owedBy string) (*entity.Expense, error) {
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

	owedByUUID, err := uuid.Parse(owedBy)
	if err != nil {
		return nil, fmt.Errorf("invalid owedBy user ID: %w", err)
	}

	if paidBy == owedBy {
		return nil, fmt.Errorf("paidBy and owedBy cannot be the same user")
	}

	// Validate that both users exist
	_, err = s.userRepo.FindByID(paidBy)
	if err != nil {
		return nil, fmt.Errorf("paidBy user not found: %w", err)
	}

	_, err = s.userRepo.FindByID(owedBy)
	if err != nil {
		return nil, fmt.Errorf("owedBy user not found: %w", err)
	}

	expense := entity.NewUserToUserExpense(description, amount, paidByUUID, owedByUUID)
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

// GetExpensesByUser retrieves all expenses for a user (both paid and owed)
func (s *expenseService) GetExpensesByUser(userID string) ([]*entity.Expense, error) {
	expenses, err := s.expenseRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}
	return expenses, nil
}

// UpdateExpense updates an existing expense
func (s *expenseService) UpdateExpense(id, description string, amount float64, paidBy, groupID, owedBy string) (*entity.Expense, error) {
	if description == "" {
		return nil, fmt.Errorf("description cannot be empty")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}

	// Get existing expense
	expense, err := s.expenseRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("expense not found: %w", err)
	}

	// Validate paidBy user exists
	paidByUUID, err := uuid.Parse(paidBy)
	if err != nil {
		return nil, fmt.Errorf("invalid paidBy user ID: %w", err)
	}
	_, err = s.userRepo.FindByID(paidBy)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update fields
	expense.Description = description
	expense.Amount = amount
	expense.PaidBy = paidByUUID
	expense.UpdatedAt = time.Now()

	// Handle group_id and owed_by (mutually exclusive)
	if groupID != "" {
		groupUUID, err := uuid.Parse(groupID)
		if err != nil {
			return nil, fmt.Errorf("invalid group ID: %w", err)
		}
		// Validate group exists
		_, err = s.groupRepo.FindByID(groupID)
		if err != nil {
			return nil, fmt.Errorf("group not found: %w", err)
		}
		expense.GroupID = &groupUUID
		expense.OwedBy = nil
	} else if owedBy != "" {
		owedByUUID, err := uuid.Parse(owedBy)
		if err != nil {
			return nil, fmt.Errorf("invalid owedBy user ID: %w", err)
		}
		if paidBy == owedBy {
			return nil, fmt.Errorf("paidBy and owedBy cannot be the same user")
		}
		// Validate owedBy user exists
		_, err = s.userRepo.FindByID(owedBy)
		if err != nil {
			return nil, fmt.Errorf("owedBy user not found: %w", err)
		}
		expense.OwedBy = &owedByUUID
		expense.GroupID = nil
	} else {
		return nil, fmt.Errorf("either group_id or owed_by must be provided")
	}

	// Update in repository
	if err := s.expenseRepo.Update(expense); err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	// Update group timestamp if it's a group expense
	if expense.GroupID != nil {
		group, err := s.groupRepo.FindByID(expense.GroupID.String())
		if err == nil {
			group.UpdateTimestamp()
			s.groupRepo.Update(group)
		}
	}

	return expense, nil
}

// DeleteExpense deletes an expense (soft delete)
func (s *expenseService) DeleteExpense(id string) error {
	// Verify expense exists
	expense, err := s.expenseRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("expense not found: %w", err)
	}

	// Delete expense
	if err := s.expenseRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete expense: %w", err)
	}

	// Update group timestamp if it's a group expense
	if expense.GroupID != nil {
		group, err := s.groupRepo.FindByID(expense.GroupID.String())
		if err == nil {
			group.UpdateTimestamp()
			s.groupRepo.Update(group)
		}
	}

	return nil
}


