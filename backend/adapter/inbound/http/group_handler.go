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
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	UserIDs []string `json:"user_ids"`
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
		ID:      group.ID.String(),
		Name:    group.Name,
		UserIDs: userIDs,
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
		ID:      group.ID.String(),
		Name:    group.Name,
		UserIDs: userIDs,
	})
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

