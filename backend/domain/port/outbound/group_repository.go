package outbound

import "splitwise/domain/entity"

// GroupRepository defines the outbound port (driven port) for group persistence
type GroupRepository interface {
	Create(group *entity.Group) error
	FindByID(id string) (*entity.Group, error)
	FindByUserID(userID string) ([]*entity.Group, error)
	FindAll() ([]*entity.Group, error)
	Update(group *entity.Group) error
	Delete(id string) error
}


