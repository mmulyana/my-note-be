package notes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	h := NewHandler(db)

	g := r.Group("/notes")
	{
		g.GET("", h.FindAll)
		g.GET("/counts", h.Counts)
		g.GET("/:id", h.FindOne)
		g.POST("", h.Create)
		g.PATCH("/:id", h.Save)
		g.PATCH("/:id/position", h.Move)
		g.PATCH("/:id/flags", h.SetFlags)
		g.DELETE("/:id", h.Remove)
	}
}
