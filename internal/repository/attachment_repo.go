package repository

import (
	"database/sql"

	"github.com/DoDuy2004/slack-clone/backend/internal/database"
	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/google/uuid"
)

type AttachmentRepository interface {
	Create(attachment *models.Attachment) error
	ListByMessageID(messageID uuid.UUID) ([]*models.Attachment, error)
	GetByID(id uuid.UUID) (*models.Attachment, error)
	LinkToMessage(attachmentID, messageID uuid.UUID) error
}

type postgresAttachmentRepository struct {
	db *database.DB
}

func NewAttachmentRepository(db *database.DB) AttachmentRepository {
	return &postgresAttachmentRepository{db: db}
}

func (r *postgresAttachmentRepository) Create(attachment *models.Attachment) error {
	query := `
		INSERT INTO attachments (id, message_id, file_name, file_url, file_type, file_size)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING uploaded_at
	`
	return r.db.QueryRow(
		query,
		attachment.ID,
		attachment.MessageID,
		attachment.FileName,
		attachment.FileURL,
		attachment.FileType,
		attachment.FileSize,
	).Scan(&attachment.UploadedAt)
}

func (r *postgresAttachmentRepository) ListByMessageID(messageID uuid.UUID) ([]*models.Attachment, error) {
	query := `
		SELECT id, message_id, file_name, file_url, file_type, file_size, uploaded_at
		FROM attachments
		WHERE message_id = $1
		ORDER BY uploaded_at ASC
	`
	rows, err := r.db.Query(query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*models.Attachment
	for rows.Next() {
		a := &models.Attachment{}
		if err := rows.Scan(
			&a.ID, &a.MessageID, &a.FileName, &a.FileURL, &a.FileType, &a.FileSize, &a.UploadedAt,
		); err != nil {
			return nil, err
		}
		attachments = append(attachments, a)
	}
	return attachments, nil
}

func (r *postgresAttachmentRepository) GetByID(id uuid.UUID) (*models.Attachment, error) {
	a := &models.Attachment{}
	query := `
		SELECT id, message_id, file_name, file_url, file_type, file_size, uploaded_at
		FROM attachments
		WHERE id = $1
	`
	err := r.db.QueryRow(query, id).Scan(
		&a.ID, &a.MessageID, &a.FileName, &a.FileURL, &a.FileType, &a.FileSize, &a.UploadedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *postgresAttachmentRepository) LinkToMessage(attachmentID, messageID uuid.UUID) error {
	query := `
		UPDATE attachments
		SET message_id = $1
		WHERE id = $2
	`
	_, err := r.db.Exec(query, messageID, attachmentID)
	return err
}
