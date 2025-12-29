package service

import (
	"fmt"

	"splitwise/domain/entity"
	"splitwise/domain/port/inbound"
	"splitwise/domain/port/outbound"
)

// userService implements the UserService port
type userService struct {
	userRepo outbound.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo outbound.UserRepository) inbound.UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(name, email string) (*entity.User, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	user := entity.NewUser(name, email)
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(id string) (*entity.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetAllUsers retrieves all users
func (s *userService) GetAllUsers() ([]*entity.User, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return users, nil
}


