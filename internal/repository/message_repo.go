package repository

import (
	"database/sql"

	"github.com/DoDuy2004/slack-clone/backend/internal/database"
	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/google/uuid"
)

type MessageRepository interface {
	Create(message *models.Message) error
	FindByID(id uuid.UUID) (*models.Message, error)
	ListByChannelID(channelID uuid.UUID, limit, offset int) ([]*models.Message, error)
	ListByDMID(dmID uuid.UUID, limit, offset int) ([]*models.Message, error)
	ListReplies(parentID uuid.UUID) ([]*models.Message, error)
	Update(message *models.Message) error
	SoftDelete(id uuid.UUID) error
	Search(workspaceID uuid.UUID, query string, limit, offset int) ([]*models.Message, error)
}

type postgresMessageRepository struct {
	db *database.DB
}

func NewMessageRepository(db *database.DB) MessageRepository {
	return &postgresMessageRepository{db: db}
}

func (r *postgresMessageRepository) Create(message *models.Message) error {
	query := `
		INSERT INTO messages (id, content, sender_id, channel_id, dm_id, parent_message_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		message.ID,
		message.Content,
		message.SenderID,
		message.ChannelID,
		message.DMID,
		message.ParentMessageID,
	).Scan(&message.CreatedAt, &message.UpdatedAt)
}

func (r *postgresMessageRepository) FindByID(id uuid.UUID) (*models.Message, error) {
	m := &models.Message{}
	query := `
		SELECT id, content, sender_id, channel_id, dm_id, parent_message_id, edited_at, deleted_at, created_at, updated_at
		FROM messages
		WHERE id = $1
	`
	err := r.db.QueryRow(query, id).Scan(
		&m.ID, &m.Content, &m.SenderID, &m.ChannelID, &m.DMID, &m.ParentMessageID, &m.EditedAt, &m.DeletedAt, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Attach reactions
	messages := []*models.Message{m}
	if err := r.attachReactions(messages); err != nil {
		return nil, err
	}

	return m, nil
}

func (r *postgresMessageRepository) ListByChannelID(channelID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT m.id, m.content, m.sender_id, m.channel_id, m.dm_id, m.parent_message_id, m.edited_at, m.deleted_at, m.created_at, m.updated_at,
		       u.username, u.avatar_url, u.full_name,
		       (SELECT COUNT(*) FROM messages WHERE parent_message_id = m.id) as reply_count
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.channel_id = $1 AND m.parent_message_id IS NULL
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, channelID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		m := &models.Message{}
		var username, fullName, avatarURL sql.NullString
		if err := rows.Scan(
			&m.ID, &m.Content, &m.SenderID, &m.ChannelID, &m.DMID, &m.ParentMessageID, &m.EditedAt, &m.DeletedAt, &m.CreatedAt, &m.UpdatedAt,
			&username, &avatarURL, &fullName, &m.ReplyCount,
		); err != nil {
			return nil, err
		}

		if username.Valid {
			m.Sender = &models.User{
				ID:       *m.SenderID,
				Username: username.String,
			}
			if avatarURL.Valid {
				m.Sender.AvatarURL = &avatarURL.String
			}
			if fullName.Valid {
				m.Sender.FullName = &fullName.String
			}
		}
		messages = append(messages, m)
	}

	// Attach reactions to messages
	if len(messages) > 0 {
		if err := r.attachReactions(messages); err != nil {
			return nil, err
		}
	}

	return messages, nil
}

func (r *postgresMessageRepository) ListByDMID(dmID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	query := `
		SELECT m.id, m.content, m.sender_id, m.channel_id, m.dm_id, m.parent_message_id, m.edited_at, m.deleted_at, m.created_at, m.updated_at,
		       u.username, u.avatar_url, u.full_name,
		       (SELECT COUNT(*) FROM messages WHERE parent_message_id = m.id) as reply_count
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.dm_id = $1 AND m.parent_message_id IS NULL
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, dmID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		m := &models.Message{}
		var username, fullName, avatarURL sql.NullString
		if err := rows.Scan(
			&m.ID, &m.Content, &m.SenderID, &m.ChannelID, &m.DMID, &m.ParentMessageID, &m.EditedAt, &m.DeletedAt, &m.CreatedAt, &m.UpdatedAt,
			&username, &avatarURL, &fullName, &m.ReplyCount,
		); err != nil {
			return nil, err
		}

		if username.Valid {
			m.Sender = &models.User{
				ID:       *m.SenderID,
				Username: username.String,
			}
			if avatarURL.Valid {
				m.Sender.AvatarURL = &avatarURL.String
			}
			if fullName.Valid {
				m.Sender.FullName = &fullName.String
			}
		}
		messages = append(messages, m)
	}

	// Attach reactions
	if len(messages) > 0 {
		if err := r.attachReactions(messages); err != nil {
			return nil, err
		}
		if err := r.attachAttachments(messages); err != nil {
			return nil, err
		}
	}

	return messages, nil
}

