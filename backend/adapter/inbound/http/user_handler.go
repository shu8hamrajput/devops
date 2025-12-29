package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"splitwise/domain/port/inbound"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService inbound.UserService
}

// NewUserHandler creates a new user HTTP handler
func NewUserHandler(userService inbound.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserResponse represents the response for user operations
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.userService.CreateUser(req.Name, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(UserResponse{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
	})
}

// GetUser handles GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(UserResponse{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
	})
}

// GetAllUsers handles GET /users
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, UserResponse{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

