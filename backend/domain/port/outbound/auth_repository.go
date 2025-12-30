package outbound

import "splitwise/domain/entity"

// AuthRepository defines the outbound port for authentication-related persistence
type AuthRepository interface {
	FindUserByEmail(email string) (*entity.User, error)
	CreateUser(user *entity.User) error
}
