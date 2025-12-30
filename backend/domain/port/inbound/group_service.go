package inbound

import "splitwise/domain/entity"

// GroupService defines the inbound port (driving port) for group operations
type GroupService interface {
	CreateGroup(name string, userIDs []string) (*entity.Group, error)
	GetGroupByID(id string) (*entity.Group, error)
	GetGroupsByUser(userID string) ([]*entity.Group, error)
	GetAllGroups() ([]*entity.Group, error)
	UpdateGroup(id, name string, userIDs []string) (*entity.Group, error)
	DeleteGroup(id string) error
	AddUserToGroup(groupID, userID string) error
	RemoveUserFromGroup(groupID, userID string) error
}


