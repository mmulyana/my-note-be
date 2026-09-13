package links

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	h := NewHandler(db)

	g := r.Group("/links")
	{
		g.GET("", h.FindAll)
		g.GET("/preview", h.Preview)
	}
}
