package entity

import "github.com/google/uuid"

// Group represents a group of users
type Group struct {
	ID      uuid.UUID
	Name    string
	UserIDs []uuid.UUID
}

// NewGroup creates a new group entity
func NewGroup(name string, userIDs []uuid.UUID) *Group {
	return &Group{
		ID:      uuid.New(),
		Name:    name,
		UserIDs: userIDs,
	}
}


