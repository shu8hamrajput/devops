package inbound

import "splitwise/domain/entity"

// GroupService defines the inbound port (driving port) for group operations
type GroupService interface {
	CreateGroup(name string, userIDs []string) (*entity.Group, error)
	GetGroupByID(id string) (*entity.Group, error)
	AddUserToGroup(groupID, userID string) error
}


