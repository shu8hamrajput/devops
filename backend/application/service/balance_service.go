package service

import (
	"fmt"
	"math"

	"splitwise/domain/port/inbound"
	"splitwise/domain/port/outbound"
)

// balanceService implements the BalanceService port
type balanceService struct {
	expenseRepo     outbound.ExpenseRepository
	expenseSplitRepo outbound.ExpenseSplitRepository
	userRepo        outbound.UserRepository
	groupRepo       outbound.GroupRepository
}

// NewBalanceService creates a new balance service
func NewBalanceService(
	expenseRepo outbound.ExpenseRepository,
	expenseSplitRepo outbound.ExpenseSplitRepository,
	userRepo outbound.UserRepository,
	groupRepo outbound.GroupRepository,
) inbound.BalanceService {
	return &balanceService{
		expenseRepo:      expenseRepo,
		expenseSplitRepo: expenseSplitRepo,
		userRepo:         userRepo,
		groupRepo:        groupRepo,
	}
}

// CalculateGroupBalance calculates who owes whom in a group
func (s *balanceService) CalculateGroupBalance(groupID string) (*inbound.GroupBalance, error) {
	// Validate group exists
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	// Get all expenses for the group
	expenses, err := s.expenseRepo.FindByGroupID(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	// Calculate balances: map[fromUserID][toUserID] = amount
	balances := make(map[string]map[string]float64)

	// Initialize balance maps for all group members
	for _, userID := range group.UserIDs {
		balances[userID.String()] = make(map[string]float64)
	}

	// Process each expense
	for _, expense := range expenses {
		// Get splits for this expense
		splits, err := s.expenseSplitRepo.FindByExpenseID(expense.ID.String())
		if err != nil {
			// If no splits found, skip this expense (might be old expense without splits)
			continue
		}

		paidByID := expense.PaidBy.String()

		// For each split, calculate what each user owes to the payer
		for _, split := range splits {
			// Skip if user is the payer (they don't owe themselves)
			if split.UserID == expense.PaidBy {
				continue
			}

			owedByID := split.UserID.String()

			// Add to balance: owedBy owes paidBy the split amount
			if balances[owedByID] == nil {
				balances[owedByID] = make(map[string]float64)
			}
			balances[owedByID][paidByID] += split.Amount
		}
	}

	// Convert map to list of balances
	balanceList := make([]inbound.Balance, 0)
	for fromUserID, toUsers := range balances {
		for toUserID, amount := range toUsers {
			if amount > 0.01 { // Only include balances > 1 cent
				balanceList = append(balanceList, inbound.Balance{
					FromUserID: fromUserID,
					ToUserID:   toUserID,
					Amount:     math.Round(amount*100) / 100,
				})
			}
		}
	}

	return &inbound.GroupBalance{
		GroupID:  groupID,
		Balances: balanceList,
	}, nil
}

// CalculateUserBalance calculates a user's overall balance across all groups
func (s *balanceService) CalculateUserBalance(userID string) (*inbound.UserBalance, error) {
	// Validate user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get all expense splits for this user
	splits, err := s.expenseSplitRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense splits: %w", err)
	}

	// Get all expenses where user paid
	expenses, err := s.expenseRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	totalOwed := 0.0    // User owes to others
	totalOwedTo := 0.0  // Others owe to user

	// Calculate what user owes (from splits)
	for _, split := range splits {
		// Get the expense to find who paid
		expense, err := s.expenseRepo.FindByID(split.ExpenseID.String())
		if err != nil {
			continue
		}

		// If user is not the payer, they owe the split amount
		if expense.PaidBy.String() != userID {
			totalOwed += split.Amount
		}
	}

	// Calculate what others owe to user (from expenses where user paid)
	for _, expense := range expenses {
		if expense.GroupID == nil {
			// User-to-user expense
			if expense.OwedBy != nil {
				// Someone owes the user
				totalOwedTo += expense.Amount
			}
		} else {
			// Group expense - get splits to see who owes
			splits, err := s.expenseSplitRepo.FindByExpenseID(expense.ID.String())
			if err != nil {
				continue
			}

			for _, split := range splits {
				// If split is for another user (not the payer), they owe the user
				if split.UserID.String() != userID {
					totalOwedTo += split.Amount
				}
			}
		}
	}

	netBalance := totalOwedTo - totalOwed

	return &inbound.UserBalance{
		UserID:      userID,
		TotalOwed:   math.Round(totalOwed*100) / 100,
		TotalOwedTo: math.Round(totalOwedTo*100) / 100,
		NetBalance:  math.Round(netBalance*100) / 100,
	}, nil
}

// CalculateUserToUserBalance calculates the balance between two specific users
func (s *balanceService) CalculateUserToUserBalance(userID1, userID2 string) (float64, error) {
	// Validate both users exist
	_, err := s.userRepo.FindByID(userID1)
	if err != nil {
		return 0, fmt.Errorf("user1 not found: %w", err)
	}

	_, err = s.userRepo.FindByID(userID2)
	if err != nil {
		return 0, fmt.Errorf("user2 not found: %w", err)
	}

	// Get all expense splits for user1
	splits, err := s.expenseSplitRepo.FindByUserID(userID1)
	if err != nil {
		return 0, fmt.Errorf("failed to get expense splits: %w", err)
	}

	balance := 0.0

	// Calculate balance: positive = user1 owes user2, negative = user2 owes user1
	for _, split := range splits {
		// Get the expense to find who paid
		expense, err := s.expenseRepo.FindByID(split.ExpenseID.String())
		if err != nil {
			continue
		}

		paidByID := expense.PaidBy.String()

		// If user2 paid and user1 owes, add to balance
		if paidByID == userID2 {
			balance += split.Amount
		}
	}

	// Also check expenses where user1 paid and user2 owes
	expenses, err := s.expenseRepo.FindByUserID(userID1)
	if err != nil {
		return 0, fmt.Errorf("failed to get expenses: %w", err)
	}

	for _, expense := range expenses {
		if expense.GroupID == nil {
			// User-to-user expense
			if expense.OwedBy != nil && expense.OwedBy.String() == userID2 {
				balance -= expense.Amount
			}
		} else {
			// Group expense - get splits
			splits, err := s.expenseSplitRepo.FindByExpenseID(expense.ID.String())
			if err != nil {
				continue
			}

			for _, split := range splits {
				if split.UserID.String() == userID2 {
					balance -= split.Amount
				}
			}
		}
	}

	return math.Round(balance*100) / 100, nil
}

