package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"splitwise/domain/entity"
)

func TestGroupHandler_UpdateGroup(t *testing.T) {
	user1ID := uuid.New().String()
	user2ID := uuid.New().String()
	user3ID := uuid.New().String()

	mockService := &mockGroupService{
		groups: make(map[string]*entity.Group),
	}

	// Create a test group
	group, _ := mockService.CreateGroup("Test Group", []string{user1ID, user2ID})
	groupID := group.ID.String()

	handler := NewGroupHandler(mockService)

	t.Run("Update group successfully", func(t *testing.T) {
		reqBody := UpdateGroupRequest{
			Name:    "Updated Group",
			UserIDs: []string{user1ID, user2ID, user3ID},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/groups/"+groupID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", user1ID) // Authorized user (member)
		
		// Set up mux vars
		vars := map[string]string{"id": groupID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.UpdateGroup(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response GroupResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Name != "Updated Group" {
			t.Errorf("Expected name 'Updated Group', got '%s'", response.Name)
		}

		if len(response.UserIDs) != 3 {
			t.Errorf("Expected 3 user IDs, got %d", len(response.UserIDs))
		}
	})

	t.Run("Update group unauthorized", func(t *testing.T) {
		// Create a fresh group for this test
		freshGroup, _ := mockService.CreateGroup("Unauthorized Test Group", []string{user1ID, user2ID})
		freshGroupID := freshGroup.ID.String()

		reqBody := UpdateGroupRequest{
			Name:    "Updated Group",
			UserIDs: []string{user1ID, user2ID},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/groups/"+freshGroupID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", user3ID) // Not a member (unauthorized)
		
		// Set up mux vars
		vars := map[string]string{"id": freshGroupID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.UpdateGroup(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("Update non-existent group", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		reqBody := UpdateGroupRequest{
			Name:    "Updated Group",
			UserIDs: []string{user1ID},
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/groups/"+nonExistentID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", user1ID)
		
		// Set up mux vars
		vars := map[string]string{"id": nonExistentID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.UpdateGroup(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestGroupHandler_DeleteGroup(t *testing.T) {
	user1ID := uuid.New().String()
	user2ID := uuid.New().String()
	user3ID := uuid.New().String()

	mockService := &mockGroupService{
		groups: make(map[string]*entity.Group),
	}

	// Create a test group
	group, _ := mockService.CreateGroup("Test Group", []string{user1ID, user2ID})
	groupID := group.ID.String()

	handler := NewGroupHandler(mockService)

	t.Run("Delete group successfully", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/groups/"+groupID, nil)
		req.Header.Set("X-User-ID", user1ID) // Authorized user (member)
		
		// Set up mux vars
		vars := map[string]string{"id": groupID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.DeleteGroup(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		// Verify group was deleted
		_, err := mockService.GetGroupByID(groupID)
		if err == nil {
			t.Error("Expected group to be deleted")
		}
	})

	t.Run("Delete group unauthorized", func(t *testing.T) {
		// Create another group
		group2, _ := mockService.CreateGroup("Test Group 2", []string{user1ID, user2ID})
		group2ID := group2.ID.String()

		req := httptest.NewRequest("DELETE", "/groups/"+group2ID, nil)
		req.Header.Set("X-User-ID", user3ID) // Not a member (unauthorized)
		
		// Set up mux vars
		vars := map[string]string{"id": group2ID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.DeleteGroup(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("Delete non-existent group", func(t *testing.T) {
		nonExistentID := uuid.New().String()

		req := httptest.NewRequest("DELETE", "/groups/"+nonExistentID, nil)
		req.Header.Set("X-User-ID", user1ID)
		
		// Set up mux vars
		vars := map[string]string{"id": nonExistentID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.DeleteGroup(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestGroupHandler_RemoveUserFromGroup(t *testing.T) {
	user1ID := uuid.New().String()
	user2ID := uuid.New().String()
	user3ID := uuid.New().String()

	mockService := &mockGroupService{
		groups: make(map[string]*entity.Group),
	}

	// Create a test group
	group, _ := mockService.CreateGroup("Test Group", []string{user1ID, user2ID, user3ID})
	groupID := group.ID.String()

	handler := NewGroupHandler(mockService)

	t.Run("Remove user from group successfully", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/groups/"+groupID+"/users/"+user3ID, nil)
		req.Header.Set("X-User-ID", user1ID) // Authorized user (member)
		
		// Set up mux vars
		vars := map[string]string{"id": groupID, "user_id": user3ID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.RemoveUserFromGroup(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		// Verify user was removed
		updatedGroup, _ := mockService.GetGroupByID(groupID)
		found := false
		for _, uid := range updatedGroup.UserIDs {
			if uid.String() == user3ID {
				found = true
				break
			}
		}
		if found {
			t.Error("Expected user to be removed from group")
		}
	})

	t.Run("Remove user unauthorized", func(t *testing.T) {
		// Create another group
		group2, _ := mockService.CreateGroup("Test Group 2", []string{user1ID, user2ID})
		group2ID := group2.ID.String()

		req := httptest.NewRequest("DELETE", "/groups/"+group2ID+"/users/"+user2ID, nil)
		req.Header.Set("X-User-ID", user3ID) // Not a member (unauthorized)
		
		// Set up mux vars
		vars := map[string]string{"id": group2ID, "user_id": user2ID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.RemoveUserFromGroup(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("Remove user from non-existent group", func(t *testing.T) {
		nonExistentID := uuid.New().String()

		req := httptest.NewRequest("DELETE", "/groups/"+nonExistentID+"/users/"+user1ID, nil)
		req.Header.Set("X-User-ID", user1ID)
		
		// Set up mux vars
		vars := map[string]string{"id": nonExistentID, "user_id": user1ID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.RemoveUserFromGroup(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("Remove user not in group", func(t *testing.T) {
		// Create a group without user3
		group3, _ := mockService.CreateGroup("Test Group 3", []string{user1ID, user2ID})
		group3ID := group3.ID.String()

		req := httptest.NewRequest("DELETE", "/groups/"+group3ID+"/users/"+user3ID, nil)
		req.Header.Set("X-User-ID", user1ID)
		
		// Set up mux vars
		vars := map[string]string{"id": group3ID, "user_id": user3ID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.RemoveUserFromGroup(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

