package repository

import (
	"database/sql"
	"fmt"

	"github.com/DoDuy2004/slack-clone/backend/internal/database"
	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/google/uuid"
)

type ChannelRepository interface {
	Create(channel *models.Channel) error
	FindByID(id uuid.UUID) (*models.Channel, error)
	ListByWorkspaceID(workspaceID uuid.UUID, userID uuid.UUID) ([]*models.Channel, error)
	Update(channel *models.Channel) error
	Delete(id uuid.UUID) error

	// Member operations
	AddMember(channelID, userID uuid.UUID) error
	RemoveMember(channelID, userID uuid.UUID) error
	IsMember(channelID, userID uuid.UUID) (bool, error)
	ListMembers(channelID uuid.UUID) ([]*models.ChannelMember, error)
}

type postgresChannelRepository struct {
	db *database.DB
}

func NewChannelRepository(db *database.DB) ChannelRepository {
	return &postgresChannelRepository{db: db}
}

func (r *postgresChannelRepository) Create(channel *models.Channel) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert Channel
	query := `
		INSERT INTO channels (id, workspace_id, name, description, is_private, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`
	err = tx.QueryRow(
		query,
		channel.ID,
		channel.WorkspaceID,
		channel.Name,
		channel.Description,
		channel.IsPrivate,
		channel.CreatedBy,
	).Scan(&channel.CreatedAt, &channel.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create channel: %w", err)
	}

	// 2. Add Creator as first member
	if channel.CreatedBy != nil {
		memberQuery := `
			INSERT INTO channel_members (id, channel_id, user_id)
			VALUES ($1, $2, $3)
		`
		_, err = tx.Exec(memberQuery, uuid.New(), channel.ID, *channel.CreatedBy)
		if err != nil {
			return fmt.Errorf("failed to add creator as channel member: %w", err)
		}
	}

	return tx.Commit()
}

func (r *postgresChannelRepository) FindByID(id uuid.UUID) (*models.Channel, error) {
	c := &models.Channel{}
	query := `SELECT id, workspace_id, name, description, is_private, created_by, created_at, updated_at FROM channels WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&c.ID, &c.WorkspaceID, &c.Name, &c.Description, &c.IsPrivate, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *postgresChannelRepository) ListByWorkspaceID(workspaceID uuid.UUID, userID uuid.UUID) ([]*models.Channel, error) {
	// List public channels OR private channels where user is a member
	query := `
		SELECT c.id, c.workspace_id, c.name, c.description, c.is_private, c.created_by, c.created_at, c.updated_at
		FROM channels c
		LEFT JOIN channel_members cm ON c.id = cm.channel_id AND cm.user_id = $2
		WHERE c.workspace_id = $1 AND (c.is_private = false OR cm.user_id IS NOT NULL)
		ORDER BY c.is_private ASC, c.name ASC
	`
	rows, err := r.db.Query(query, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []*models.Channel
	for rows.Next() {
		c := &models.Channel{}
		if err := rows.Scan(&c.ID, &c.WorkspaceID, &c.Name, &c.Description, &c.IsPrivate, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		channels = append(channels, c)
	}
	return channels, nil
}

func (r *postgresChannelRepository) Update(channel *models.Channel) error {
	query := `
		UPDATE channels
		SET name = $1, description = $2, is_private = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`
	_, err := r.db.Exec(query, channel.Name, channel.Description, channel.IsPrivate, channel.ID)
	return err
}

func (r *postgresChannelRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM channels WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *postgresChannelRepository) AddMember(channelID, userID uuid.UUID) error {
	query := `
		INSERT INTO channel_members (id, channel_id, user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING
	`
	_, err := r.db.Exec(query, uuid.New(), channelID, userID)
	return err
}

func (r *postgresChannelRepository) RemoveMember(channelID, userID uuid.UUID) error {
	query := `DELETE FROM channel_members WHERE channel_id = $1 AND user_id = $2`
	_, err := r.db.Exec(query, channelID, userID)
	return err
}

func (r *postgresChannelRepository) IsMember(channelID, userID uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM channel_members WHERE channel_id = $1 AND user_id = $2)`
	err := r.db.QueryRow(query, channelID, userID).Scan(&exists)
	return exists, err
}

func (r *postgresChannelRepository) ListMembers(channelID uuid.UUID) ([]*models.ChannelMember, error) {
	query := `SELECT id, channel_id, user_id, joined_at, last_read_at FROM channel_members WHERE channel_id = $1`
	rows, err := r.db.Query(query, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.ChannelMember
	for rows.Next() {
		m := &models.ChannelMember{}
		if err := rows.Scan(&m.ID, &m.ChannelID, &m.UserID, &m.JoinedAt, &m.LastReadAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}
