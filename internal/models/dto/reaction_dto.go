package dto

import (
	"time"

	"github.com/google/uuid"
)

type ReactionRequest struct {
	Emoji string `json:"emoji" binding:"required"`
}

type ReactionResponse struct {
	ID        uuid.UUID `json:"id"`
	MessageID uuid.UUID `json:"message_id"`
	UserID    uuid.UUID `json:"user_id"`
	Emoji     string    `json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}
