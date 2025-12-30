package repository

import (
	"fmt"
	"sort"
	"sync"

	"github.com/google/uuid"
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

// FindByUserID retrieves all groups where the user is a member, sorted by UpdatedAt descending
func (r *memoryGroupRepository) FindByUserID(userID string) ([]*entity.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	groups := make([]*entity.Group, 0)
	for _, group := range r.groups {
		for _, groupUserID := range group.UserIDs {
			if groupUserID == userUUID {
				groups = append(groups, group)
				break
			}
		}
	}

	// Sort by UpdatedAt descending (most recently modified first)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].UpdatedAt.After(groups[j].UpdatedAt)
	})

	return groups, nil
}

// FindAll retrieves all groups, sorted by UpdatedAt descending
func (r *memoryGroupRepository) FindAll() ([]*entity.Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	groups := make([]*entity.Group, 0, len(r.groups))
	for _, group := range r.groups {
		groups = append(groups, group)
	}

	// Sort by UpdatedAt descending (most recently modified first)
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].UpdatedAt.After(groups[j].UpdatedAt)
	})

	return groups, nil
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

// Delete performs a soft delete on a group (removes from map)
func (r *memoryGroupRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.groups[id]; !exists {
		return fmt.Errorf("group not found: %s", id)
	}

	delete(r.groups, id)
	return nil
}

