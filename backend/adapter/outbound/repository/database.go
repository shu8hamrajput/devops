package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
	_ "github.com/lib/pq"
)

// InitPostgresDB initializes and returns a PostgreSQL database connection
func InitPostgresDB() (*sql.DB, error) {
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

	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "splitwise"
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// InitRedisClient initializes and returns a Redis client
func InitRedisClient() (*redis.Client, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	password := os.Getenv("REDIS_PASSWORD")
	if password == "" {
		password = "" // No password by default
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// CreateTables creates necessary database tables
func CreateTables(db *sql.DB) error {
	// Create users table
	usersQuery := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);
	`
	_, err := db.Exec(usersQuery)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create expenses table
	expensesQuery := `
		CREATE TABLE IF NOT EXISTS expenses (
			id UUID PRIMARY KEY,
			description VARCHAR(500) NOT NULL,
			amount DECIMAL(10, 2) NOT NULL,
			paid_by UUID NOT NULL,
			group_id UUID,
			owed_by UUID,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP,
			CONSTRAINT fk_expense_paid_by FOREIGN KEY (paid_by) REFERENCES users(id) ON DELETE CASCADE,
			CONSTRAINT fk_expense_owed_by FOREIGN KEY (owed_by) REFERENCES users(id) ON DELETE SET NULL,
			CHECK (amount > 0),
			CHECK (
				(group_id IS NOT NULL AND owed_by IS NULL) OR
				(group_id IS NULL AND owed_by IS NOT NULL)
			)
		);
	`
	_, err = db.Exec(expensesQuery)
	if err != nil {
		return fmt.Errorf("failed to create expenses table: %w", err)
	}

	// Create indexes for better query performance
	indexesQuery := `
		CREATE INDEX IF NOT EXISTS idx_expenses_paid_by ON expenses(paid_by);
		CREATE INDEX IF NOT EXISTS idx_expenses_owed_by ON expenses(owed_by);
		CREATE INDEX IF NOT EXISTS idx_expenses_group_id ON expenses(group_id);
		CREATE INDEX IF NOT EXISTS idx_expenses_updated_at ON expenses(updated_at DESC);
	`
	_, err = db.Exec(indexesQuery)
	if err != nil {
		return fmt.Errorf("failed to create expense indexes: %w", err)
	}

	// Create groups table
	groupsQuery := `
		CREATE TABLE IF NOT EXISTS groups (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			deleted_at TIMESTAMP
		);
	`
	_, err = db.Exec(groupsQuery)
	if err != nil {
		return fmt.Errorf("failed to create groups table: %w", err)
	}

	// Create group_users junction table
	groupUsersQuery := `
		CREATE TABLE IF NOT EXISTS group_users (
			group_id UUID NOT NULL,
			user_id UUID NOT NULL,
			PRIMARY KEY (group_id, user_id),
			CONSTRAINT fk_group_users_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
			CONSTRAINT fk_group_users_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
	`
	_, err = db.Exec(groupUsersQuery)
	if err != nil {
		return fmt.Errorf("failed to create group_users table: %w", err)
	}

	// Create indexes for groups
	groupIndexesQuery := `
		CREATE INDEX IF NOT EXISTS idx_groups_updated_at ON groups(updated_at DESC);
		CREATE INDEX IF NOT EXISTS idx_group_users_group_id ON group_users(group_id);
		CREATE INDEX IF NOT EXISTS idx_group_users_user_id ON group_users(user_id);
	`
	_, err = db.Exec(groupIndexesQuery)
	if err != nil {
		return fmt.Errorf("failed to create group indexes: %w", err)
	}

	return nil
}
