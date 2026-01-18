package dto

import (
	"time"

	"github.com/google/uuid"
)

type FileResponse struct {
	ID         uuid.UUID  `json:"id"`
	MessageID  *uuid.UUID `json:"message_id,omitempty"`
	FileName   string     `json:"file_name"`
	FileURL    string     `json:"file_url"`
	FileType   *string    `json:"file_type,omitempty"`
	FileSize   *int64     `json:"file_size,omitempty"`
	UploadedAt time.Time  `json:"uploaded_at"`
}

type UploadResponse struct {
	AttachmentID uuid.UUID `json:"attachment_id"`
	URL          string    `json:"url"`
}
