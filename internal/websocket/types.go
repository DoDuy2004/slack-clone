package websocket

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Event types
const (
	EventMessageNew      = "message.new"
	EventMessageUpdated  = "message.updated"
	EventMessageDeleted  = "message.deleted"
	EventUserTyping      = "user.typing"
	EventUserPresence    = "user.presence"
	EventChannelJoined   = "channel.joined"
	EventWorkspaceJoined = "workspace.joined"
)

// WSMessage represents the structure of messages sent over WebSocket
type WSMessage struct {
	Type        string          `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	WorkspaceID *uuid.UUID      `json:"workspace_id,omitempty"`
	ChannelID   *uuid.UUID      `json:"channel_id,omitempty"`
	DMID        *uuid.UUID      `json:"dm_id,omitempty"`
	UserID      *uuid.UUID      `json:"user_id,omitempty"`
}

// TypingPayload represents the payload for typing indicators
type TypingPayload struct {
	UserID    uuid.UUID `json:"user_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	IsTyping  bool      `json:"is_typing"`
}

// PresencePayload represents the payload for user status changes
type PresencePayload struct {
	UserID uuid.UUID `json:"user_id"`
	Status string    `json:"status"` // online, offline, away
}
