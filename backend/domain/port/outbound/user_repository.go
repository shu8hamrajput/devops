package outbound

import "splitwise/domain/entity"

// UserRepository defines the outbound port (driven port) for user persistence
// This is the interface that the domain expects, implemented by adapters
type UserRepository interface {
	Create(user *entity.User) error
	FindByID(id string) (*entity.User, error)
	FindByEmail(email string) (*entity.User, error)
	FindAll() ([]*entity.User, error)
}


