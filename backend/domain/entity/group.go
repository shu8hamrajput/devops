package entity

import (
	"time"

	"github.com/google/uuid"
)

// Group represents a group of users
type Group struct {
	ID        uuid.UUID
	Name      string
	UserIDs   []uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewGroup creates a new group entity
func NewGroup(name string, userIDs []uuid.UUID) *Group {
	now := time.Now()
	return &Group{
		ID:        uuid.New(),
		Name:      name,
		UserIDs:   userIDs,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateTimestamp updates the UpdatedAt timestamp
func (g *Group) UpdateTimestamp() {
	g.UpdatedAt = time.Now()
}


