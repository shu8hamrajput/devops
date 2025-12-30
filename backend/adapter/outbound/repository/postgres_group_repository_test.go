package repository

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"splitwise/domain/entity"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	// Use test database or create a connection string
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "postgres"
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		// Try test database first, fall back to main database
		testDB := "splitwise_test"
		dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + testDB + " sslmode=" + sslmode
		testConn, err := sql.Open("postgres", dsn)
		if err == nil {
			if err := testConn.Ping(); err == nil {
				testConn.Close()
				dbname = testDB
			} else {
				testConn.Close()
			}
		}
		if dbname == "" {
			dbname = "splitwise" // Fall back to main database
		}
	}

	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=" + sslmode

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Skipf("Skipping test: failed to connect to test database: %v", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test: failed to ping test database: %v", err)
		return nil
	}

	// Create test tables
	if err := CreateTables(db); err != nil {
		t.Fatalf("Failed to create test tables: %v", err)
	}

	return db
}

// cleanupTestData removes test data from the database
func cleanupTestData(t *testing.T, db *sql.DB) {
	queries := []string{
		"DELETE FROM group_users",
		"DELETE FROM groups",
		"DELETE FROM expenses",
		"DELETE FROM users",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			t.Logf("Warning: failed to cleanup test data: %v", err)
		}
	}
}

