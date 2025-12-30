package service

import (
	"testing"

	"github.com/google/uuid"
	"splitwise/domain/entity"
)

// mockExpenseSplitRepository is a mock implementation for testing
type mockExpenseSplitRepository struct {
	splits map[string][]*entity.ExpenseSplit
}

func (m *mockExpenseSplitRepository) Create(split *entity.ExpenseSplit) error {
	if m.splits == nil {
		m.splits = make(map[string][]*entity.ExpenseSplit)
	}
	expenseID := split.ExpenseID.String()
	m.splits[expenseID] = append(m.splits[expenseID], split)
	return nil
}

func (m *mockExpenseSplitRepository) FindByExpenseID(expenseID string) ([]*entity.ExpenseSplit, error) {
	return m.splits[expenseID], nil
}

func (m *mockExpenseSplitRepository) FindByUserID(userID string) ([]*entity.ExpenseSplit, error) {
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

func (m *mockExpenseSplitRepository) FindByExpenseAndUser(expenseID, userID string) (*entity.ExpenseSplit, error) {
	splits := m.splits[expenseID]
	for _, split := range splits {
		if split.UserID.String() == userID {
			return split, nil
		}
	}
	return nil, nil
}

func (m *mockExpenseSplitRepository) DeleteByExpenseID(expenseID string) error {
	delete(m.splits, expenseID)
	return nil
}

func (m *mockExpenseSplitRepository) Update(split *entity.ExpenseSplit) error {
	return nil
}

// mockUserRepository for testing
type mockUserRepository struct {
	users map[string]*entity.User
}

func (m *mockUserRepository) Create(user *entity.User) error {
	if m.users == nil {
		m.users = make(map[string]*entity.User)
	}
	m.users[user.ID.String()] = user
	return nil
}

func (m *mockUserRepository) FindByID(id string) (*entity.User, error) {
	return m.users[id], nil
}

func (m *mockUserRepository) FindByEmail(email string) (*entity.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepository) FindAll() ([]*entity.User, error) {
	users := make([]*entity.User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users, nil
}

func (m *mockUserRepository) Update(user *entity.User) error {
	return nil
}

// mockGroupRepository for testing
type mockGroupRepository struct {
	groups map[string]*entity.Group
}

func (m *mockGroupRepository) Create(group *entity.Group) error {
	if m.groups == nil {
		m.groups = make(map[string]*entity.Group)
	}
	m.groups[group.ID.String()] = group
	return nil
}

func (m *mockGroupRepository) FindByID(id string) (*entity.Group, error) {
	return m.groups[id], nil
}

func (m *mockGroupRepository) FindByUserID(userID string) ([]*entity.Group, error) {
	groups := make([]*entity.Group, 0)
	for _, group := range m.groups {
		for _, uid := range group.UserIDs {
			if uid.String() == userID {
				groups = append(groups, group)
				break
			}
		}
	}
	return groups, nil
}

func (m *mockGroupRepository) Update(group *entity.Group) error {
	return nil
}

func (m *mockGroupRepository) FindAll() ([]*entity.Group, error) {
	groups := make([]*entity.Group, 0, len(m.groups))
	for _, group := range m.groups {
		groups = append(groups, group)
	}
	return groups, nil
}

func (m *mockGroupRepository) Delete(id string) error {
	delete(m.groups, id)
	return nil
}

func TestExpenseSplitService_SplitEqual(t *testing.T) {
	splitRepo := &mockExpenseSplitRepository{}
	userRepo := &mockUserRepository{}
	groupRepo := &mockGroupRepository{}

	service := NewExpenseSplitService(splitRepo, userRepo, groupRepo)

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

	// Create expense
	expense := entity.NewExpense("Test Expense", 300.0, user1.ID, group.ID)

	// Split equally
	splits, err := service.SplitExpense(expense, entity.ShareTypeEqual, nil)
	if err != nil {
		t.Fatalf("Failed to split expense: %v", err)
	}

	if len(splits) != 3 {
		t.Errorf("Expected 3 splits, got %d", len(splits))
	}

	total := 0.0
	for _, split := range splits {
		total += split.Amount
		if split.ShareType != entity.ShareTypeEqual {
			t.Errorf("Expected ShareTypeEqual, got %s", split.ShareType)
		}
	}

	if total != 300.0 {
		t.Errorf("Expected total to be 300.0, got %.2f", total)
	}
}

func TestExpenseSplitService_SplitByPercentage(t *testing.T) {
	splitRepo := &mockExpenseSplitRepository{}
	userRepo := &mockUserRepository{}
	groupRepo := &mockGroupRepository{}

	service := NewExpenseSplitService(splitRepo, userRepo, groupRepo)

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

	// Create expense
	expense := entity.NewExpense("Test Expense", 100.0, user1.ID, group.ID)

	// Split by percentage: 50%, 30%, 20%
	userShares := map[string]float64{
		user1.ID.String(): 50.0,
		user2.ID.String(): 30.0,
		user3.ID.String(): 20.0,
	}

	splits, err := service.SplitExpense(expense, entity.ShareTypePercentage, userShares)
	if err != nil {
		t.Fatalf("Failed to split expense: %v", err)
	}

	if len(splits) != 3 {
		t.Errorf("Expected 3 splits, got %d", len(splits))
	}

	total := 0.0
	for _, split := range splits {
		total += split.Amount
		if split.ShareType != entity.ShareTypePercentage {
			t.Errorf("Expected ShareTypePercentage, got %s", split.ShareType)
		}
	}

	if total < 99.99 || total > 100.01 {
		t.Errorf("Expected total to be approximately 100.0, got %.2f", total)
	}
}

func TestExpenseSplitService_SplitByExactAmount(t *testing.T) {
	splitRepo := &mockExpenseSplitRepository{}
	userRepo := &mockUserRepository{}
	groupRepo := &mockGroupRepository{}

	service := NewExpenseSplitService(splitRepo, userRepo, groupRepo)

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

	// Create expense
	expense := entity.NewExpense("Test Expense", 100.0, user1.ID, group.ID)

	// Split by exact amounts: 40, 35, 25
	userShares := map[string]float64{
		user1.ID.String(): 40.0,
		user2.ID.String(): 35.0,
		user3.ID.String(): 25.0,
	}

	splits, err := service.SplitExpense(expense, entity.ShareTypeExactAmount, userShares)
	if err != nil {
		t.Fatalf("Failed to split expense: %v", err)
	}

	if len(splits) != 3 {
		t.Errorf("Expected 3 splits, got %d", len(splits))
	}

	total := 0.0
	for _, split := range splits {
		total += split.Amount
		if split.ShareType != entity.ShareTypeExactAmount {
			t.Errorf("Expected ShareTypeExactAmount, got %s", split.ShareType)
		}
	}

	if total != 100.0 {
		t.Errorf("Expected total to be 100.0, got %.2f", total)
	}
}

func TestExpenseSplitService_SplitByShares(t *testing.T) {
	splitRepo := &mockExpenseSplitRepository{}
	userRepo := &mockUserRepository{}
	groupRepo := &mockGroupRepository{}

	service := NewExpenseSplitService(splitRepo, userRepo, groupRepo)

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

	// Create expense
	expense := entity.NewExpense("Test Expense", 100.0, user1.ID, group.ID)

	// Split by shares: 2:1:1 (user1 gets 2/4 = 50%, user2 gets 1/4 = 25%, user3 gets 1/4 = 25%)
	userShares := map[string]float64{
		user1.ID.String(): 2.0,
		user2.ID.String(): 1.0,
		user3.ID.String(): 1.0,
	}

	splits, err := service.SplitExpense(expense, entity.ShareTypeShares, userShares)
	if err != nil {
		t.Fatalf("Failed to split expense: %v", err)
	}

	if len(splits) != 3 {
		t.Errorf("Expected 3 splits, got %d", len(splits))
	}

	total := 0.0
	for _, split := range splits {
		total += split.Amount
		if split.ShareType != entity.ShareTypeShares {
			t.Errorf("Expected ShareTypeShares, got %s", split.ShareType)
		}
	}

	if total < 99.99 || total > 100.01 {
		t.Errorf("Expected total to be approximately 100.0, got %.2f", total)
	}
}

func TestExpenseSplitService_ValidateSplitSum(t *testing.T) {
	splitRepo := &mockExpenseSplitRepository{}
	userRepo := &mockUserRepository{}
	groupRepo := &mockGroupRepository{}

	service := NewExpenseSplitService(splitRepo, userRepo, groupRepo)

	// Create test users
	user1 := entity.NewUser("User 1", "user1@test.com")
	user2 := entity.NewUser("User 2", "user2@test.com")
	userRepo.Create(user1)
	userRepo.Create(user2)

	// Create test group
	group := entity.NewGroup("Test Group", []uuid.UUID{user1.ID, user2.ID})
	groupRepo.Create(group)

	// Create expense
	expense := entity.NewExpense("Test Expense", 100.0, user1.ID, group.ID)

	// Try to split with amounts that don't sum to expense amount
	userShares := map[string]float64{
		user1.ID.String(): 40.0,
		user2.ID.String(): 50.0, // Total is 90, not 100
	}

	_, err := service.SplitExpense(expense, entity.ShareTypeExactAmount, userShares)
	if err == nil {
		t.Error("Expected error when splits don't sum to expense amount")
	}
}

