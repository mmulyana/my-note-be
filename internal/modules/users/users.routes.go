package users

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	h := NewHandler(db)

	g := r.Group("/auth")
	{
		g.POST("/register", h.Register)
		g.POST("/login", h.Login)
	}
}
