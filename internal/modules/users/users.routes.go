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
		g.POST("/guest", h.Guest)
		g.POST("/refresh", h.RefreshToken)
		g.POST("/logout", h.Logout)
	}
}

func RegisterProtectedRoutes(r *gin.RouterGroup, db *gorm.DB) {
	h := NewHandler(db)
	r.GET("/me", h.Me)
	r.PATCH("/me", h.UpdateProfile)
	r.PATCH("/me/password", h.ChangePassword)
	r.POST("/auth/guest/upgrade", h.UpgradeGuest)
}
