package inbound

import "splitwise/domain/entity"

// AuthService defines the inbound port (driving port) for authentication operations
type AuthService interface {
	Register(name, email, password string) (*entity.User, error)
	Login(email, password string) (*entity.TokenPair, error)
	RefreshToken(refreshToken string) (*entity.TokenPair, error)
	ValidateToken(token string) (string, error) // Returns user ID if valid
}
