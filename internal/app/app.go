package app

import (
	"my-note-be/internal/config"
	"my-note-be/internal/database"
	"my-note-be/internal/middleware"
	"my-note-be/internal/modules/feedback"
	"my-note-be/internal/modules/folders"
	"my-note-be/internal/modules/labels"
	"my-note-be/internal/modules/links"
	"my-note-be/internal/modules/notes"
	"my-note-be/internal/modules/todos"
	"my-note-be/internal/modules/uploads"
	"my-note-be/internal/modules/users"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func Run() {
	cfg := config.Load()

	database.RunMigrations(cfg.DatabaseURL)
	db := database.Connect(cfg.DatabaseURL)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}))

	middleware.SetSecret(cfg.JWTSecret)

	r.Static("/uploads", "./uploads")

	api := r.Group("/api")
	users.RegisterRoutes(api, db)

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	users.RegisterProtectedRoutes(protected, db)
	notes.RegisterRoutes(protected, db)
	folders.RegisterRoutes(protected, db)
	labels.RegisterRoutes(protected, db)
	todos.RegisterRoutes(protected, db)
	links.RegisterRoutes(protected, db)
	uploads.RegisterRoutes(protected)
	feedback.RegisterRoutes(protected, db, feedback.RelayHubConfig{
		BaseURL: cfg.RelayHubBaseURL,
		APIKey:  cfg.RelayHubToken,
	})

	r.Run(":" + cfg.Port)
}
