package handler

import (
	"encoding/json"
	"net/http"

	"github.com/DoDuy2004/slack-clone/backend/internal/models"
	"github.com/DoDuy2004/slack-clone/backend/internal/models/dto"
	"github.com/DoDuy2004/slack-clone/backend/internal/service"
	"github.com/DoDuy2004/slack-clone/backend/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReactionHandler struct {
	reactionService service.ReactionService
	messageService  service.MessageService
	hub             *websocket.Hub
}

func NewReactionHandler(reactionService service.ReactionService, messageService service.MessageService, hub *websocket.Hub) *ReactionHandler {
	return &ReactionHandler{
		reactionService: reactionService,
		messageService:  messageService,
		hub:             hub,
	}
}

func (h *ReactionHandler) Add(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID := userIDStr.(uuid.UUID)

	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var req dto.ReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reaction, message, err := h.reactionService.AddReaction(userID, messageID, req.Emoji)
	if err != nil {
		if err == service.ErrReactionAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast
	h.broadcastReaction(websocket.EventReactionAdded, reaction, message)

	c.JSON(http.StatusCreated, reaction)
}

func (h *ReactionHandler) Remove(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID := userIDStr.(uuid.UUID)

	messageIDStr := c.Param("id")
	messageID, err := uuid.Parse(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	emoji := c.Param("emoji")
	if emoji == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Emoji is required"})
		return
	}

	message, err := h.reactionService.RemoveReaction(userID, messageID, emoji)
	if err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast
	h.broadcastReaction(websocket.EventReactionRemoved, &models.Reaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}, message)

	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
}

func (h *ReactionHandler) broadcastReaction(eventType string, reaction *models.Reaction, message *models.Message) {
	payload, _ := json.Marshal(reaction)

	h.hub.Broadcast(&websocket.WSMessage{
		Type:      eventType,
		Payload:   payload,
		ChannelID: message.ChannelID,
		DMID:      message.DMID,
	})
}
