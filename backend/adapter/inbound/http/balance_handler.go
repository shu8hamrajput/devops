package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"splitwise/domain/port/inbound"
)

// BalanceHandler handles HTTP requests for balance operations
type BalanceHandler struct {
	balanceService inbound.BalanceService
}

// NewBalanceHandler creates a new balance HTTP handler
func NewBalanceHandler(balanceService inbound.BalanceService) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
	}
}

// GetGroupBalance handles GET /groups/{id}/balance
func (h *BalanceHandler) GetGroupBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["id"]

	groupBalance, err := h.balanceService.CalculateGroupBalance(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groupBalance)
}

// GetUserBalance handles GET /users/{id}/balance
func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	userBalance, err := h.balanceService.CalculateUserBalance(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userBalance)
}

// GetUserToUserBalance handles GET /users/{id}/balance/{other_user_id}
func (h *BalanceHandler) GetUserToUserBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]
	otherUserID := vars["other_user_id"]

	balance, err := h.balanceService.CalculateUserToUserBalance(userID, otherUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"user_id":        userID,
		"other_user_id":  otherUserID,
		"balance":        balance,
		"description":    getBalanceDescription(userID, otherUserID, balance),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getBalanceDescription returns a human-readable description of the balance
func getBalanceDescription(userID1, userID2 string, balance float64) string {
	if balance > 0.01 {
		return "User owes other user"
	} else if balance < -0.01 {
		return "Other user owes user"
	}
	return "No balance"
}

