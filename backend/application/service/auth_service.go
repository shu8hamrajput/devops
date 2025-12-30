package service

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"splitwise/domain/entity"
	"splitwise/domain/port/inbound"
	"splitwise/domain/port/outbound"
)

// authService implements the AuthService port
type authService struct {
	authRepo outbound.AuthRepository
	cache    outbound.Cache
	jwtSecret []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(authRepo outbound.AuthRepository, cache outbound.Cache) inbound.AuthService {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key-change-in-production" // Default for development
	}

	accessExpiry := 15 * time.Minute  // Access token expires in 15 minutes
	refreshExpiry := 7 * 24 * time.Hour // Refresh token expires in 7 days

	return &authService{
		authRepo:          authRepo,
		cache:             cache,
		jwtSecret:         []byte(jwtSecret),
		accessTokenExpiry: accessExpiry,
		refreshTokenExpiry: refreshExpiry,
	}
}

// Register creates a new user account
func (s *authService) Register(name, email, password string) (*entity.User, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	// Check if user already exists
	existingUser, err := s.authRepo.FindUserByEmail(email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// Create new user
	user := entity.NewUser(name, email)
	if err := user.SetPassword(password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.authRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns token pair
func (s *authService) Login(email, password string) (*entity.TokenPair, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	user, err := s.authRepo.FindUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if !user.CheckPassword(password) {
		return nil, fmt.Errorf("invalid email or password")
	}

	return s.generateTokenPair(user.ID.String())
}

// RefreshToken generates a new token pair from a refresh token
func (s *authService) RefreshToken(refreshToken string) (*entity.TokenPair, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	// Get refresh token data from cache
	tokenData, err := s.cache.GetRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if token is expired
	if time.Now().After(tokenData.ExpiresAt) {
		s.cache.DeleteRefreshToken(refreshToken)
		return nil, fmt.Errorf("refresh token expired")
	}

	// Generate new token pair
	return s.generateTokenPair(tokenData.UserID)
}

// ValidateToken validates an access token and returns the user ID
func (s *authService) ValidateToken(token string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	if !tkn.Valid {
		return "", errors.New("invalid token")
	}

	if claims.Subject == "" {
		return "", errors.New("token missing subject")
	}

	return claims.Subject, nil
}

// generateTokenPair creates both access and refresh tokens
func (s *authService) generateTokenPair(userID string) (*entity.TokenPair, error) {
	// Generate access token
	accessToken, accessExpiresAt, err := s.generateAccessToken(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken := uuid.New().String()
	refreshExpiresAt := time.Now().Add(s.refreshTokenExpiry)

	// Store refresh token in cache
	tokenData := &entity.RefreshTokenData{
		UserID:    userID,
		ExpiresAt: refreshExpiresAt,
	}
	if err := s.cache.SetRefreshToken(refreshToken, tokenData); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	expiresIn := int64(accessExpiresAt.Sub(time.Now()).Seconds())

	return &entity.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

// generateAccessToken creates a JWT access token
func (s *authService) generateAccessToken(userID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.accessTokenExpiry)
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
