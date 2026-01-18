package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateChannelRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=80"`
	Description string `json:"description,omitempty" binding:"max=255"`
	IsPrivate   bool   `json:"is_private"`
}

type UpdateChannelRequest struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=80"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=255"`
	IsPrivate   *bool   `json:"is_private,omitempty"`
}

type ChannelResponse struct {
	ID          uuid.UUID  `json:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	IsPrivate   bool       `json:"is_private"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