// createTestUser creates a test user in the database
func createTestUser(t *testing.T, db *sql.DB, name, email string) *entity.User {
	user := &entity.User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: "test_hash",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	query := `
		INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.Exec(query, user.ID, user.Name, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

func TestPostgresGroupRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestData(t, db)

	repo := NewPostgresGroupRepository(db)

	// Create test users
	user1 := createTestUser(t, db, "User 1", "user1@test.com")
	user2 := createTestUser(t, db, "User 2", "user2@test.com")

	t.Run("Create group with users", func(t *testing.T) {
		group := entity.NewGroup("Test Group", []uuid.UUID{user1.ID, user2.ID})

		err := repo.Create(group)
		if err != nil {
			t.Fatalf("Failed to create group: %v", err)
		}

		// Verify group was created
		retrieved, err := repo.FindByID(group.ID.String())
		if err != nil {
			t.Fatalf("Failed to retrieve created group: %v", err)
		}

		if retrieved.Name != "Test Group" {
			t.Errorf("Expected name 'Test Group', got '%s'", retrieved.Name)
		}

		if len(retrieved.UserIDs) != 2 {
			t.Errorf("Expected 2 users, got %d", len(retrieved.UserIDs))
		}
	})

	t.Run("Create group with invalid user ID", func(t *testing.T) {
		invalidUserID := uuid.New()
		group := entity.NewGroup("Invalid Group", []uuid.UUID{invalidUserID})

		err := repo.Create(group)
		if err == nil {
			t.Error("Expected error when creating group with invalid user ID")
		}
	})

	t.Run("Create group with no users", func(t *testing.T) {
		group := entity.NewGroup("Empty Group", []uuid.UUID{})

		err := repo.Create(group)
		if err != nil {
			t.Fatalf("Failed to create group with no users: %v", err)
		}

		retrieved, err := repo.FindByID(group.ID.String())
		if err != nil {
			t.Fatalf("Failed to retrieve created group: %v", err)
		}

		if len(retrieved.UserIDs) != 0 {
			t.Errorf("Expected 0 users, got %d", len(retrieved.UserIDs))
		}
	})
}

func TestPostgresGroupRepository_FindByID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestData(t, db)

	repo := NewPostgresGroupRepository(db)

	// Create test users
	user1 := createTestUser(t, db, "User 1", "user1@test.com")
	user2 := createTestUser(t, db, "User 2", "user2@test.com")
	user3 := createTestUser(t, db, "User 3", "user3@test.com")

	t.Run("Find existing group", func(t *testing.T) {
		group := entity.NewGroup("Test Group", []uuid.UUID{user1.ID, user2.ID, user3.ID})
		err := repo.Create(group)
		if err != nil {
			t.Fatalf("Failed to create group: %v", err)
		}

		retrieved, err := repo.FindByID(group.ID.String())
		if err != nil {
			t.Fatalf("Failed to find group: %v", err)
		}

		if retrieved.ID != group.ID {
			t.Errorf("Expected ID %s, got %s", group.ID, retrieved.ID)
		}

		if retrieved.Name != "Test Group" {
			t.Errorf("Expected name 'Test Group', got '%s'", retrieved.Name)
		}

		if len(retrieved.UserIDs) != 3 {
			t.Errorf("Expected 3 users, got %d", len(retrieved.UserIDs))
		}
	})

	t.Run("Find non-existent group", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		_, err := repo.FindByID(nonExistentID)
		if err == nil {
			t.Error("Expected error when finding non-existent group")
		}
	})

	t.Run("Find group with invalid ID format", func(t *testing.T) {
		_, err := repo.FindByID("invalid-uuid")
		if err == nil {
			t.Error("Expected error when finding group with invalid ID")
		}
	})
}

func TestPostgresGroupRepository_FindByUserID(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestData(t, db)

	repo := NewPostgresGroupRepository(db)

	// Create test users
	user1 := createTestUser(t, db, "User 1", "user1@test.com")
	user2 := createTestUser(t, db, "User 2", "user2@test.com")
	user3 := createTestUser(t, db, "User 3", "user3@test.com")

	t.Run("Find groups for user", func(t *testing.T) {
		// Create groups
		group1 := entity.NewGroup("Group 1", []uuid.UUID{user1.ID, user2.ID})
		group2 := entity.NewGroup("Group 2", []uuid.UUID{user1.ID, user3.ID})
		group3 := entity.NewGroup("Group 3", []uuid.UUID{user2.ID, user3.ID})

		// Add delay to ensure different updated_at timestamps
		time.Sleep(10 * time.Millisecond)

		err := repo.Create(group1)
		if err != nil {
			t.Fatalf("Failed to create group1: %v", err)
		}

		time.Sleep(10 * time.Millisecond)
		err = repo.Create(group2)
		if err != nil {
			t.Fatalf("Failed to create group2: %v", err)
		}

		time.Sleep(10 * time.Millisecond)
		err = repo.Create(group3)
		if err != nil {
			t.Fatalf("Failed to create group3: %v", err)
		}

		// Find groups for user1
		groups, err := repo.FindByUserID(user1.ID.String())
		if err != nil {
			t.Fatalf("Failed to find groups: %v", err)
		}

		if len(groups) != 2 {
			t.Errorf("Expected 2 groups for user1, got %d", len(groups))
		}

		// Verify groups are sorted by updated_at DESC
		if len(groups) >= 2 {
			if !groups[0].UpdatedAt.After(groups[1].UpdatedAt) && !groups[0].UpdatedAt.Equal(groups[1].UpdatedAt) {
				t.Error("Groups should be sorted by updated_at DESC")
			}
		}
	})

	t.Run("Find groups for user with no groups", func(t *testing.T) {
		user4 := createTestUser(t, db, "User 4", "user4@test.com")
		groups, err := repo.FindByUserID(user4.ID.String())
		if err != nil {
			t.Fatalf("Failed to find groups: %v", err)
		}

		if len(groups) != 0 {
			t.Errorf("Expected 0 groups, got %d", len(groups))
		}
	})

	t.Run("Find groups with invalid user ID", func(t *testing.T) {
		_, err := repo.FindByUserID("invalid-uuid")
		if err == nil {
			t.Error("Expected error when finding groups with invalid user ID")
		}
	})
}

func TestPostgresGroupRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestData(t, db)

	repo := NewPostgresGroupRepository(db)

	// Create test users
	user1 := createTestUser(t, db, "User 1", "user1@test.com")
	user2 := createTestUser(t, db, "User 2", "user2@test.com")
	user3 := createTestUser(t, db, "User 3", "user3@test.com")

	t.Run("Update group name and users", func(t *testing.T) {
		group := entity.NewGroup("Original Name", []uuid.UUID{user1.ID, user2.ID})
		err := repo.Create(group)
		if err != nil {
			t.Fatalf("Failed to create group: %v", err)
		}

		// Update group
		group.Name = "Updated Name"
		group.UserIDs = []uuid.UUID{user1.ID, user2.ID, user3.ID}
		group.UpdateTimestamp()

		err = repo.Update(group)
		if err != nil {
			t.Fatalf("Failed to update group: %v", err)
		}

		// Verify update
		retrieved, err := repo.FindByID(group.ID.String())
		if err != nil {
			t.Fatalf("Failed to retrieve updated group: %v", err)
		}

		if retrieved.Name != "Updated Name" {
			t.Errorf("Expected name 'Updated Name', got '%s'", retrieved.Name)
		}

		if len(retrieved.UserIDs) != 3 {
			t.Errorf("Expected 3 users, got %d", len(retrieved.UserIDs))
		}
	})

	t.Run("Update group to remove all users", func(t *testing.T) {
		group := entity.NewGroup("Group with Users", []uuid.UUID{user1.ID, user2.ID})
		err := repo.Create(group)
		if err != nil {
			t.Fatalf("Failed to create group: %v", err)
		}

		// Update to remove all users
		group.UserIDs = []uuid.UUID{}
		group.UpdateTimestamp()

		err = repo.Update(group)
		if err != nil {
			t.Fatalf("Failed to update group: %v", err)
		}

		// Verify update
		retrieved, err := repo.FindByID(group.ID.String())
		if err != nil {
			t.Fatalf("Failed to retrieve updated group: %v", err)
		}

		if len(retrieved.UserIDs) != 0 {
			t.Errorf("Expected 0 users, got %d", len(retrieved.UserIDs))
		}
	})

	t.Run("Update non-existent group", func(t *testing.T) {
		nonExistentGroup := entity.NewGroup("Non-existent", []uuid.UUID{user1.ID})
		nonExistentGroup.ID = uuid.New()
		nonExistentGroup.UpdateTimestamp()

		err := repo.Update(nonExistentGroup)
		if err == nil {
			t.Error("Expected error when updating non-existent group")
		}
	})

	t.Run("Update group with invalid user ID", func(t *testing.T) {
		group := entity.NewGroup("Valid Group", []uuid.UUID{user1.ID})
		err := repo.Create(group)
		if err != nil {
			t.Fatalf("Failed to create group: %v", err)
		}

		// Try to update with invalid user
		invalidUserID := uuid.New()
		group.UserIDs = []uuid.UUID{invalidUserID}
		group.UpdateTimestamp()

		err = repo.Update(group)
		if err == nil {
			t.Error("Expected error when updating group with invalid user ID")
		}
	})
}

func TestPostgresGroupRepository_Integration(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()
	defer cleanupTestData(t, db)

	repo := NewPostgresGroupRepository(db)

	// Create test users
	user1 := createTestUser(t, db, "User 1", "user1@test.com")
	user2 := createTestUser(t, db, "User 2", "user2@test.com")
	user3 := createTestUser(t, db, "User 3", "user3@test.com")

	t.Run("Full CRUD workflow", func(t *testing.T) {
		// Create
		group := entity.NewGroup("Integration Test Group", []uuid.UUID{user1.ID, user2.ID})
		err := repo.Create(group)
		if err != nil {
			t.Fatalf("Failed to create group: %v", err)
		}

		// Read
		retrieved, err := repo.FindByID(group.ID.String())
		if err != nil {
			t.Fatalf("Failed to find group: %v", err)
		}
		if retrieved.Name != "Integration Test Group" {
			t.Errorf("Group name mismatch")
		}

		// Update
		retrieved.Name = "Updated Integration Test Group"
		retrieved.UserIDs = append(retrieved.UserIDs, user3.ID)
		retrieved.UpdateTimestamp()
		err = repo.Update(retrieved)
		if err != nil {
			t.Fatalf("Failed to update group: %v", err)
		}

		// Verify update
		updated, err := repo.FindByID(group.ID.String())
		if err != nil {
			t.Fatalf("Failed to find updated group: %v", err)
		}
		if updated.Name != "Updated Integration Test Group" {
			t.Errorf("Group name not updated correctly")
		}
		if len(updated.UserIDs) != 3 {
			t.Errorf("Expected 3 users after update, got %d", len(updated.UserIDs))
		}

		// Find by user
		groups, err := repo.FindByUserID(user1.ID.String())
		if err != nil {
			t.Fatalf("Failed to find groups by user: %v", err)
		}
		found := false
		for _, g := range groups {
			if g.ID == group.ID {
				found = true
				break
			}
		}
		if !found {
			t.Error("Group not found in user's groups")
		}
	})
}

