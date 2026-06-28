package response

import "github.com/gin-gonic/gin"

type envelope struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func OK(c *gin.Context, message string, data any) {
	c.JSON(200, envelope{Message: message, Data: data})
}

func Created(c *gin.Context, message string, data any) {
	c.JSON(201, envelope{Message: message, Data: data})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, envelope{Message: message, Data: nil})
}
