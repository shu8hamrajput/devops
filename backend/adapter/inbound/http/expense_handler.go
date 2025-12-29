package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
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
	GroupID     string  `json:"group_id"`
}

// ExpenseResponse represents the response for expense operations
type ExpenseResponse struct {
	ID          string  `json:"id"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	PaidBy      string  `json:"paid_by"`
	GroupID     string  `json:"group_id"`
	CreatedAt   string  `json:"created_at"`
}

// CreateExpense handles POST /expenses
func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	var req CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	expense, err := h.expenseService.CreateExpense(req.Description, req.Amount, req.PaidBy, req.GroupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ExpenseResponse{
		ID:          expense.ID.String(),
		Description: expense.Description,
		Amount:      expense.Amount,
		PaidBy:      expense.PaidBy.String(),
		GroupID:     expense.GroupID.String(),
		CreatedAt:   expense.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ExpenseResponse{
		ID:          expense.ID.String(),
		Description: expense.Description,
		Amount:      expense.Amount,
		PaidBy:      expense.PaidBy.String(),
		GroupID:     expense.GroupID.String(),
		CreatedAt:   expense.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
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
		responses = append(responses, ExpenseResponse{
			ID:          expense.ID.String(),
			Description: expense.Description,
			Amount:      expense.Amount,
			PaidBy:      expense.PaidBy.String(),
			GroupID:     expense.GroupID.String(),
			CreatedAt:   expense.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

