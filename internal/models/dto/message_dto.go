package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateMessageRequest struct {
	Content         string      `json:"content" binding:"required,min=1"`
	ParentMessageID *uuid.UUID  `json:"parent_message_id,omitempty"`
	AttachmentIDs   []uuid.UUID `json:"attachment_ids,omitempty"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" binding:"required,min=1"`
}

type MessageResponse struct {
	ID              uuid.UUID    `json:"id"`
	Content         string       `json:"content"`
	SenderID        *uuid.UUID   `json:"sender_id,omitempty"`
	ChannelID       *uuid.UUID   `json:"channel_id,omitempty"`
	DMID            *uuid.UUID   `json:"dm_id,omitempty"`
	ParentMessageID *uuid.UUID   `json:"parent_message_id,omitempty"`
	EditedAt        *time.Time   `json:"edited_at,omitempty"`
	DeletedAt       *time.Time   `json:"deleted_at,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	ReplyCount      int          `json:"reply_count"`
	Sender          *UserSummary `json:"sender,omitempty"`
}

type UserSummary struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	FullName  *string   `json:"full_name,omitempty"`
}
