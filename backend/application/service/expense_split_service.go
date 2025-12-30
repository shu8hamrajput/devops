package service

import (
	"fmt"
	"math"

	"github.com/google/uuid"
	"splitwise/domain/entity"
	"splitwise/domain/port/outbound"
)

// expenseSplitService handles expense splitting logic
type expenseSplitService struct {
	splitRepo outbound.ExpenseSplitRepository
	userRepo  outbound.UserRepository
	groupRepo outbound.GroupRepository
}

// NewExpenseSplitService creates a new expense split service
func NewExpenseSplitService(
	splitRepo outbound.ExpenseSplitRepository,
	userRepo outbound.UserRepository,
	groupRepo outbound.GroupRepository,
) *expenseSplitService {
	return &expenseSplitService{
		splitRepo: splitRepo,
		userRepo:  userRepo,
		groupRepo: groupRepo,
	}
}

// SplitExpenseRequest represents a request to split an expense
type SplitExpenseRequest struct {
	ExpenseID  string
	ShareType  entity.ShareType
	UserShares map[string]float64 // userID -> share value (percentage, amount, or share count)
}

// SplitExpense splits an expense among users based on the share type
func (s *expenseSplitService) SplitExpense(expense *entity.Expense, shareType entity.ShareType, userShares map[string]float64) ([]*entity.ExpenseSplit, error) {
	if expense.GroupID == nil {
		return nil, fmt.Errorf("expense must be associated with a group to split")
	}

	// Get group members
	group, err := s.groupRepo.FindByID(expense.GroupID.String())
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	if len(group.UserIDs) == 0 {
		return nil, fmt.Errorf("group has no members")
	}

	var splits []*entity.ExpenseSplit

	switch shareType {
	case entity.ShareTypeEqual:
		splits, err = s.splitEqual(expense, group.UserIDs)
	case entity.ShareTypePercentage:
		splits, err = s.splitByPercentage(expense, group.UserIDs, userShares)
	case entity.ShareTypeExactAmount:
		splits, err = s.splitByExactAmount(expense, group.UserIDs, userShares)
	case entity.ShareTypeShares:
		splits, err = s.splitByShares(expense, group.UserIDs, userShares)
	default:
		return nil, fmt.Errorf("invalid share type: %s", shareType)
	}

	if err != nil {
		return nil, err
	}

	// Validate that splits sum to expense amount (with small tolerance for floating point)
	total := 0.0
	for _, split := range splits {
		total += split.Amount
	}

	tolerance := 0.01 // 1 cent tolerance
	if math.Abs(total-expense.Amount) > tolerance {
		return nil, fmt.Errorf("splits sum to %.2f, but expense amount is %.2f", total, expense.Amount)
	}

	// Save splits to database
	for _, split := range splits {
		if err := s.splitRepo.Create(split); err != nil {
			// Clean up any splits that were created
			s.splitRepo.DeleteByExpenseID(expense.ID.String())
			return nil, fmt.Errorf("failed to create expense split: %w", err)
		}
	}

	return splits, nil
}

// splitEqual splits expense equally among all group members
func (s *expenseSplitService) splitEqual(expense *entity.Expense, userIDs []uuid.UUID) ([]*entity.ExpenseSplit, error) {
	if len(userIDs) == 0 {
		return nil, fmt.Errorf("no users to split expense")
	}

	amountPerUser := expense.Amount / float64(len(userIDs))
	// Round to 2 decimal places
	amountPerUser = math.Round(amountPerUser*100) / 100

	splits := make([]*entity.ExpenseSplit, 0, len(userIDs))
	
	// Distribute the amount, handling rounding differences
	totalDistributed := 0.0
	for i, userID := range userIDs {
		var amount float64
		if i == len(userIDs)-1 {
			// Last user gets the remainder to ensure total equals expense amount
			amount = expense.Amount - totalDistributed
		} else {
			amount = amountPerUser
		}
		
		split := entity.NewExpenseSplit(
			expense.ID,
			userID,
			amount,
			entity.ShareTypeEqual,
			1.0, // Equal share value is 1.0
		)
		splits = append(splits, split)
		totalDistributed += amount
	}

	return splits, nil
}

