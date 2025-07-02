package main

import (
	"fmt"
	"log"

	"github.com/Shobayosamuel/tap-me/config"
	"github.com/Shobayosamuel/tap-me/internal/auth"
	"github.com/Shobayosamuel/tap-me/internal/chat"
	"github.com/Shobayosamuel/tap-me/internal/middleware"
	"github.com/Shobayosamuel/tap-me/internal/models"
	"github.com/Shobayosamuel/tap-me/internal/repository"
	"github.com/Shobayosamuel/tap-me/internal/ws"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load config
	cfg := config.Load()

	// Setup database
	db := setupDatabase(cfg)

	// Auto migrate
	db.AutoMigrate(&models.User{}, &models.Room{}, &models.Message{}, &models.RoomMember{})

	// Setup repositories
	userRepo := repository.NewUserRepository(db)
	roomRepo := repository.NewRoomRepository(db)
	messageRepo := repository.NewMessageRepository(db)

	// Setup services
	authService := auth.NewService(userRepo)
	chatService := chat.NewService(roomRepo, messageRepo, userRepo)

	// Setup WebSocket hub
	hub := ws.NewHub(chatService)
	go hub.Run()

	// Setup handlers
	authHandler := auth.NewHandler(authService)
	chatHandler := chat.NewHandler(chatService, authService, hub)

	// Setup router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
	}

	// WebSocket endpoint (token-based auth)
	r.GET("/ws", chatHandler.HandleWebSocket)

	// Protected routes
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.AuthMiddleware(authService))
	{
		// Auth routes
		apiGroup.GET("/profile", authHandler.GetProfile)

		// Chat routes
		chatGroup := apiGroup.Group("/chat")
		{
			chatGroup.POST("/rooms", chatHandler.CreateRoom)
			chatGroup.GET("/rooms", chatHandler.GetUserRooms)
			chatGroup.GET("/rooms/:roomId/messages", chatHandler.GetRoomMessages)
			chatGroup.POST("/rooms/:roomId/join", chatHandler.JoinRoom)
			chatGroup.GET("/rooms/:roomId/online", chatHandler.GetOnlineUsers)
		}
	}

	// Start server
	log.Printf("Server starting on :%s", cfg.Server.Port)
	r.Run(":" + cfg.Server.Port)
}

func setupDatabase(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	return db
}
