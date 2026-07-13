package uploads

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup) {
	h := NewHandler()
	r.POST("/uploads", h.Upload)
}
