package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"splitwise/domain/port/inbound"
)

// GroupHandler handles HTTP requests for group operations
type GroupHandler struct {
	groupService inbound.GroupService
}

// NewGroupHandler creates a new group HTTP handler
func NewGroupHandler(groupService inbound.GroupService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

// CreateGroupRequest represents the request body for creating a group
type CreateGroupRequest struct {
	Name    string   `json:"name"`
	UserIDs []string `json:"user_ids"`
}

// GroupResponse represents the response for group operations
type GroupResponse struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	UserIDs   []string `json:"user_ids"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

// CreateGroup handles POST /groups
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group, err := h.groupService.CreateGroup(req.Name, req.UserIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userIDs := make([]string, 0, len(group.UserIDs))
	for _, id := range group.UserIDs {
		userIDs = append(userIDs, id.String())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(GroupResponse{
		ID:        group.ID.String(),
		Name:      group.Name,
		UserIDs:   userIDs,
		CreatedAt: group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetGroup handles GET /groups/{id}
func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	group, err := h.groupService.GetGroupByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	userIDs := make([]string, 0, len(group.UserIDs))
	for _, userID := range group.UserIDs {
		userIDs = append(userIDs, userID.String())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GroupResponse{
		ID:        group.ID.String(),
		Name:      group.Name,
		UserIDs:   userIDs,
		CreatedAt: group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// GetAllGroups handles GET /groups
func (h *GroupHandler) GetAllGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.groupService.GetAllGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]GroupResponse, 0, len(groups))
	for _, group := range groups {
		userIDs := make([]string, 0, len(group.UserIDs))
		for _, id := range group.UserIDs {
			userIDs = append(userIDs, id.String())
		}
		responses = append(responses, GroupResponse{
			ID:        group.ID.String(),
			Name:      group.Name,
			UserIDs:   userIDs,
			CreatedAt: group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// GetGroupsByUser handles GET /users/{user_id}/groups
func (h *GroupHandler) GetGroupsByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	groups, err := h.groupService.GetGroupsByUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]GroupResponse, 0, len(groups))
	for _, group := range groups {
		userIDs := make([]string, 0, len(group.UserIDs))
		for _, id := range group.UserIDs {
			userIDs = append(userIDs, id.String())
		}
		responses = append(responses, GroupResponse{
			ID:        group.ID.String(),
			Name:      group.Name,
			UserIDs:   userIDs,
			CreatedAt: group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// AddUserToGroup handles POST /groups/{id}/users
func (h *GroupHandler) AddUserToGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.groupService.AddUserToGroup(groupID, req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User added to group successfully"})
}

// UpdateGroupRequest represents the request body for updating a group
type UpdateGroupRequest struct {
	Name    string   `json:"name"`
	UserIDs []string `json:"user_ids"`
}

// UpdateGroup handles PUT /groups/{id}
func (h *GroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get current user ID from header (set by auth middleware)
	currentUserID := r.Header.Get("X-User-ID")
	if currentUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get existing group to check authorization
	existingGroup, err := h.groupService.GetGroupByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check: user must be a member of the group
	isMember := false
	for _, userID := range existingGroup.UserIDs {
		if userID.String() == currentUserID {
			isMember = true
			break
		}
	}
	if !isMember {
		http.Error(w, "Forbidden: only group members can update the group", http.StatusForbidden)
		return
	}

	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group, err := h.groupService.UpdateGroup(id, req.Name, req.UserIDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userIDs := make([]string, 0, len(group.UserIDs))
	for _, id := range group.UserIDs {
		userIDs = append(userIDs, id.String())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GroupResponse{
		ID:        group.ID.String(),
		Name:      group.Name,
		UserIDs:   userIDs,
		CreatedAt: group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// DeleteGroup handles DELETE /groups/{id}
func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get current user ID from header (set by auth middleware)
	currentUserID := r.Header.Get("X-User-ID")
	if currentUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get existing group to check authorization
	existingGroup, err := h.groupService.GetGroupByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check: user must be a member of the group
	isMember := false
	for _, userID := range existingGroup.UserIDs {
		if userID.String() == currentUserID {
			isMember = true
			break
		}
	}
	if !isMember {
		http.Error(w, "Forbidden: only group members can delete the group", http.StatusForbidden)
		return
	}

	if err := h.groupService.DeleteGroup(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Group deleted successfully"})
}

// RemoveUserFromGroup handles DELETE /groups/{id}/users/{user_id}
func (h *GroupHandler) RemoveUserFromGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]
	userID := vars["user_id"]

	// Get current user ID from header (set by auth middleware)
	currentUserID := r.Header.Get("X-User-ID")
	if currentUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get existing group to check authorization
	existingGroup, err := h.groupService.GetGroupByID(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check: user must be a member of the group
	isMember := false
	for _, uid := range existingGroup.UserIDs {
		if uid.String() == currentUserID {
			isMember = true
			break
		}
	}
	if !isMember {
		http.Error(w, "Forbidden: only group members can remove users from the group", http.StatusForbidden)
		return
	}

	if err := h.groupService.RemoveUserFromGroup(groupID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User removed from group successfully"})
}

