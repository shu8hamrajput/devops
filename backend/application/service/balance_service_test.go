package service

import (
	"testing"

	"github.com/google/uuid"
	"splitwise/domain/entity"
)

// mockExpenseRepositoryForBalance is a mock for balance testing
type mockExpenseRepositoryForBalance struct {
	expenses map[string]*entity.Expense
}

func (m *mockExpenseRepositoryForBalance) Create(expense *entity.Expense) error {
	if m.expenses == nil {
		m.expenses = make(map[string]*entity.Expense)
	}
	m.expenses[expense.ID.String()] = expense
	return nil
}

func (m *mockExpenseRepositoryForBalance) FindByID(id string) (*entity.Expense, error) {
	return m.expenses[id], nil
}

func (m *mockExpenseRepositoryForBalance) FindByGroupID(groupID string) ([]*entity.Expense, error) {
	expenses := make([]*entity.Expense, 0)
	for _, expense := range m.expenses {
		if expense.GroupID != nil && expense.GroupID.String() == groupID {
			expenses = append(expenses, expense)
		}
	}
	return expenses, nil
}

func (m *mockExpenseRepositoryForBalance) FindByUserID(userID string) ([]*entity.Expense, error) {
	expenses := make([]*entity.Expense, 0)
	for _, expense := range m.expenses {
		if expense.PaidBy.String() == userID {
			expenses = append(expenses, expense)
		}
	}
	return expenses, nil
}

func (m *mockExpenseRepositoryForBalance) Update(expense *entity.Expense) error {
	return nil
}

func (m *mockExpenseRepositoryForBalance) Delete(id string) error {
	delete(m.expenses, id)
	return nil
}

// mockExpenseSplitRepositoryForBalance is a mock for balance testing
type mockExpenseSplitRepositoryForBalance struct {
	splits map[string][]*entity.ExpenseSplit
}

func (m *mockExpenseSplitRepositoryForBalance) Create(split *entity.ExpenseSplit) error {
	if m.splits == nil {
		m.splits = make(map[string][]*entity.ExpenseSplit)
	}
	expenseID := split.ExpenseID.String()
	m.splits[expenseID] = append(m.splits[expenseID], split)
	return nil
}

func (m *mockExpenseSplitRepositoryForBalance) FindByExpenseID(expenseID string) ([]*entity.ExpenseSplit, error) {
	return m.splits[expenseID], nil
}

func (m *mockExpenseSplitRepositoryForBalance) FindByUserID(userID string) ([]*entity.ExpenseSplit, error) {
	allSplits := make([]*entity.ExpenseSplit, 0)
	for _, splits := range m.splits {
		for _, split := range splits {
			if split.UserID.String() == userID {
				allSplits = append(allSplits, split)
			}
		}
	}
	return allSplits, nil
}

func (m *mockExpenseSplitRepositoryForBalance) FindByExpenseAndUser(expenseID, userID string) (*entity.ExpenseSplit, error) {
	splits := m.splits[expenseID]
	for _, split := range splits {
		if split.UserID.String() == userID {
			return split, nil
		}
	}
	return nil, nil
}

func (m *mockExpenseSplitRepositoryForBalance) DeleteByExpenseID(expenseID string) error {
	delete(m.splits, expenseID)
	return nil
}

func (m *mockExpenseSplitRepositoryForBalance) Update(split *entity.ExpenseSplit) error {
	return nil
}