func (r *postgresMessageRepository) attachReactions(messages []*models.Message) error {
	messageIDs := make([]uuid.UUID, len(messages))
	msgMap := make(map[uuid.UUID]*models.Message)
	for i, m := range messages {
		messageIDs[i] = m.ID
		msgMap[m.ID] = m
		m.Reactions = []models.Reaction{}     // Initialize
		m.Attachments = []models.Attachment{} // Initialize
	}

	query := `
		SELECT r.id, r.message_id, r.user_id, r.emoji, r.created_at,
		       u.username, u.avatar_url, u.full_name
		FROM reactions r
		JOIN users u ON r.user_id = u.id
		WHERE r.message_id = ANY($1)
		ORDER BY r.created_at ASC
	`
	rows, err := r.db.Query(query, messageIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		re := models.Reaction{}
		re.User = &models.User{}
		if err := rows.Scan(
			&re.ID, &re.MessageID, &re.UserID, &re.Emoji, &re.CreatedAt,
			&re.User.Username, &re.User.AvatarURL, &re.User.FullName,
		); err != nil {
			return err
		}
		re.User.ID = re.UserID
		if m, ok := msgMap[re.MessageID]; ok {
			m.Reactions = append(m.Reactions, re)
		}
	}
	return nil
}

func (r *postgresMessageRepository) attachAttachments(messages []*models.Message) error {
	messageIDs := make([]uuid.UUID, len(messages))
	msgMap := make(map[uuid.UUID]*models.Message)
	for i, m := range messages {
		messageIDs[i] = m.ID
		msgMap[m.ID] = m
	}

	query := `
		SELECT id, message_id, file_name, file_url, file_type, file_size, uploaded_at
		FROM attachments
		WHERE message_id = ANY($1)
		ORDER BY uploaded_at ASC
	`
	rows, err := r.db.Query(query, messageIDs)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		a := models.Attachment{}
		if err := rows.Scan(
			&a.ID, &a.MessageID, &a.FileName, &a.FileURL, &a.FileType, &a.FileSize, &a.UploadedAt,
		); err != nil {
			return err
		}
		if m, ok := msgMap[a.MessageID]; ok {
			m.Attachments = append(m.Attachments, a)
		}
	}
	return nil
}

func (r *postgresMessageRepository) ListReplies(parentID uuid.UUID) ([]*models.Message, error) {
	query := `
		SELECT m.id, m.content, m.sender_id, m.channel_id, m.dm_id, m.parent_message_id, m.edited_at, m.deleted_at, m.created_at, m.updated_at,
		       u.username, u.avatar_url, u.full_name
		FROM messages m
		LEFT JOIN users u ON m.sender_id = u.id
		WHERE m.parent_message_id = $1
		ORDER BY m.created_at ASC
	`
	rows, err := r.db.Query(query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		m := &models.Message{}
		var username, fullName, avatarURL sql.NullString
		if err := rows.Scan(
			&m.ID, &m.Content, &m.SenderID, &m.ChannelID, &m.DMID, &m.ParentMessageID, &m.EditedAt, &m.DeletedAt, &m.CreatedAt, &m.UpdatedAt,
			&username, &avatarURL, &fullName,
		); err != nil {
			return nil, err
		}

		if username.Valid {
			m.Sender = &models.User{
				ID:       *m.SenderID,
				Username: username.String,
			}
			if avatarURL.Valid {
				m.Sender.AvatarURL = &avatarURL.String
			}
			if fullName.Valid {
				m.Sender.FullName = &fullName.String
			}
		}
		messages = append(messages, m)
	}

	// Attach reactions
	if len(messages) > 0 {
		if err := r.attachReactions(messages); err != nil {
			return nil, err
		}
	}

	return messages, nil
}

func (r *postgresMessageRepository) Update(message *models.Message) error {
	query := `
		UPDATE messages
		SET content = $1, edited_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err := r.db.Exec(query, message.Content, message.ID)
	return err
}

func (r *postgresMessageRepository) SoftDelete(id uuid.UUID) error {
	query := `
		UPDATE messages
		SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *postgresMessageRepository) Search(workspaceID uuid.UUID, query string, limit, offset int) ([]*models.Message, error) {
	sqlQuery := `
		SELECT m.id, m.content, m.sender_id, m.channel_id, m.dm_id, m.parent_message_id, m.edited_at, m.deleted_at, m.created_at, m.updated_at,
		       u.username, u.avatar_url, u.full_name
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE (
			m.channel_id IN (SELECT id FROM channels WHERE workspace_id = $1)
			OR 
			m.dm_id IN (SELECT id FROM direct_messages WHERE workspace_id = $1)
		)
		AND to_tsvector('english', m.content) @@ plainto_tsquery('english', $2)
		AND m.deleted_at IS NULL
		ORDER BY ts_rank(to_tsvector('english', m.content), plainto_tsquery('english', $2)) DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.Query(sqlQuery, workspaceID, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		m := &models.Message{}
		var username, fullName, avatarURL sql.NullString
		if err := rows.Scan(
			&m.ID, &m.Content, &m.SenderID, &m.ChannelID, &m.DMID, &m.ParentMessageID, &m.EditedAt, &m.DeletedAt, &m.CreatedAt, &m.UpdatedAt,
			&username, &avatarURL, &fullName,
		); err != nil {
			return nil, err
		}

		if username.Valid {
			m.Sender = &models.User{
				ID:       *m.SenderID,
				Username: username.String,
			}
			if avatarURL.Valid {
				m.Sender.AvatarURL = &avatarURL.String
			}
			if fullName.Valid {
				m.Sender.FullName = &fullName.String
			}
		}
		messages = append(messages, m)
	}

	if len(messages) > 0 {
		if err := r.attachReactions(messages); err != nil {
			return nil, err
		}
		if err := r.attachAttachments(messages); err != nil {
			return nil, err
		}
	}

	return messages, nil
}
