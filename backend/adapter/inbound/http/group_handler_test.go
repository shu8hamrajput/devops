package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"splitwise/domain/entity"
)

// mockGroupService is a mock implementation of GroupService for testing
type mockGroupService struct {
	groups map[string]*entity.Group
}

func (m *mockGroupService) CreateGroup(name string, userIDs []string) (*entity.Group, error) {
	uuids := make([]uuid.UUID, len(userIDs))
	for i, id := range userIDs {
		uuids[i], _ = uuid.Parse(id)
	}
	group := entity.NewGroup(name, uuids)
	m.groups[group.ID.String()] = group
	return group, nil
}

func (m *mockGroupService) GetGroupByID(id string) (*entity.Group, error) {
	group, exists := m.groups[id]
	if !exists {
		return nil, &NotFoundError{Message: "group not found"}
	}
	return group, nil
}

func (m *mockGroupService) GetGroupsByUser(userID string) ([]*entity.Group, error) {
	userUUID, _ := uuid.Parse(userID)
	result := []*entity.Group{}
	for _, group := range m.groups {
		for _, uid := range group.UserIDs {
			if uid == userUUID {
				result = append(result, group)
				break
			}
		}
	}
	return result, nil
}

func (m *mockGroupService) GetAllGroups() ([]*entity.Group, error) {
	groups := make([]*entity.Group, 0, len(m.groups))
	for _, group := range m.groups {
		groups = append(groups, group)
	}
	return groups, nil
}

func (m *mockGroupService) AddUserToGroup(groupID, userID string) error {
	group, exists := m.groups[groupID]
	if !exists {
		return fmt.Errorf("group not found")
	}
	userUUID, _ := uuid.Parse(userID)
	for _, id := range group.UserIDs {
		if id == userUUID {
			return fmt.Errorf("user already in group")
		}
	}
	group.UserIDs = append(group.UserIDs, userUUID)
	return nil
}

func (m *mockGroupService) UpdateGroup(id, name string, userIDs []string) (*entity.Group, error) {
	group, exists := m.groups[id]
	if !exists {
		return nil, fmt.Errorf("group not found")
	}
	uuids := make([]uuid.UUID, len(userIDs))
	for i, uid := range userIDs {
		uuids[i], _ = uuid.Parse(uid)
	}
	group.Name = name
	group.UserIDs = uuids
	return group, nil
}

func (m *mockGroupService) DeleteGroup(id string) error {
	if _, exists := m.groups[id]; !exists {
		return fmt.Errorf("group not found")
	}
	delete(m.groups, id)
	return nil
}

func (m *mockGroupService) RemoveUserFromGroup(groupID, userID string) error {
	group, exists := m.groups[groupID]
	if !exists {
		return fmt.Errorf("group not found")
	}
	userUUID, _ := uuid.Parse(userID)
	newUserIDs := []uuid.UUID{}
	found := false
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
	group.UserIDs = newUserIDs
	return nil
}

func TestGroupHandler_GetAllGroups(t *testing.T) {
	// Create mock service with test data
	user1ID := uuid.New()
	user2ID := uuid.New()
	mockService := &mockGroupService{
		groups: make(map[string]*entity.Group),
	}
	
	group1 := entity.NewGroup("Test Group 1", []uuid.UUID{user1ID, user2ID})
	group2 := entity.NewGroup("Test Group 2", []uuid.UUID{user1ID})
	mockService.groups[group1.ID.String()] = group1
	mockService.groups[group2.ID.String()] = group2

	handler := NewGroupHandler(mockService)

	t.Run("Get all groups successfully", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/groups", nil)
		w := httptest.NewRecorder()

		handler.GetAllGroups(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response []GroupResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response) != 2 {
			t.Errorf("Expected 2 groups, got %d", len(response))
		}

		if response[0].Name != "Test Group 1" && response[0].Name != "Test Group 2" {
			t.Errorf("Unexpected group name: %s", response[0].Name)
		}
	})

	t.Run("Get all groups when empty", func(t *testing.T) {
		emptyService := &mockGroupService{groups: make(map[string]*entity.Group)}
		handler := NewGroupHandler(emptyService)

		req := httptest.NewRequest("GET", "/groups", nil)
		w := httptest.NewRecorder()

		handler.GetAllGroups(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response []GroupResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if len(response) != 0 {
			t.Errorf("Expected 0 groups, got %d", len(response))
		}
	})
}

func TestGroupHandler_CreateGroup(t *testing.T) {
	mockService := &mockGroupService{groups: make(map[string]*entity.Group)}
	handler := NewGroupHandler(mockService)

	t.Run("Create group successfully", func(t *testing.T) {
		user1ID := uuid.New().String()
		user2ID := uuid.New().String()

		reqBody := CreateGroupRequest{
			Name:    "New Group",
			UserIDs: []string{user1ID, user2ID},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("POST", "/groups", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.CreateGroup(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
		}

		var response GroupResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Name != "New Group" {
			t.Errorf("Expected name 'New Group', got '%s'", response.Name)
		}

		if len(response.UserIDs) != 2 {
			t.Errorf("Expected 2 user IDs, got %d", len(response.UserIDs))
		}
	})
}

// Add a simple NotFoundError type for the mock
type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

