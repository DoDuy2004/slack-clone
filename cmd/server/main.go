package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"os"

	"github.com/DoDuy2004/slack-clone/backend/internal/config"
	"github.com/DoDuy2004/slack-clone/backend/internal/database"
	"github.com/DoDuy2004/slack-clone/backend/internal/handler"
	"github.com/DoDuy2004/slack-clone/backend/internal/middleware"
	"github.com/DoDuy2004/slack-clone/backend/internal/repository"
	"github.com/DoDuy2004/slack-clone/backend/internal/service"
	"github.com/DoDuy2004/slack-clone/backend/internal/websocket"
	"github.com/DoDuy2004/slack-clone/backend/pkg/jwt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Connect to PostgreSQL
	db, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Connect to Redis
	redisClient, err := database.NewRedisClient(cfg.RedisURL, cfg.RedisPassword)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Initialize JWT manager
	privateKey, err := os.ReadFile("private.pem")
	if err != nil {
		log.Fatal("Failed to read private key:", err)
	}
	publicKey, err := os.ReadFile("public.pem")
	if err != nil {
		log.Fatal("Failed to read public key:", err)
	}

	jwtManager, err := jwt.NewJWTManager(
		string(privateKey),
		string(publicKey),
		cfg.JWTAccessExpiry,
		cfg.JWTRefreshExpiry,
	)
	if err != nil {
		log.Fatal("Failed to initialize JWT manager:", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	workspaceRepo := repository.NewWorkspaceRepository(db)
	channelRepo := repository.NewChannelRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	dmRepo := repository.NewDMRepository(db)
	reactionRepo := repository.NewReactionRepository(db)

	// Initialize services
	authService := service.NewAuthService(userRepo, jwtManager)
	workspaceService := service.NewWorkspaceService(workspaceRepo)
	channelService := service.NewChannelService(channelRepo, workspaceRepo)
	messageService := service.NewMessageService(messageRepo, channelRepo, workspaceRepo, dmRepo)
	dmService := service.NewDMService(dmRepo, workspaceRepo, userRepo)
	reactionService := service.NewReactionService(reactionRepo, messageRepo, channelRepo, dmRepo, workspaceRepo)

	// Initialize WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	presenceService := service.NewPresenceService(userRepo, hub)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, cfg)
	workspaceHandler := handler.NewWorkspaceHandler(workspaceService)
	channelHandler := handler.NewChannelHandler(channelService)
	messageHandler := handler.NewMessageHandler(messageService, hub) // Inject hub
	dmHandler := handler.NewDMHandler(dmService)
	reactionHandler := handler.NewReactionHandler(reactionService, messageService, hub)
	wsHandler := websocket.NewHandler(hub, jwtManager, presenceService)

	// Create Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-CSRF-Token", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// CSRF Protection for state-changing requests
		api.Use(middleware.CSRFMiddleware(cfg.AllowedOrigins))

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.POST("/logout", authHandler.Logout)
		}

		// Protected routes (require authentication)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			// WebSocket endpoint
			protected.GET("/ws", wsHandler.ServeWS)

			// User routes
			users := protected.Group("/users")
			{
				users.GET("/me", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Get current user - TODO"})
				})
				users.PUT("/me", func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "Update user - TODO"})
				})
			}

			// Workspace routes
			workspaces := protected.Group("/workspaces")
			{
				workspaces.GET("", workspaceHandler.List)
				workspaces.POST("", workspaceHandler.Create)
				workspaces.GET("/:id", workspaceHandler.Get)
				workspaces.PUT("/:id", workspaceHandler.Update)
				workspaces.DELETE("/:id", workspaceHandler.Delete)

				// Channel routes within a workspace
				workspaces.GET("/:workspace_id/channels", channelHandler.ListByWorkspace)
				workspaces.POST("/:workspace_id/channels", channelHandler.Create)

				// DM routes within a workspace
				workspaces.GET("/:workspace_id/dms", dmHandler.List)
				workspaces.POST("/:workspace_id/dms", dmHandler.GetOrCreate)
			}

			// Individual channel routes
			channels := protected.Group("/channels")
			{
				channels.GET("/:id", channelHandler.Get)
				channels.PUT("/:id", channelHandler.Update)
				channels.DELETE("/:id", channelHandler.Delete)

				// Message routes within a channel
				channels.GET("/:id/messages", messageHandler.ListByChannel)
				channels.POST("/:id/messages", messageHandler.SendChannel)
			}

			// Individual DM routes
			dms := protected.Group("/dms")
			{
				dms.GET("/:id/messages", messageHandler.ListByDM)
				dms.POST("/:id/messages", messageHandler.SendDM)
			}

			// Individual message actions
			messages := protected.Group("/messages")
			{
				messages.GET("/:id/thread", messageHandler.GetThread)
				messages.PUT("/:id", messageHandler.Update)
				messages.DELETE("/:id", messageHandler.Delete)

				// Reaction routes
				messages.POST("/:id/reactions", reactionHandler.Add)
				messages.DELETE("/:id/reactions/:emoji", reactionHandler.Remove)
			}
		}
	}

	// WebRTC signaling endpoint
	router.GET("/webrtc/signaling", func(c *gin.Context) {
		// TODO: Implement WebRTC signaling handler
		c.JSON(http.StatusOK, gin.H{"message": "WebRTC signaling - TODO"})
	})

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Server starting on %s", addr)
	log.Printf("📝 Gin mode: %s", cfg.GinMode)
	log.Printf("🌐 Allowed origins: %v", cfg.AllowedOrigins)

	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
