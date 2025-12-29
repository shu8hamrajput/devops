package inbound

import "splitwise/domain/entity"

// UserService defines the inbound port (driving port) for user operations
// This is the interface that external adapters (like HTTP handlers) will use
type UserService interface {
	CreateUser(name, email string) (*entity.User, error)
	GetUserByID(id string) (*entity.User, error)
	GetAllUsers() ([]*entity.User, error)
}


