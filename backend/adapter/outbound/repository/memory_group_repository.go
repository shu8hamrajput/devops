package repository

import (
	"fmt"
	"sync"

	"splitwise/domain/entity"
	"splitwise/domain/port/outbound"
)

// memoryGroupRepository is an in-memory implementation of GroupRepository
type memoryGroupRepository struct {
	groups map[string]*entity.Group
	mu     sync.RWMutex
}

// NewMemoryGroupRepository creates a new in-memory group repository
func NewMemoryGroupRepository() outbound.GroupRepository {
	return &memoryGroupRepository{
		groups: make(map[string]*entity.Group),
	}
}

// Create stores a new group
func (r *memoryGroupRepository) Create(group *entity.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.groups[group.ID.String()] = group
	return nil
}

// FindByID retrieves a group by ID
func (r *memoryGroupRepository) FindByID(id string) (*entity.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	group, exists := r.groups[id]
	if !exists {
		return nil, fmt.Errorf("group not found: %s", id)
	}

	return group, nil
}

// Update updates an existing group
func (r *memoryGroupRepository) Update(group *entity.Group) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.groups[group.ID.String()]; !exists {
		return fmt.Errorf("group not found: %s", group.ID.String())
	}

	r.groups[group.ID.String()] = group
	return nil
}

