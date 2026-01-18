package repository

import (
	"database/sql"
	"fmt"

	"github.com/DoDuy2004/slack-clone/backend/internal/database"
	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByEmail(email string) (*models.User, error)
	FindByID(id uuid.UUID) (*models.User, error)
	Update(user *models.User) error
	UpdateStatus(userID uuid.UUID, status string) error
}

type postgresUserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, email, username, password_hash, full_name, avatar_url, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at
	`
	err := r.db.QueryRow(
		query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.Status,
	).Scan(&user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *postgresUserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, username, password_hash, full_name, avatar_url, status, status_message, created_at, updated_at, last_seen_at
		FROM users
		WHERE email = $1
	`
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FullName,
		&user.AvatarURL,
		&user.Status,
		&user.StatusMessage,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastSeenAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	return user, nil
}

func (r *postgresUserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, username, password_hash, full_name, avatar_url, status, status_message, created_at, updated_at, last_seen_at
		FROM users
		WHERE id = $1
	`
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.PasswordHash,
		&user.FullName,
		&user.AvatarURL,
		&user.Status,
		&user.StatusMessage,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastSeenAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return user, nil
}

func (r *postgresUserRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET email = $1, username = $2, full_name = $3, avatar_url = $4, status = $5, status_message = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
	`
	_, err := r.db.Exec(
		query,
		user.Email,
		user.Username,
		user.FullName,
		user.AvatarURL,
		user.Status,
		user.StatusMessage,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (r *postgresUserRepository) UpdateStatus(userID uuid.UUID, status string) error {
	query := `
		UPDATE users
		SET status = $1, last_seen_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err := r.db.Exec(query, status, userID)
	return err
}
