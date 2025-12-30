package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"splitwise/domain/entity"
)

func setupTestDBForExpense(t *testing.T) *sql.DB {
	// Try to connect to test database first, fallback to main database
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=splitwise_test sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		// Fallback to main database
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=splitwise sslmode=disable"
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			t.Skipf("Skipping test: could not connect to database: %v", err)
			return nil
		}
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: could not ping database: %v", err)
		return nil
	}

	return db
}

func createTestUserForExpense(t *testing.T, db *sql.DB, name, email string) *entity.User {
	user := entity.NewUser(name, email)
	query := `INSERT INTO users (id, name, email, password_hash, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(query, user.ID, user.Name, user.Email, "hashed_password", user.CreatedAt, user.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

func createTestGroupForExpense(t *testing.T, db *sql.DB, name string, userIDs []uuid.UUID) *entity.Group {
	group := entity.NewGroup(name, userIDs)
	
	// Insert group
	query := `INSERT INTO groups (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, group.ID, group.Name, group.CreatedAt, group.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create test group: %v", err)
	}

	// Insert group_users relationships
	for _, userID := range userIDs {
		_, err = db.Exec(`INSERT INTO group_users (group_id, user_id) VALUES ($1, $2)`, group.ID, userID)
		if err != nil {
			t.Fatalf("Failed to add user to group: %v", err)
		}
	}

	return group
}

func cleanupTestExpenseData(t *testing.T, db *sql.DB) {
	// Clean up in reverse order of dependencies
	db.Exec("DELETE FROM expenses")
	db.Exec("DELETE FROM group_users")
	db.Exec("DELETE FROM groups")
	db.Exec("DELETE FROM users")
}

