package service

import (
	"fmt"

	"github.com/google/uuid"
	"splitwise/domain/entity"
	"splitwise/domain/port/inbound"
	"splitwise/domain/port/outbound"
)

// groupService implements the GroupService port
type groupService struct {
	groupRepo outbound.GroupRepository
	userRepo  outbound.UserRepository
}

// NewGroupService creates a new group service
func NewGroupService(
	groupRepo outbound.GroupRepository,
	userRepo outbound.UserRepository,
) inbound.GroupService {
	return &groupService{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// CreateGroup creates a new group
func (s *groupService) CreateGroup(name string, userIDs []string) (*entity.Group, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	// Convert string IDs to UUIDs and validate users exist
	uuids := make([]uuid.UUID, 0, len(userIDs))
	for _, userID := range userIDs {
		uuid, err := uuid.Parse(userID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID: %s", userID)
		}

		// Validate user exists
		_, err = s.userRepo.FindByID(userID)
		if err != nil {
			return nil, fmt.Errorf("user not found: %s", userID)
		}

		uuids = append(uuids, uuid)
	}

	group := entity.NewGroup(name, uuids)
	if err := s.groupRepo.Create(group); err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return group, nil
}

// GetGroupByID retrieves a group by ID
func (s *groupService) GetGroupByID(id string) (*entity.Group, error) {
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	return group, nil
}

// AddUserToGroup adds a user to a group
func (s *groupService) AddUserToGroup(groupID, userID string) error {
	// Validate user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Get group
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	// Check if user already in group
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	for _, id := range group.UserIDs {
		if id == userUUID {
			return fmt.Errorf("user already in group")
		}
	}

	// Add user to group
	group.UserIDs = append(group.UserIDs, userUUID)
	group.UpdateTimestamp()
	if err := s.groupRepo.Update(group); err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	return nil
}

// GetGroupsByUser retrieves all groups where the user is a member
func (s *groupService) GetGroupsByUser(userID string) ([]*entity.Group, error) {
	groups, err := s.groupRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}
	return groups, nil
}

// GetAllGroups retrieves all groups in the system
func (s *groupService) GetAllGroups() ([]*entity.Group, error) {
	groups, err := s.groupRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get all groups: %w", err)
	}
	return groups, nil
}

// UpdateGroup updates an existing group
func (s *groupService) UpdateGroup(id, name string, userIDs []string) (*entity.Group, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	// Get existing group
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("group not found: %w", err)
	}

	// Convert string IDs to UUIDs and validate users exist
	uuids := make([]uuid.UUID, 0, len(userIDs))
	for _, userID := range userIDs {
		uuid, err := uuid.Parse(userID)
		if err != nil {
			return nil, fmt.Errorf("invalid user ID: %s", userID)
		}

		// Validate user exists
		_, err = s.userRepo.FindByID(userID)
		if err != nil {
			return nil, fmt.Errorf("user not found: %s", userID)
		}

		uuids = append(uuids, uuid)
	}

	// Update group
	group.Name = name
	group.UserIDs = uuids
	group.UpdateTimestamp()

	if err := s.groupRepo.Update(group); err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	return group, nil
}

// DeleteGroup deletes a group (soft delete)
func (s *groupService) DeleteGroup(id string) error {
	// Verify group exists
	_, err := s.groupRepo.FindByID(id)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	// Delete group
	if err := s.groupRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	return nil
}

// RemoveUserFromGroup removes a user from a group
func (s *groupService) RemoveUserFromGroup(groupID, userID string) error {
	// Validate user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Get group
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return fmt.Errorf("group not found: %w", err)
	}

	// Check if user is in group
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	found := false
	newUserIDs := make([]uuid.UUID, 0, len(group.UserIDs))
	for _, id := range group.UserIDs {
		if id == userUUID {
			found = true
		} else {
			newUserIDs = append(newUserIDs, id)
		}
	}

	if !found {
		return fmt.Errorf("user not in group")
	}

	// Update group with user removed
	group.UserIDs = newUserIDs
	group.UpdateTimestamp()
	if err := s.groupRepo.Update(group); err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	return nil
}


