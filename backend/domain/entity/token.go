package entity

import "time"

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // seconds until access token expires
}

// RefreshTokenData represents refresh token metadata stored in cache
type RefreshTokenData struct {
	UserID    string
	ExpiresAt time.Time
}
