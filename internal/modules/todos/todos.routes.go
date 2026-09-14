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
		g.GET("/group/created", h.FindGroupByCreatedAt)
		g.GET("/group/deadline", h.FindGroupByDeadline)
		g.GET("/group/today", h.FindGroupByToday)
		g.GET("/:id", h.FindOne)
		g.POST("", h.Create)
		g.PATCH("/:id", h.Update)
		g.DELETE("/:id", h.Remove)
	}
}
