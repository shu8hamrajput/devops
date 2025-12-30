package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"splitwise/domain/entity"
)

// mockExpenseService is a mock implementation of ExpenseService for testing
type mockExpenseService struct {
	expenses map[string]*entity.Expense
}

func (m *mockExpenseService) CreateExpense(description string, amount float64, paidBy, groupID string) (*entity.Expense, error) {
	paidByUUID, _ := uuid.Parse(paidBy)
	groupUUID, _ := uuid.Parse(groupID)
	expense := entity.NewExpense(description, amount, paidByUUID, groupUUID)
	m.expenses[expense.ID.String()] = expense
	return expense, nil
}

func (m *mockExpenseService) CreateExpenseWithSplit(description string, amount float64, paidBy, groupID string, shareType entity.ShareType, userShares map[string]float64) (*entity.Expense, error) {
	// For testing, just create expense normally
	return m.CreateExpense(description, amount, paidBy, groupID)
}

func (m *mockExpenseService) CreateUserToUserExpense(description string, amount float64, paidBy, owedBy string) (*entity.Expense, error) {
	paidByUUID, _ := uuid.Parse(paidBy)
	owedByUUID, _ := uuid.Parse(owedBy)
	expense := entity.NewUserToUserExpense(description, amount, paidByUUID, owedByUUID)
	m.expenses[expense.ID.String()] = expense
	return expense, nil
}

func (m *mockExpenseService) GetExpenseByID(id string) (*entity.Expense, error) {
	expense, exists := m.expenses[id]
	if !exists {
		return nil, fmt.Errorf("expense not found")
	}
	return expense, nil
}

func (m *mockExpenseService) GetExpensesByGroup(groupID string) ([]*entity.Expense, error) {
	return []*entity.Expense{}, nil
}

func (m *mockExpenseService) GetExpensesByUser(userID string) ([]*entity.Expense, error) {
	return []*entity.Expense{}, nil
}

func (m *mockExpenseService) UpdateExpense(id, description string, amount float64, paidBy, groupID, owedBy string) (*entity.Expense, error) {
	expense, exists := m.expenses[id]
	if !exists {
		return nil, fmt.Errorf("expense not found")
	}
	expense.Description = description
	expense.Amount = amount
	paidByUUID, _ := uuid.Parse(paidBy)
	expense.PaidBy = paidByUUID
	expense.UpdatedAt = time.Now()
	if groupID != "" {
		groupUUID, _ := uuid.Parse(groupID)
		expense.GroupID = &groupUUID
		expense.OwedBy = nil
	} else if owedBy != "" {
		owedByUUID, _ := uuid.Parse(owedBy)
		expense.OwedBy = &owedByUUID
		expense.GroupID = nil
	}
	return expense, nil
}

func (m *mockExpenseService) DeleteExpense(id string) error {
	if _, exists := m.expenses[id]; !exists {
		return fmt.Errorf("expense not found")
	}
	delete(m.expenses, id)
	return nil
}

func TestExpenseHandler_UpdateExpense(t *testing.T) {
	user1ID := uuid.New().String()
	user2ID := uuid.New().String()
	groupID := uuid.New().String()

	mockService := &mockExpenseService{
		expenses: make(map[string]*entity.Expense),
	}

	// Create a test expense
	expense, _ := mockService.CreateExpense("Test Expense", 100.0, user1ID, groupID)
	expenseID := expense.ID.String()

	handler := NewExpenseHandler(mockService)

	t.Run("Update expense successfully", func(t *testing.T) {
		reqBody := UpdateExpenseRequest{
			Description: "Updated Expense",
			Amount:      150.0,
			PaidBy:      user1ID,
			GroupID:     groupID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/expenses/"+expenseID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", user1ID) // Authorized user
		
		// Set up mux vars
		vars := map[string]string{"id": expenseID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.UpdateExpense(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		var response ExpenseResponse
		if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Description != "Updated Expense" {
			t.Errorf("Expected description 'Updated Expense', got '%s'", response.Description)
		}

		if response.Amount != 150.0 {
			t.Errorf("Expected amount 150.0, got %f", response.Amount)
		}
	})

	t.Run("Update expense unauthorized", func(t *testing.T) {
		reqBody := UpdateExpenseRequest{
			Description: "Updated Expense",
			Amount:      150.0,
			PaidBy:      user2ID,
			GroupID:     groupID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/expenses/"+expenseID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", user2ID) // Different user (unauthorized)
		
		// Set up mux vars
		vars := map[string]string{"id": expenseID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.UpdateExpense(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("Update non-existent expense", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		reqBody := UpdateExpenseRequest{
			Description: "Updated Expense",
			Amount:      150.0,
			PaidBy:      user1ID,
			GroupID:     groupID,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest("PUT", "/expenses/"+nonExistentID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", user1ID)
		
		// Set up mux vars
		vars := map[string]string{"id": nonExistentID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.UpdateExpense(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

func TestExpenseHandler_DeleteExpense(t *testing.T) {
	user1ID := uuid.New().String()
	user2ID := uuid.New().String()
	groupID := uuid.New().String()

	mockService := &mockExpenseService{
		expenses: make(map[string]*entity.Expense),
	}

	// Create a test expense
	expense, _ := mockService.CreateExpense("Test Expense", 100.0, user1ID, groupID)
	expenseID := expense.ID.String()

	handler := NewExpenseHandler(mockService)

	t.Run("Delete expense successfully", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/expenses/"+expenseID, nil)
		req.Header.Set("X-User-ID", user1ID) // Authorized user
		
		// Set up mux vars
		vars := map[string]string{"id": expenseID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.DeleteExpense(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}

		// Verify expense was deleted
		_, err := mockService.GetExpenseByID(expenseID)
		if err == nil {
			t.Error("Expected expense to be deleted")
		}
	})

	t.Run("Delete expense unauthorized", func(t *testing.T) {
		// Create another expense
		expense2, _ := mockService.CreateExpense("Test Expense 2", 200.0, user1ID, groupID)
		expense2ID := expense2.ID.String()

		req := httptest.NewRequest("DELETE", "/expenses/"+expense2ID, nil)
		req.Header.Set("X-User-ID", user2ID) // Different user (unauthorized)
		
		// Set up mux vars
		vars := map[string]string{"id": expense2ID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.DeleteExpense(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status code %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("Delete non-existent expense", func(t *testing.T) {
		nonExistentID := uuid.New().String()

		req := httptest.NewRequest("DELETE", "/expenses/"+nonExistentID, nil)
		req.Header.Set("X-User-ID", user1ID)
		
		// Set up mux vars
		vars := map[string]string{"id": nonExistentID}
		req = mux.SetURLVars(req, vars)
		
		w := httptest.NewRecorder()

		handler.DeleteExpense(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
		}
	})
}

