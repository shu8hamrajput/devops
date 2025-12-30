package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"splitwise/domain/entity"
	"splitwise/domain/port/inbound"
)

// ExpenseHandler handles HTTP requests for expense operations
type ExpenseHandler struct {
	expenseService inbound.ExpenseService
}

// NewExpenseHandler creates a new expense HTTP handler
func NewExpenseHandler(expenseService inbound.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{
		expenseService: expenseService,
	}
}

// CreateExpenseRequest represents the request body for creating an expense
type CreateExpenseRequest struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	PaidBy      string  `json:"paid_by"`
	GroupID     string  `json:"group_id,omitempty"`
	OwedBy      string  `json:"owed_by,omitempty"` // For user-to-user expenses
}

// ExpenseResponse represents the response for expense operations
type ExpenseResponse struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	PaidBy      string  `json:"paid_by"`
	GroupID     *string `json:"group_id,omitempty"`
	OwedBy      *string `json:"owed_by,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreateExpense handles POST /expenses
func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	var req CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var expense *entity.Expense
	var err error

	// Determine if it's a group expense or user-to-user expense
	if req.GroupID != "" {
		expense, err = h.expenseService.CreateExpense(req.Description, req.Amount, req.PaidBy, req.GroupID)
	} else if req.OwedBy != "" {
		expense, err = h.expenseService.CreateUserToUserExpense(req.Description, req.Amount, req.PaidBy, req.OwedBy)
	} else {
		http.Error(w, "Either group_id or owed_by must be provided", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	groupID := ""
	if expense.GroupID != nil {
		groupID = expense.GroupID.String()
	}
	owedBy := ""
	if expense.OwedBy != nil {
		owedBy = expense.OwedBy.String()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := ExpenseResponse{
		ID:          expense.ID.String(),
		Description: expense.Description,
		Amount:      expense.Amount,
		PaidBy:      expense.PaidBy.String(),
		CreatedAt:   expense.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   expense.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if groupID != "" {
		response.GroupID = &groupID
	}
	if owedBy != "" {
		response.OwedBy = &owedBy
	}
	json.NewEncoder(w).Encode(response)
}

// GetExpense handles GET /expenses/{id}
func (h *ExpenseHandler) GetExpense(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	expense, err := h.expenseService.GetExpenseByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	groupID := ""
	if expense.GroupID != nil {
		groupID = expense.GroupID.String()
	}
	owedBy := ""
	if expense.OwedBy != nil {
		owedBy = expense.OwedBy.String()
	}

	response := ExpenseResponse{
		ID:          expense.ID.String(),
		Description: expense.Description,
		Amount:      expense.Amount,
		PaidBy:      expense.PaidBy.String(),
		CreatedAt:   expense.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   expense.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if groupID != "" {
		response.GroupID = &groupID
	}
	if owedBy != "" {
		response.OwedBy = &owedBy
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetExpensesByGroup handles GET /groups/{group_id}/expenses
func (h *ExpenseHandler) GetExpensesByGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupID := vars["group_id"]

	expenses, err := h.expenseService.GetExpensesByGroup(groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]ExpenseResponse, 0, len(expenses))
	for _, expense := range expenses {
		groupID := ""
		if expense.GroupID != nil {
			groupID = expense.GroupID.String()
		}
		owedBy := ""
		if expense.OwedBy != nil {
			owedBy = expense.OwedBy.String()
		}

		response := ExpenseResponse{
			ID:          expense.ID.String(),
			Description: expense.Description,
			Amount:      expense.Amount,
			PaidBy:      expense.PaidBy.String(),
			CreatedAt:   expense.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   expense.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if groupID != "" {
			response.GroupID = &groupID
		}
		if owedBy != "" {
			response.OwedBy = &owedBy
		}
		responses = append(responses, response)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// GetExpensesByUser handles GET /users/{user_id}/expenses
func (h *ExpenseHandler) GetExpensesByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]

	expenses, err := h.expenseService.GetExpensesByUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responses := make([]ExpenseResponse, 0, len(expenses))
	for _, expense := range expenses {
		groupID := ""
		if expense.GroupID != nil {
			groupID = expense.GroupID.String()
		}
		owedBy := ""
		if expense.OwedBy != nil {
			owedBy = expense.OwedBy.String()
		}

		response := ExpenseResponse{
			ID:          expense.ID.String(),
			Description: expense.Description,
			Amount:      expense.Amount,
			PaidBy:      expense.PaidBy.String(),
			CreatedAt:   expense.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   expense.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if groupID != "" {
			response.GroupID = &groupID
		}
		if owedBy != "" {
			response.OwedBy = &owedBy
		}
		responses = append(responses, response)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// UpdateExpenseRequest represents the request body for updating an expense
type UpdateExpenseRequest struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	PaidBy      string  `json:"paid_by"`
	GroupID     string  `json:"group_id,omitempty"`
	OwedBy      string  `json:"owed_by,omitempty"`
}

// UpdateExpense handles PUT /expenses/{id}
func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get current user ID from header (set by auth middleware)
	currentUserID := r.Header.Get("X-User-ID")
	if currentUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get existing expense to check authorization
	existingExpense, err := h.expenseService.GetExpenseByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check: only the user who paid can update the expense
	if existingExpense.PaidBy.String() != currentUserID {
		http.Error(w, "Forbidden: only the user who paid can update this expense", http.StatusForbidden)
		return
	}

	var req UpdateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	expense, err := h.expenseService.UpdateExpense(id, req.Description, req.Amount, req.PaidBy, req.GroupID, req.OwedBy)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	groupID := ""
	if expense.GroupID != nil {
		groupID = expense.GroupID.String()
	}
	owedBy := ""
	if expense.OwedBy != nil {
		owedBy = expense.OwedBy.String()
	}

	w.Header().Set("Content-Type", "application/json")
	response := ExpenseResponse{
		ID:          expense.ID.String(),
		Description: expense.Description,
		Amount:      expense.Amount,
		PaidBy:      expense.PaidBy.String(),
		CreatedAt:   expense.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   expense.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if groupID != "" {
		response.GroupID = &groupID
	}
	if owedBy != "" {
		response.OwedBy = &owedBy
	}
	json.NewEncoder(w).Encode(response)
}

// DeleteExpense handles DELETE /expenses/{id}
func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get current user ID from header (set by auth middleware)
	currentUserID := r.Header.Get("X-User-ID")
	if currentUserID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get existing expense to check authorization
	existingExpense, err := h.expenseService.GetExpenseByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Authorization check: only the user who paid can delete the expense
	if existingExpense.PaidBy.String() != currentUserID {
		http.Error(w, "Forbidden: only the user who paid can delete this expense", http.StatusForbidden)
		return
	}

	if err := h.expenseService.DeleteExpense(id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Expense deleted successfully"})
}