func TestBalanceService_CalculateGroupBalance(t *testing.T) {
	expenseRepo := &mockExpenseRepositoryForBalance{}
	splitRepo := &mockExpenseSplitRepositoryForBalance{}
	userRepo := &mockUserRepository{}
	groupRepo := &mockGroupRepository{}

	service := NewBalanceService(expenseRepo, splitRepo, userRepo, groupRepo)

	// Create test users
	user1 := entity.NewUser("User 1", "user1@test.com")
	user2 := entity.NewUser("User 2", "user2@test.com")
	user3 := entity.NewUser("User 3", "user3@test.com")
	userRepo.Create(user1)
	userRepo.Create(user2)
	userRepo.Create(user3)

	// Create test group
	group := entity.NewGroup("Test Group", []uuid.UUID{user1.ID, user2.ID, user3.ID})
	groupRepo.Create(group)

	// Create expense where user1 paid $100, split equally
	expense := entity.NewExpense("Test Expense", 100.0, user1.ID, group.ID)
	expenseRepo.Create(expense)

	// Create splits: each user owes $33.33 (approximately)
	split1 := entity.NewExpenseSplit(expense.ID, user1.ID, 33.33, entity.ShareTypeEqual, 1.0)
	split2 := entity.NewExpenseSplit(expense.ID, user2.ID, 33.33, entity.ShareTypeEqual, 1.0)
	split3 := entity.NewExpenseSplit(expense.ID, user3.ID, 33.34, entity.ShareTypeEqual, 1.0)
	splitRepo.Create(split1)
	splitRepo.Create(split2)
	splitRepo.Create(split3)

	// Calculate group balance
	groupBalance, err := service.CalculateGroupBalance(group.ID.String())
	if err != nil {
		t.Fatalf("Failed to calculate group balance: %v", err)
	}

	if groupBalance.GroupID != group.ID.String() {
		t.Errorf("Expected group ID %s, got %s", group.ID.String(), groupBalance.GroupID)
	}

	// Should have 2 balances: user2 owes user1, user3 owes user1
	if len(groupBalance.Balances) != 2 {
		t.Errorf("Expected 2 balances, got %d", len(groupBalance.Balances))
	}

	// Check that user2 and user3 owe user1
	foundUser2 := false
	foundUser3 := false
	for _, balance := range groupBalance.Balances {
		if balance.FromUserID == user2.ID.String() && balance.ToUserID == user1.ID.String() {
			foundUser2 = true
			if balance.Amount < 33.0 || balance.Amount > 34.0 {
				t.Errorf("Expected user2 to owe approximately 33.33, got %.2f", balance.Amount)
			}
		}
		if balance.FromUserID == user3.ID.String() && balance.ToUserID == user1.ID.String() {
			foundUser3 = true
			if balance.Amount < 33.0 || balance.Amount > 34.0 {
				t.Errorf("Expected user3 to owe approximately 33.34, got %.2f", balance.Amount)
			}
		}
	}

	if !foundUser2 {
		t.Error("Expected balance from user2 to user1")
	}
	if !foundUser3 {
		t.Error("Expected balance from user3 to user1")
	}
}

func TestBalanceService_CalculateUserBalance(t *testing.T) {
	expenseRepo := &mockExpenseRepositoryForBalance{}
	splitRepo := &mockExpenseSplitRepositoryForBalance{}
	userRepo := &mockUserRepository{}
	groupRepo := &mockGroupRepository{}

	service := NewBalanceService(expenseRepo, splitRepo, userRepo, groupRepo)

	// Create test users
	user1 := entity.NewUser("User 1", "user1@test.com")
	user2 := entity.NewUser("User 2", "user2@test.com")
	userRepo.Create(user1)
	userRepo.Create(user2)

	// Create test group
	group := entity.NewGroup("Test Group", []uuid.UUID{user1.ID, user2.ID})
	groupRepo.Create(group)

	// Create expense where user1 paid $100, split equally
	expense1 := entity.NewExpense("Expense 1", 100.0, user1.ID, group.ID)
	expenseRepo.Create(expense1)

	// Create splits: user2 owes user1 $50
	split1 := entity.NewExpenseSplit(expense1.ID, user1.ID, 50.0, entity.ShareTypeEqual, 1.0)
	split2 := entity.NewExpenseSplit(expense1.ID, user2.ID, 50.0, entity.ShareTypeEqual, 1.0)
	splitRepo.Create(split1)
	splitRepo.Create(split2)

	// Create another expense where user2 paid $60, split equally
	expense2 := entity.NewExpense("Expense 2", 60.0, user2.ID, group.ID)
	expenseRepo.Create(expense2)

	// Create splits: user1 owes user2 $30
	split3 := entity.NewExpenseSplit(expense2.ID, user1.ID, 30.0, entity.ShareTypeEqual, 1.0)
	split4 := entity.NewExpenseSplit(expense2.ID, user2.ID, 30.0, entity.ShareTypeEqual, 1.0)
	splitRepo.Create(split3)
	splitRepo.Create(split4)

	// Calculate user1 balance
	userBalance, err := service.CalculateUserBalance(user1.ID.String())
	if err != nil {
		t.Fatalf("Failed to calculate user balance: %v", err)
	}

	if userBalance.UserID != user1.ID.String() {
		t.Errorf("Expected user ID %s, got %s", user1.ID.String(), userBalance.UserID)
	}

	// User1 owes $30 (from expense2), is owed $50 (from expense1)
	// Net balance should be $20 (owed to user1)
	if userBalance.TotalOwed < 29.0 || userBalance.TotalOwed > 31.0 {
		t.Errorf("Expected totalOwed to be approximately 30.0, got %.2f", userBalance.TotalOwed)
	}

	if userBalance.TotalOwedTo < 49.0 || userBalance.TotalOwedTo > 51.0 {
		t.Errorf("Expected totalOwedTo to be approximately 50.0, got %.2f", userBalance.TotalOwedTo)
	}

	if userBalance.NetBalance < 19.0 || userBalance.NetBalance > 21.0 {
		t.Errorf("Expected netBalance to be approximately 20.0, got %.2f", userBalance.NetBalance)
	}
}

