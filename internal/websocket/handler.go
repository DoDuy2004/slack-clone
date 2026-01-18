package websocket

import (
	"net/http"

	"github.com/DoDuy2004/slack-clone/backend/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should check against allowed origins
		return true
	},
}

type Handler struct {
	hub        *Hub
	jwtManager *jwt.JWTManager
	presence   PresenceProvider
}

func NewHandler(hub *Hub, jwtManager *jwt.JWTManager, presence PresenceProvider) *Handler {
	return &Handler{
		hub:        hub,
		jwtManager: jwtManager,
		presence:   presence,
	}
}

func (h *Handler) ServeWS(c *gin.Context) {
	// 1. Authenticate (try cookie first then Auth header)
	tokenString, err := c.Cookie("access_token")
	if err != nil {
		// Fallback to query param or header if needed for WS
		tokenString = c.Query("token")
	}

	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	claims, err := h.jwtManager.VerifyToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID := claims.UserID

	// 2. Upgrade to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// Upgrade already handles responding to client on error
		return
	}

	// 3. Create Client
	client := &Client{
		hub:      h.hub,
		conn:     conn,
		send:     make(chan *WSMessage, 256),
		userID:   userID,
		presence: h.presence,
	}

	// 4. Register client
	client.hub.register <- client

	// 5. Notify online
	if h.presence != nil {
		h.presence.SetOnline(userID)
	}

	// 6. Start pumps
	go client.writePump()
	go client.readPump()
}
