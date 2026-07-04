package app

import (
	"my-note-be/internal/config"
	"my-note-be/internal/database"
	"my-note-be/internal/middleware"
	"my-note-be/internal/modules/folders"
	"my-note-be/internal/modules/labels"
	"my-note-be/internal/modules/notes"
	"my-note-be/internal/modules/todos"
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

	api := r.Group("/api")
	users.RegisterRoutes(api, db)

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	users.RegisterProtectedRoutes(protected, db)
	notes.RegisterRoutes(protected, db)
	folders.RegisterRoutes(protected, db)
	labels.RegisterRoutes(protected, db)
	todos.RegisterRoutes(protected, db)

	r.Run(":" + cfg.Port)
}