// splitByPercentage splits expense by percentage
func (s *expenseSplitService) splitByPercentage(expense *entity.Expense, userIDs []uuid.UUID, userShares map[string]float64) ([]*entity.ExpenseSplit, error) {
	if len(userShares) != len(userIDs) {
		return nil, fmt.Errorf("number of user shares must match number of group members")
	}

	totalPercentage := 0.0
	for _, percentage := range userShares {
		if percentage < 0 || percentage > 100 {
			return nil, fmt.Errorf("percentage must be between 0 and 100")
		}
		totalPercentage += percentage
	}

	// Allow small tolerance for floating point errors
	if math.Abs(totalPercentage-100.0) > 0.01 {
		return nil, fmt.Errorf("percentages must sum to 100, got %.2f", totalPercentage)
	}

	splits := make([]*entity.ExpenseSplit, 0, len(userIDs))
	for _, userID := range userIDs {
		userIDStr := userID.String()
		percentage, exists := userShares[userIDStr]
		if !exists {
			return nil, fmt.Errorf("user %s not found in user shares", userIDStr)
		}

		amount := (expense.Amount * percentage) / 100.0
		amount = math.Round(amount*100) / 100

		split := entity.NewExpenseSplit(
			expense.ID,
			userID,
			amount,
			entity.ShareTypePercentage,
			percentage,
		)
		splits = append(splits, split)
	}

	// Adjust last split to account for rounding
	totalDistributed := 0.0
	for i := 0; i < len(splits)-1; i++ {
		totalDistributed += splits[i].Amount
	}
	splits[len(splits)-1].Amount = expense.Amount - totalDistributed

	return splits, nil
}

// splitByExactAmount splits expense by exact amounts
func (s *expenseSplitService) splitByExactAmount(expense *entity.Expense, userIDs []uuid.UUID, userShares map[string]float64) ([]*entity.ExpenseSplit, error) {
	if len(userShares) != len(userIDs) {
		return nil, fmt.Errorf("number of user shares must match number of group members")
	}

	totalAmount := 0.0
	for _, amount := range userShares {
		if amount < 0 {
			return nil, fmt.Errorf("amount cannot be negative")
		}
		totalAmount += amount
	}

	if math.Abs(totalAmount-expense.Amount) > 0.01 {
		return nil, fmt.Errorf("exact amounts must sum to expense amount %.2f, got %.2f", expense.Amount, totalAmount)
	}

	splits := make([]*entity.ExpenseSplit, 0, len(userIDs))
	for _, userID := range userIDs {
		userIDStr := userID.String()
		amount, exists := userShares[userIDStr]
		if !exists {
			return nil, fmt.Errorf("user %s not found in user shares", userIDStr)
		}

		split := entity.NewExpenseSplit(
			expense.ID,
			userID,
			amount,
			entity.ShareTypeExactAmount,
			amount,
		)
		splits = append(splits, split)
	}

	return splits, nil
}

// splitByShares splits expense by shares (e.g., 2:1:1 means user1 gets 2/4, user2 gets 1/4, user3 gets 1/4)
func (s *expenseSplitService) splitByShares(expense *entity.Expense, userIDs []uuid.UUID, userShares map[string]float64) ([]*entity.ExpenseSplit, error) {
	if len(userShares) != len(userIDs) {
		return nil, fmt.Errorf("number of user shares must match number of group members")
	}

	totalShares := 0.0
	for _, share := range userShares {
		if share < 0 {
			return nil, fmt.Errorf("share cannot be negative")
		}
		totalShares += share
	}

	if totalShares == 0 {
		return nil, fmt.Errorf("total shares cannot be zero")
	}

	splits := make([]*entity.ExpenseSplit, 0, len(userIDs))
	totalDistributed := 0.0
	
	for i, userID := range userIDs {
		userIDStr := userID.String()
		share, exists := userShares[userIDStr]
		if !exists {
			return nil, fmt.Errorf("user %s not found in user shares", userIDStr)
		}

		var amount float64
		if i == len(userIDs)-1 {
			// Last user gets the remainder
			amount = expense.Amount - totalDistributed
		} else {
			amount = (expense.Amount * share) / totalShares
			amount = math.Round(amount*100) / 100
		}

		split := entity.NewExpenseSplit(
			expense.ID,
			userID,
			amount,
			entity.ShareTypeShares,
			share,
		)
		splits = append(splits, split)
		totalDistributed += amount
	}

	return splits, nil
}

