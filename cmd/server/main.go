package main

import (
	"fmt"
	"log"

	"github.com/Shobayosamuel/tap-me/config"
	"github.com/Shobayosamuel/tap-me/internal/auth"
	"github.com/Shobayosamuel/tap-me/internal/middleware"
	"github.com/Shobayosamuel/tap-me/internal/models"
	"github.com/Shobayosamuel/tap-me/internal/repository"
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
	db.AutoMigrate(&models.User{})

	// Setup repositories
	userRepo := repository.NewUserRepository(db)

	// Setup services
	authService := auth.NewService(userRepo)

	// Setup handlers
	authHandler := auth.NewHandler(authService)

	// Setup router
	r := gin.Default()

	// Public routes
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected routes
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.AuthMiddleware(authService))
	{
		apiGroup.GET("/profile", authHandler.GetProfile)
		// Add other protected routes here
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
