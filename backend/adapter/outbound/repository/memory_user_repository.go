package repository

import (
	"fmt"
	"sync"

	"splitwise/domain/entity"
	"splitwise/domain/port/outbound"
)

// memoryUserRepository is an in-memory implementation of UserRepository
type memoryUserRepository struct {
	users map[string]*entity.User
	mu    sync.RWMutex
}

// NewMemoryUserRepository creates a new in-memory user repository
func NewMemoryUserRepository() outbound.UserRepository {
	return &memoryUserRepository{
		users: make(map[string]*entity.User),
	}
}

// Create stores a new user
func (r *memoryUserRepository) Create(user *entity.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.users[user.ID.String()] = user
	return nil
}

// FindByID retrieves a user by ID
func (r *memoryUserRepository) FindByID(id string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}

	return user, nil
}

// FindByEmail retrieves a user by email
func (r *memoryUserRepository) FindByEmail(email string) (*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found with email: %s", email)
}

// FindAll retrieves all users
func (r *memoryUserRepository) FindAll() ([]*entity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*entity.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}

