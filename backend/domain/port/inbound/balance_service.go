package inbound

// Balance represents a balance between two users
type Balance struct {
	FromUserID string  // User who owes money
	ToUserID   string  // User who is owed money
	Amount     float64 // Amount owed
}

// UserBalance represents a user's overall balance
type UserBalance struct {
	UserID      string
	TotalOwed   float64 // Total amount this user owes to others
	TotalOwedTo float64 // Total amount others owe to this user
	NetBalance  float64 // Net balance (positive = owed to user, negative = user owes)
}

// GroupBalance represents balances within a group
type GroupBalance struct {
	GroupID string
	Balances []Balance // All balances in the group
}

// BalanceService defines the inbound port for balance operations
type BalanceService interface {
	// CalculateGroupBalance calculates who owes whom in a group
	CalculateGroupBalance(groupID string) (*GroupBalance, error)
	
	// CalculateUserBalance calculates a user's overall balance across all groups
	CalculateUserBalance(userID string) (*UserBalance, error)
	
	// CalculateUserToUserBalance calculates the balance between two specific users
	CalculateUserToUserBalance(userID1, userID2 string) (float64, error)
}