func TestPostgresExpenseRepository_Update(t *testing.T) {
	db := setupTestDBForExpense(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestExpenseData(t, db)

	repo := NewPostgresExpenseRepository(db)

	// Create test users
	user1 := createTestUserForExpense(t, db, "User 1", "user1@test.com")
	user2 := createTestUserForExpense(t, db, "User 2", "user2@test.com")
	group := createTestGroupForExpense(t, db, "Test Group", []uuid.UUID{user1.ID, user2.ID})

	t.Run("Update expense description and amount", func(t *testing.T) {
		expense := entity.NewExpense("Original Description", 100.0, user1.ID, group.ID)
		err := repo.Create(expense)
		if err != nil {
			t.Fatalf("Failed to create expense: %v", err)
		}

		// Update expense
		expense.Description = "Updated Description"
		expense.Amount = 150.0
		expense.UpdatedAt = time.Now()

		err = repo.Update(expense)
		if err != nil {
			t.Fatalf("Failed to update expense: %v", err)
		}

		// Verify update
		retrieved, err := repo.FindByID(expense.ID.String())
		if err != nil {
			t.Fatalf("Failed to retrieve updated expense: %v", err)
		}

		if retrieved.Description != "Updated Description" {
			t.Errorf("Expected description 'Updated Description', got '%s'", retrieved.Description)
		}
		if retrieved.Amount != 150.0 {
			t.Errorf("Expected amount 150.0, got %f", retrieved.Amount)
		}
	})

	t.Run("Update expense with different paid_by user", func(t *testing.T) {
		expense := entity.NewExpense("Test Expense", 50.0, user1.ID, group.ID)
		err := repo.Create(expense)
		if err != nil {
			t.Fatalf("Failed to create expense: %v", err)
		}

		// Update paid_by
		expense.PaidBy = user2.ID
		expense.UpdatedAt = time.Now()

		err = repo.Update(expense)
		if err != nil {
			t.Fatalf("Failed to update expense: %v", err)
		}

		// Verify update
		retrieved, err := repo.FindByID(expense.ID.String())
		if err != nil {
			t.Fatalf("Failed to retrieve updated expense: %v", err)
		}

		if retrieved.PaidBy != user2.ID {
			t.Errorf("Expected paid_by to be %s, got %s", user2.ID, retrieved.PaidBy)
		}
	})

	t.Run("Update non-existent expense", func(t *testing.T) {
		nonExistentExpense := entity.NewExpense("Non-existent", 100.0, user1.ID, group.ID)
		nonExistentExpense.ID = uuid.New()
		nonExistentExpense.UpdatedAt = time.Now()

		err := repo.Update(nonExistentExpense)
		if err == nil {
			t.Error("Expected error when updating non-existent expense")
		}
	})

	t.Run("Update soft-deleted expense", func(t *testing.T) {
		expense := entity.NewExpense("To be deleted", 100.0, user1.ID, group.ID)
		err := repo.Create(expense)
		if err != nil {
			t.Fatalf("Failed to create expense: %v", err)
		}

		// Soft delete the expense
		err = repo.Delete(expense.ID.String())
		if err != nil {
			t.Fatalf("Failed to delete expense: %v", err)
		}

		// Try to update soft-deleted expense
		expense.Description = "Updated after delete"
		expense.UpdatedAt = time.Now()
		err = repo.Update(expense)
		if err == nil {
			t.Error("Expected error when updating soft-deleted expense")
		}
	})
}

func TestPostgresExpenseRepository_Delete(t *testing.T) {
	db := setupTestDBForExpense(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestExpenseData(t, db)

	repo := NewPostgresExpenseRepository(db)

	// Create test users
	user1 := createTestUserForExpense(t, db, "User 1", "user1@test.com")
	user2 := createTestUserForExpense(t, db, "User 2", "user2@test.com")
	group := createTestGroupForExpense(t, db, "Test Group", []uuid.UUID{user1.ID, user2.ID})

	t.Run("Delete expense successfully", func(t *testing.T) {
		expense := entity.NewExpense("To be deleted", 100.0, user1.ID, group.ID)
		err := repo.Create(expense)
		if err != nil {
			t.Fatalf("Failed to create expense: %v", err)
		}

		// Delete expense
		err = repo.Delete(expense.ID.String())
		if err != nil {
			t.Fatalf("Failed to delete expense: %v", err)
		}

		// Verify expense is soft-deleted (should not be found)
		_, err = repo.FindByID(expense.ID.String())
		if err == nil {
			t.Error("Expected error when finding soft-deleted expense")
		}
	})

	t.Run("Delete non-existent expense", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		err := repo.Delete(nonExistentID)
		if err == nil {
			t.Error("Expected error when deleting non-existent expense")
		}
	})

	t.Run("Delete already deleted expense", func(t *testing.T) {
		expense := entity.NewExpense("To be deleted twice", 100.0, user1.ID, group.ID)
		err := repo.Create(expense)
		if err != nil {
			t.Fatalf("Failed to create expense: %v", err)
		}

		// Delete expense first time
		err = repo.Delete(expense.ID.String())
		if err != nil {
			t.Fatalf("Failed to delete expense first time: %v", err)
		}

		// Try to delete again
		err = repo.Delete(expense.ID.String())
		if err == nil {
			t.Error("Expected error when deleting already deleted expense")
		}
	})

	t.Run("Delete expense with invalid ID", func(t *testing.T) {
		err := repo.Delete("invalid-uuid")
		if err == nil {
			t.Error("Expected error when deleting expense with invalid ID")
		}
	})

	t.Run("Verify deleted expense not in FindByGroupID", func(t *testing.T) {
		expense1 := entity.NewExpense("Expense 1", 100.0, user1.ID, group.ID)
		expense2 := entity.NewExpense("Expense 2", 200.0, user2.ID, group.ID)
		
		err := repo.Create(expense1)
		if err != nil {
			t.Fatalf("Failed to create expense1: %v", err)
		}
		err = repo.Create(expense2)
		if err != nil {
			t.Fatalf("Failed to create expense2: %v", err)
		}

		// Delete expense1
		err = repo.Delete(expense1.ID.String())
		if err != nil {
			t.Fatalf("Failed to delete expense1: %v", err)
		}

		// Find expenses by group - should only return expense2
		expenses, err := repo.FindByGroupID(group.ID.String())
		if err != nil {
			t.Fatalf("Failed to find expenses by group: %v", err)
		}

		if len(expenses) != 1 {
			t.Errorf("Expected 1 expense, got %d", len(expenses))
		}
		if expenses[0].ID != expense2.ID {
			t.Errorf("Expected expense2, got expense with ID %s", expenses[0].ID)
		}
	})
}

func TestPostgresExpenseRepository_Integration(t *testing.T) {
	db := setupTestDBForExpense(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestExpenseData(t, db)

	repo := NewPostgresExpenseRepository(db)

	// Create test users
	user1 := createTestUserForExpense(t, db, "User 1", "user1@test.com")
	user2 := createTestUserForExpense(t, db, "User 2", "user2@test.com")
	group := createTestGroupForExpense(t, db, "Test Group", []uuid.UUID{user1.ID, user2.ID})

	t.Run("Full CRUD workflow", func(t *testing.T) {
		// Create
		expense := entity.NewExpense("Integration Test Expense", 100.0, user1.ID, group.ID)
		err := repo.Create(expense)
		if err != nil {
			t.Fatalf("Failed to create expense: %v", err)
		}

		// Read
		retrieved, err := repo.FindByID(expense.ID.String())
		if err != nil {
			t.Fatalf("Failed to find expense: %v", err)
		}
		if retrieved.Description != "Integration Test Expense" {
			t.Errorf("Expense description mismatch")
		}

		// Update
		retrieved.Description = "Updated Integration Test Expense"
		retrieved.Amount = 150.0
		retrieved.UpdatedAt = time.Now()
		err = repo.Update(retrieved)
		if err != nil {
			t.Fatalf("Failed to update expense: %v", err)
		}

		// Verify update
		updated, err := repo.FindByID(expense.ID.String())
		if err != nil {
			t.Fatalf("Failed to find updated expense: %v", err)
		}
		if updated.Description != "Updated Integration Test Expense" {
			t.Errorf("Expense description not updated correctly")
		}
		if updated.Amount != 150.0 {
			t.Errorf("Expense amount not updated correctly")
		}

		// Find by group
		expenses, err := repo.FindByGroupID(group.ID.String())
		if err != nil {
			t.Fatalf("Failed to find expenses by group: %v", err)
		}
		found := false
		for _, e := range expenses {
			if e.ID == expense.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expense not found in group's expenses")
		}

		// Delete (soft delete)
		err = repo.Delete(expense.ID.String())
		if err != nil {
			t.Fatalf("Failed to delete expense: %v", err)
		}

		// Verify delete
		_, err = repo.FindByID(expense.ID.String())
		if err == nil {
			t.Error("Expected error when finding deleted expense")
		}
	})
}

