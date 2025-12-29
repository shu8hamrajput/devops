package entity

import "github.com/google/uuid"

// User represents a user in the system
type User struct {
	ID    uuid.UUID
	Name  string
	Email string
}

// NewUser creates a new user entity
func NewUser(name, email string) *User {
	return &User{
		ID:    uuid.New(),
		Name:  name,
		Email: email,
	}
}


