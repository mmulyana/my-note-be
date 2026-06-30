package response

import "github.com/gin-gonic/gin"

type envelope struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type paginatedEnvelope struct {
	Message string     `json:"message"`
	Data    any        `json:"data"`
	Meta    pagination `json:"meta"`
}

type pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

func OK(c *gin.Context, message string, data any) {
	c.JSON(200, envelope{Message: message, Data: data})
}

func OKPaginated(c *gin.Context, message string, data any, page, limit int, total int64) {
	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}
	c.JSON(200, paginatedEnvelope{
		Message: message,
		Data:    data,
		Meta: pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

func Created(c *gin.Context, message string, data any) {
	c.JSON(201, envelope{Message: message, Data: data})
}

func Error(c *gin.Context, status int, message string) {
	c.JSON(status, envelope{Message: message, Data: nil})
}
