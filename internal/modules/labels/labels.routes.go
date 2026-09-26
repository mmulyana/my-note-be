package labels

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	h := NewHandler(db)

	g := r.Group("/labels")
	{
		g.GET("", h.FindAll)
		g.GET("/:id", h.FindOne)
	}
}
