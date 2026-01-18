package repository

import (
	"database/sql"

	"github.com/DoDuy2004/slack-clone/backend/internal/database"
	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/google/uuid"
)

type ReactionRepository interface {
	Add(reaction *models.Reaction) error
	Remove(messageID, userID uuid.UUID, emoji string) error
	ListByMessageID(messageID uuid.UUID) ([]*models.Reaction, error)
	GetByMessageUserEmoji(messageID, userID uuid.UUID, emoji string) (*models.Reaction, error)
}

type postgresReactionRepository struct {
	db *database.DB
}

func NewReactionRepository(db *database.DB) ReactionRepository {
	return &postgresReactionRepository{db: db}
}

func (r *postgresReactionRepository) Add(reaction *models.Reaction) error {
	query := `
		INSERT INTO reactions (id, message_id, user_id, emoji)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`
	return r.db.QueryRow(
		query,
		reaction.ID,
		reaction.MessageID,
		reaction.UserID,
		reaction.Emoji,
	).Scan(&reaction.CreatedAt)
}

func (r *postgresReactionRepository) Remove(messageID, userID uuid.UUID, emoji string) error {
	query := `
		DELETE FROM reactions
		WHERE message_id = $1 AND user_id = $2 AND emoji = $3
	`
	_, err := r.db.Exec(query, messageID, userID, emoji)
	return err
}

func (r *postgresReactionRepository) ListByMessageID(messageID uuid.UUID) ([]*models.Reaction, error) {
	query := `
		SELECT r.id, r.message_id, r.user_id, r.emoji, r.created_at,
		       u.username, u.avatar_url, u.full_name
		FROM reactions r
		JOIN users u ON r.user_id = u.id
		WHERE r.message_id = $1
		ORDER BY r.created_at ASC
	`
	rows, err := r.db.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reactions []*models.Reaction
	for rows.Next() {
		re := &models.Reaction{}
		re.User = &models.User{}
		if err := rows.Scan(
			&re.ID, &re.MessageID, &re.UserID, &re.Emoji, &re.CreatedAt,
			&re.User.Username, &re.User.AvatarURL, &re.User.FullName,
		); err != nil {
			return nil, err
		}
		re.User.ID = re.UserID
		reactions = append(reactions, re)
	}
	return reactions, nil
}

func (r *postgresReactionRepository) GetByMessageUserEmoji(messageID, userID uuid.UUID, emoji string) (*models.Reaction, error) {
	re := &models.Reaction{}
	query := `
		SELECT id, message_id, user_id, emoji, created_at
		FROM reactions
		WHERE message_id = $1 AND user_id = $2 AND emoji = $3
	`
	err := r.db.QueryRow(query, messageID, userID, emoji).Scan(
		&re.ID, &re.MessageID, &re.UserID, &re.Emoji, &re.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return re, nil
}
