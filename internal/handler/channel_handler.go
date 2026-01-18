package handler

import (
	"net/http"

	"github.com/DoDuy2004/slack-clone/backend/internal/models/dto"
	"github.com/DoDuy2004/slack-clone/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChannelHandler struct {
	channelService service.ChannelService
}

func NewChannelHandler(channelService service.ChannelService) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
	}
}

func (h *ChannelHandler) Create(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID := userIDStr.(uuid.UUID)

	workspaceIDStr := c.Param("workspace_id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	var req dto.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.channelService.CreateChannel(userID, workspaceID, &req)
	if err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, channel)
}

func (h *ChannelHandler) Get(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID := userIDStr.(uuid.UUID)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	channel, err := h.channelService.GetChannel(id, userID)
	if err != nil {
		if err == service.ErrChannelNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

func (h *ChannelHandler) ListByWorkspace(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID := userIDStr.(uuid.UUID)

	workspaceIDStr := c.Param("workspace_id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid workspace ID"})
		return
	}

	channels, err := h.channelService.ListWorkspaceChannels(workspaceID, userID)
	if err != nil {
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, channels)
}

func (h *ChannelHandler) Update(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID := userIDStr.(uuid.UUID)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	var req dto.UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.channelService.UpdateChannel(userID, id, &req)
	if err != nil {
		if err == service.ErrChannelNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, channel)
}

func (h *ChannelHandler) Delete(c *gin.Context) {
	userIDStr, _ := c.Get("user_id")
	userID := userIDStr.(uuid.UUID)

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel ID"})
		return
	}

	err = h.channelService.DeleteChannel(userID, id)
	if err != nil {
		if err == service.ErrChannelNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err == service.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel deleted successfully"})
}