func TestBalanceService_CalculateUserToUserBalance(t *testing.T) {
	expenseRepo := &mockExpenseRepositoryForBalance{}
	splitRepo := &mockExpenseSplitRepositoryForBalance{}
	userRepo := &mockUserRepository{}
	groupRepo := &mockGroupRepository{}

	service := NewBalanceService(expenseRepo, splitRepo, userRepo, groupRepo)

	// Create test users
	user1 := entity.NewUser("User 1", "user1@test.com")
	user2 := entity.NewUser("User 2", "user2@test.com")
	userRepo.Create(user1)
	userRepo.Create(user2)

	// Create test group
	group := entity.NewGroup("Test Group", []uuid.UUID{user1.ID, user2.ID})
	groupRepo.Create(group)

	// Create expense where user1 paid $100, split equally
	expense1 := entity.NewExpense("Expense 1", 100.0, user1.ID, group.ID)
	expenseRepo.Create(expense1)

	// Create splits: user2 owes user1 $50
	split1 := entity.NewExpenseSplit(expense1.ID, user1.ID, 50.0, entity.ShareTypeEqual, 1.0)
	split2 := entity.NewExpenseSplit(expense1.ID, user2.ID, 50.0, entity.ShareTypeEqual, 1.0)
	splitRepo.Create(split1)
	splitRepo.Create(split2)

	// Create another expense where user2 paid $60, split equally
	expense2 := entity.NewExpense("Expense 2", 60.0, user2.ID, group.ID)
	expenseRepo.Create(expense2)

	// Create splits: user1 owes user2 $30
	split3 := entity.NewExpenseSplit(expense2.ID, user1.ID, 30.0, entity.ShareTypeEqual, 1.0)
	split4 := entity.NewExpenseSplit(expense2.ID, user2.ID, 30.0, entity.ShareTypeEqual, 1.0)
	splitRepo.Create(split3)
	splitRepo.Create(split4)

	// Calculate balance between user1 and user2
	// user1 owes user2 $30, user2 owes user1 $50
	// Net: user2 owes user1 $20
	balance, err := service.CalculateUserToUserBalance(user1.ID.String(), user2.ID.String())
	if err != nil {
		t.Fatalf("Failed to calculate user-to-user balance: %v", err)
	}

	// Positive balance means user1 owes user2, negative means user2 owes user1
	// Since user2 owes user1 $20, balance should be negative (-20)
	if balance < -21.0 || balance > -19.0 {
		t.Errorf("Expected balance to be approximately -20.0, got %.2f", balance)
	}
}

