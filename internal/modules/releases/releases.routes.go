package releases

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, adminEmails []string) {
	h := NewHandler(db, adminEmails)

	g := r.Group("/releases")
	{
		g.GET("", h.FindAll)
		g.GET("/unread", h.Unread)
		g.POST("/seen", h.MarkSeen)
		g.GET("/:id", h.FindOne)
		g.POST("", h.Create)
		g.PATCH("/:id", h.Update)
		g.DELETE("/:id", h.Remove)
	}
}
