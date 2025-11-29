// repository/user_repository.go
package repository

import (
	"database/sql"
	"fmt"
	"testcontainers-demo/models"
)

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := "SELECT id, email, name, created_at FROM users WHERE id = $1"

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by their email
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := "SELECT id, email, name, created_at FROM users WHERE email = $1"

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Create inserts a new user
func (r *UserRepository) Create(email, name string) (*models.User, error) {
	query := `
		INSERT INTO users (email, name)
		VALUES ($1, $2)
		RETURNING id, email, name, created_at
	`

	var user models.User
	err := r.db.QueryRow(query, email, name).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// Update modifies an existing user
func (r *UserRepository) Update(id int, email, name string) error {
	query := "UPDATE users SET email = $1, name = $2 WHERE id = $3"

	result, err := r.db.Exec(query, email, name, id)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete removes a user
func (r *UserRepository) Delete(id int) error {
	query := "DELETE FROM users WHERE id = $1"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// List retrieves all users
func (r *UserRepository) List() ([]models.User, error) {
	query := "SELECT id, email, name, created_at FROM users ORDER BY id"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// FindByNamePattern finds users whose name matches a pattern
func (r *UserRepository) FindByNamePattern(pattern string) ([]models.User, error) {
	query := "SELECT id, email, name, created_at FROM users WHERE name ILIKE $1 ORDER BY id"

	rows, err := r.db.Query(query, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find users by pattern: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// CountUsers returns total number of users
func (r *UserRepository) CountUsers() (int, error) {
	query := "SELECT COUNT(*) FROM users"

	var count int
	err := r.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// GetRecentUsers returns users created in the last N days
func (r *UserRepository) GetRecentUsers(days int) ([]models.User, error) {
	query := `
		SELECT id, email, name, created_at 
		FROM users
		WHERE created_at >= NOW() - INTERVAL '1 day' * $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// TransferUserEmail transfers email from one user to another (transaction example)
func (r *UserRepository) TransferUserEmail(fromID, toID int, newEmail string) error {
	// Start transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is rolled back if something goes wrong
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Get the first user to verify it exists
	var fromUser models.User
	err = tx.QueryRow("SELECT id, email, name, created_at FROM users WHERE id = $1", fromID).Scan(
		&fromUser.ID, &fromUser.Email, &fromUser.Name, &fromUser.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to get source user: %w", err)
	}

	// Update the destination user with new email
	result, err := tx.Exec("UPDATE users SET email = $1 WHERE id = $2", newEmail, toID)
	if err != nil {
		return fmt.Errorf("failed to update destination user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("destination user not found")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
