package repository

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"splitwise/domain/entity"
	"splitwise/domain/port/outbound"
)

// PostgresAuthRepository implements AuthRepository using PostgreSQL
type PostgresAuthRepository struct {
	db *sql.DB
}

// NewPostgresAuthRepository creates a new PostgreSQL auth repository
func NewPostgresAuthRepository(db *sql.DB) outbound.AuthRepository {
	return &PostgresAuthRepository{db: db}
}

// FindUserByEmail retrieves a user by email
func (r *PostgresAuthRepository) FindUserByEmail(email string) (*entity.User, error) {
	user := &entity.User{}
	query := `
		SELECT id, name, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// CreateUser inserts a new user into the database
func (r *PostgresAuthRepository) CreateUser(user *entity.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Exec(query, user.ID, user.Name, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" { // Unique violation
			return fmt.Errorf("user with email %s already exists", user.Email)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}
