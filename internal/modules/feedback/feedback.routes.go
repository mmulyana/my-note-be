package feedback

import (
	"my-note-be/internal/modules/users"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, cfg RelayHubConfig) {
	h := NewHandler(NewService(cfg), users.NewService(db))

	g := r.Group("/feedback")
	{
		g.POST("", h.Create)
	}
}
