package repository

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"splitwise/domain/entity"
	"splitwise/domain/port/outbound"
)

// PostgresGroupRepository implements GroupRepository using PostgreSQL
type PostgresGroupRepository struct {
	db *sql.DB
}

// NewPostgresGroupRepository creates a new PostgreSQL group repository
func NewPostgresGroupRepository(db *sql.DB) outbound.GroupRepository {
	return &PostgresGroupRepository{db: db}
}

// Create inserts a new group into the database along with its user relationships
func (r *PostgresGroupRepository) Create(group *entity.Group) error {
	// Start a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert the group
	groupQuery := `
		INSERT INTO groups (id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.Exec(groupQuery, group.ID, group.Name, group.CreatedAt, group.UpdatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // Unique violation
				return fmt.Errorf("group with name %s already exists", group.Name)
			}
		}
		return fmt.Errorf("failed to create group: %w", err)
	}

	// Insert user relationships into group_users junction table
	if len(group.UserIDs) > 0 {
		userQuery := `
			INSERT INTO group_users (group_id, user_id)
			VALUES ($1, $2)
		`
		stmt, err := tx.Prepare(userQuery)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, userID := range group.UserIDs {
			_, err = stmt.Exec(group.ID, userID)
			if err != nil {
				if pqErr, ok := err.(*pq.Error); ok {
					if pqErr.Code == "23503" { // Foreign key violation
						return fmt.Errorf("invalid user ID: %s", userID.String())
					}
					if pqErr.Code == "23505" { // Unique violation
						// User already in group, skip
						continue
					}
				}
				return fmt.Errorf("failed to add user to group: %w", err)
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByID retrieves a group by ID along with its user relationships
func (r *PostgresGroupRepository) FindByID(id string) (*entity.Group, error) {
	groupID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid group ID: %w", err)
	}

	// Get the group
	group := &entity.Group{}
	groupQuery := `
		SELECT id, name, created_at, updated_at
		FROM groups
		WHERE id = $1 AND deleted_at IS NULL
	`
	err = r.db.QueryRow(groupQuery, groupID).Scan(
		&group.ID,
		&group.Name,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("group not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Get all users in the group
	userQuery := `
		SELECT user_id
		FROM group_users
		WHERE group_id = $1
		ORDER BY user_id
	`
	rows, err := r.db.Query(userQuery, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group users: %w", err)
	}
	defer rows.Close()

	userIDs := []uuid.UUID{}
	for rows.Next() {
		var userID uuid.UUID
		err := rows.Scan(&userID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group users: %w", err)
	}

	group.UserIDs = userIDs
	return group, nil
}

// FindByUserID retrieves all groups where the user is a member, sorted by UpdatedAt descending
func (r *PostgresGroupRepository) FindByUserID(userID string) ([]*entity.Group, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	query := `
		SELECT g.id, g.name, g.created_at, g.updated_at
		FROM groups g
		INNER JOIN group_users gu ON g.id = gu.group_id
		WHERE gu.user_id = $1 AND g.deleted_at IS NULL
		ORDER BY g.updated_at DESC
	`
	rows, err := r.db.Query(query, userUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}
	defer rows.Close()

	groups := []*entity.Group{}
	groupMap := make(map[uuid.UUID]*entity.Group)

	for rows.Next() {
		group := &entity.Group{}
		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		group.UserIDs = []uuid.UUID{} // Initialize empty slice
		groupMap[group.ID] = group
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating groups: %w", err)
	}

	// Get all user IDs for each group
	if len(groups) > 0 {
		groupIDs := make([]uuid.UUID, 0, len(groups))
		for _, g := range groups {
			groupIDs = append(groupIDs, g.ID)
		}

		userQuery := `
			SELECT group_id, user_id
			FROM group_users
			WHERE group_id = ANY($1)
			ORDER BY group_id, user_id
		`
		userRows, err := r.db.Query(userQuery, pq.Array(groupIDs))
		if err != nil {
			return nil, fmt.Errorf("failed to get group users: %w", err)
		}
		defer userRows.Close()

		for userRows.Next() {
			var gID, uID uuid.UUID
			err := userRows.Scan(&gID, &uID)
			if err != nil {
				return nil, fmt.Errorf("failed to scan group user: %w", err)
			}
			if group, exists := groupMap[gID]; exists {
				group.UserIDs = append(group.UserIDs, uID)
			}
		}

		if err = userRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating group users: %w", err)
		}
	}

	return groups, nil
}

// FindAll retrieves all groups, sorted by updated_at descending
func (r *PostgresGroupRepository) FindAll() ([]*entity.Group, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM groups
		WHERE deleted_at IS NULL
		ORDER BY updated_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}
	defer rows.Close()

	groups := []*entity.Group{}
	groupMap := make(map[uuid.UUID]*entity.Group)

	for rows.Next() {
		group := &entity.Group{}
		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		group.UserIDs = []uuid.UUID{} // Initialize empty slice
		groupMap[group.ID] = group
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating groups: %w", err)
	}

	// Get all user IDs for each group
	if len(groups) > 0 {
		groupIDs := make([]uuid.UUID, 0, len(groups))
		for _, g := range groups {
			groupIDs = append(groupIDs, g.ID)
		}

		userQuery := `
			SELECT group_id, user_id
			FROM group_users
			WHERE group_id = ANY($1)
			ORDER BY group_id, user_id
		`
		userRows, err := r.db.Query(userQuery, pq.Array(groupIDs))
		if err != nil {
			return nil, fmt.Errorf("failed to get group users: %w", err)
		}
		defer userRows.Close()

		for userRows.Next() {
			var gID, uID uuid.UUID
			err := userRows.Scan(&gID, &uID)
			if err != nil {
				return nil, fmt.Errorf("failed to scan group user: %w", err)
			}
			if group, exists := groupMap[gID]; exists {
				group.UserIDs = append(group.UserIDs, uID)
			}
		}

		if err = userRows.Err(); err != nil {
			return nil, fmt.Errorf("error iterating group users: %w", err)
		}
	}

	return groups, nil
}

// Update updates an existing group and its user relationships
func (r *PostgresGroupRepository) Update(group *entity.Group) error {
	// Start a transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update the group
	groupQuery := `
		UPDATE groups
		SET name = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`
	result, err := tx.Exec(groupQuery, group.Name, group.UpdatedAt, group.ID)
	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("group not found")
	}

	// Delete existing user relationships
	deleteQuery := `DELETE FROM group_users WHERE group_id = $1`
	_, err = tx.Exec(deleteQuery, group.ID)
	if err != nil {
		return fmt.Errorf("failed to delete group users: %w", err)
	}

	// Insert new user relationships
	if len(group.UserIDs) > 0 {
		insertQuery := `
			INSERT INTO group_users (group_id, user_id)
			VALUES ($1, $2)
		`
		stmt, err := tx.Prepare(insertQuery)
		if err != nil {
			return fmt.Errorf("failed to prepare statement: %w", err)
		}
		defer stmt.Close()

		for _, userID := range group.UserIDs {
			_, err = stmt.Exec(group.ID, userID)
			if err != nil {
				if pqErr, ok := err.(*pq.Error); ok {
					if pqErr.Code == "23503" { // Foreign key violation
						return fmt.Errorf("invalid user ID: %s", userID.String())
					}
				}
				return fmt.Errorf("failed to add user to group: %w", err)
			}
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Delete performs a soft delete on a group by setting deleted_at timestamp
func (r *PostgresGroupRepository) Delete(id string) error {
	groupID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid group ID: %w", err)
	}

	query := `
		UPDATE groups
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	result, err := r.db.Exec(query, groupID)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}
