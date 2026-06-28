package todos

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	h := NewHandler(db)

	g := r.Group("/todos")
	{
		g.GET("", h.FindAll)
		g.GET("/group/notes", h.FindGroupByNotes)
		g.GET("/group/deadline", h.FindGroupByDeadline)
		g.GET("/:id", h.FindOne)
		g.POST("", h.Create)
		g.PATCH("/:id", h.Update)
		g.DELETE("/:id", h.Remove)
	}
}
