package ai

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, cfg Config) {
	h := NewHandler(NewService(cfg), db, cfg.DailyTokenLimit)

	g := r.Group("/ai")
	{
		g.GET("/usage", h.Usage)
		g.POST("/stream", h.Stream)
	}
}
