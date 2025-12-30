package outbound

import "splitwise/domain/entity"

// Cache defines the outbound port for caching operations
type Cache interface {
	SetRefreshToken(token string, data *entity.RefreshTokenData) error
	GetRefreshToken(token string) (*entity.RefreshTokenData, error)
	DeleteRefreshToken(token string) error
}
