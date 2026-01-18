package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateDMRequest struct {
	RecipientID uuid.UUID `json:"recipient_id" binding:"required"`
}

type DMResponse struct {
	ID           uuid.UUID        `json:"id"`
	WorkspaceID  uuid.UUID        `json:"workspace_id"`
	Participants []UserSummary    `json:"participants"`
	LastMessage  *MessageResponse `json:"last_message,omitempty"`
	CreatedAt    time.Time        `json:"created_at"`
}
